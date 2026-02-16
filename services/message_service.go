package services

import (
	"context"
	"fmt"
	"os"
	"postswapapi/models"
	"postswapapi/repository"

	"github.com/ably/ably-go/ably"
	"github.com/google/uuid"
)

type MessageService struct {
	repo       *repository.MessageRepository
	ablyClient *ably.Realtime
}

func NewMessageService(repo *repository.MessageRepository) (*MessageService, error) {
	ablyAPIKey := os.Getenv("ABLY_KEY")
	if ablyAPIKey == "" {
		return nil, fmt.Errorf("ABLY_KEY environment variable not set")
	}

	// Initialize Ably client
	client, err := ably.NewRealtime(
		ably.WithKey(ablyAPIKey),
		ably.WithEchoMessages(false), // Don't echo messages back to sender
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ably client: %w", err)
	}

	return &MessageService{
		repo:       repo,
		ablyClient: client,
	}, nil
}

// Close cleanly closes the Ably connection
func (s *MessageService) Close() {
	s.ablyClient.Close()
}

// SendMessage creates a new message or starts a conversation
func (s *MessageService) SendMessage(senderID, recipientID uuid.UUID, messageText string) (*models.Message, error) {
	// Get or create conversation
	conversationID, err := s.repo.GetOrCreateConversation(senderID, recipientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create conversation: %w", err)
	}

	// Save message to database
	message, err := s.repo.CreateMessage(conversationID, senderID, messageText)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Publish to Ably for real-time delivery
	if err := s.publishMessageToAbly(conversationID, message); err != nil {
		// Log error but don't fail the request - message is already saved
		fmt.Printf("Warning: failed to publish message to Ably: %v\n", err)
	}

	return message, nil
}

// SendMessageToConversation sends a message to an existing conversation
func (s *MessageService) SendMessageToConversation(conversationID, senderID uuid.UUID, messageText string) (*models.Message, error) {
	// Verify sender is in conversation
	isParticipant, err := s.repo.VerifyUserInConversation(conversationID, senderID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify participant: %w", err)
	}
	if !isParticipant {
		return nil, fmt.Errorf("user is not a participant in this conversation")
	}

	// Save message to database
	message, err := s.repo.CreateMessage(conversationID, senderID, messageText)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Publish to Ably for real-time delivery
	if err := s.publishMessageToAbly(conversationID, message); err != nil {
		fmt.Printf("Warning: failed to publish message to Ably: %v\n", err)
	}

	return message, nil
}

// publishMessageToAbly publishes a message to the conversation's Ably channel
func (s *MessageService) publishMessageToAbly(conversationID uuid.UUID, message *models.Message) error {
	// Get the Ably channel for this conversation
	channelName := fmt.Sprintf("conversation:%s", conversationID.String())
	channel := s.ablyClient.Channels.Get(channelName)

	// Create the payload to send
	payload := map[string]interface{}{
		"id":              message.ID.String(),
		"conversation_id": message.ConversationID.String(),
		"sender_id":       message.SenderID.String(),
		"message_text":    message.MessageText,
		"created_at":      message.CreatedAt,
	}

	// Publish to the channel
	err := channel.Publish(context.Background(), "new_message", payload)
	if err != nil {
		return fmt.Errorf("failed to publish to ably: %w", err)
	}

	return nil
}

// GetConversationMessages retrieves messages with pagination
func (s *MessageService) GetConversationMessages(conversationID, userID uuid.UUID, limit, offset int) ([]models.MessageWithSender, error) {
	// Verify user is in conversation
	isParticipant, err := s.repo.VerifyUserInConversation(conversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify participant: %w", err)
	}
	if !isParticipant {
		return nil, fmt.Errorf("user is not a participant in this conversation")
	}

	messages, err := s.repo.GetConversationMessages(conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, nil
}

// GetUserConversations retrieves all conversations for a user
func (s *MessageService) GetUserConversations(userID uuid.UUID) ([]models.ConversationWithDetails, error) {
	conversations, err := s.repo.GetUserConversations(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	return conversations, nil
}

// MarkConversationAsRead marks all messages in a conversation as read
func (s *MessageService) MarkConversationAsRead(conversationID, userID uuid.UUID) error {
	// Verify user is in conversation
	isParticipant, err := s.repo.VerifyUserInConversation(conversationID, userID)
	if err != nil {
		return fmt.Errorf("failed to verify participant: %w", err)
	}
	if !isParticipant {
		return fmt.Errorf("user is not a participant in this conversation")
	}

	err = s.repo.MarkConversationAsRead(conversationID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark as read: %w", err)
	}

	return nil
}

// DeleteMessage soft deletes a message
func (s *MessageService) DeleteMessage(messageID, userID uuid.UUID) error {
	err := s.repo.DeleteMessage(messageID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	// Optionally: publish deletion event to Ably
	// so other clients can update their UI in real-time

	return nil
}

// GenerateAblyTokenForUser creates a token for client-side Ably authentication
// This is important for security - clients shouldn't have your API key
func (s *MessageService) GenerateAblyTokenForUser(userID uuid.UUID) (string, error) {
	ablyAPIKey := os.Getenv("ABLY_KEY")
	if ablyAPIKey == "" {
		return "", fmt.Errorf("ABLY_KEY not configured")
	}

	// Create a REST client for token generation
	restClient, err := ably.NewREST(ably.WithKey(ablyAPIKey))
	if err != nil {
		return "", fmt.Errorf("failed to create REST client: %w", err)
	}

	// Request a token with user's ID as ClientID
	token, err := restClient.Auth.RequestToken(context.Background(), &ably.TokenParams{
		ClientID: userID.String(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}

	// Return just the token string
	return token.Token, nil
}

// GetConversationChannelName returns the Ably channel name for a conversation
// Useful for frontend to know which channel to subscribe to
func (s *MessageService) GetConversationChannelName(conversationID uuid.UUID) string {
	return fmt.Sprintf("conversation:%s", conversationID.String())
}
