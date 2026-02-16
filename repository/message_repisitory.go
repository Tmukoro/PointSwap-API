package repository

import (
	"database/sql"
	"fmt"
	"postswapapi/models"
	"time"

	"github.com/google/uuid"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// GetOrCreateConversation finds existing conversation or creates new one between two users
func (r *MessageRepository) GetOrCreateConversation(user1ID, user2ID uuid.UUID) (uuid.UUID, error) {
	var conversationID uuid.UUID

	query := `SELECT get_or_create_conversation($1, $2)`
	err := r.db.QueryRow(query, user1ID, user2ID).Scan(&conversationID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get or create conversation: %w", err)
	}

	return conversationID, nil
}

// CreateMessage saves a new message to the database
func (r *MessageRepository) CreateMessage(conversationID, senderID uuid.UUID, messageText string) (*models.Message, error) {
	message := &models.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       senderID,
		MessageText:    messageText,
		IsRead:         false,
		CreatedAt:      time.Now(),
	}

	query := `
        INSERT INTO messages (id, conversation_id, sender_id, message_text, is_read, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, conversation_id, sender_id, message_text, is_read, created_at
    `

	err := r.db.QueryRow(query,
		message.ID,
		message.ConversationID,
		message.SenderID,
		message.MessageText,
		message.IsRead,
		message.CreatedAt,
	).Scan(
		&message.ID,
		&message.ConversationID,
		&message.SenderID,
		&message.MessageText,
		&message.IsRead,
		&message.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return message, nil
}

// GetConversationMessages retrieves messages for a conversation with pagination
func (r *MessageRepository) GetConversationMessages(conversationID uuid.UUID, limit, offset int) ([]models.MessageWithSender, error) {
	messages := []models.MessageWithSender{}

	query := `
        SELECT 
            m.id,
            m.conversation_id,
            m.sender_id,
            CONCAT(u.first_name, ' ', u.last_name) as sender_name,
            u.avatar_url as sender_avatar,
            m.message_text,
            m.is_read,
            m.created_at,
            m.deleted_at
        FROM messages m
        INNER JOIN users u ON m.sender_id = u.user_id
        WHERE m.conversation_id = $1 
          AND m.deleted_at IS NULL
        ORDER BY m.created_at DESC
        LIMIT $2 OFFSET $3
    `

	rows, err := r.db.Query(query, conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var msg models.MessageWithSender
		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msg.SenderName,
			&msg.SenderAvatar,
			&msg.MessageText,
			&msg.IsRead,
			&msg.CreatedAt,
			&msg.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}

// GetUserConversations retrieves all conversations for a user with details
func (r *MessageRepository) GetUserConversations(userID uuid.UUID) ([]models.ConversationWithDetails, error) {
	conversations := []models.ConversationWithDetails{}

	query := `
        SELECT 
            c.id,
            c.created_at,
            c.updated_at,
            c.last_message_at,
            other_user.user_id as other_user_id,
            CONCAT(other_user.first_name, ' ', other_user.last_name) as other_user_name,
            other_user.avatar_url as other_user_avatar,
            last_msg.message_text as last_message_text,
            COALESCE(
                (SELECT COUNT(*) 
                 FROM messages m2 
                 WHERE m2.conversation_id = c.id 
                   AND m2.sender_id != $1
                   AND m2.created_at > cp.last_read_at
                   AND m2.deleted_at IS NULL
                ), 0
            ) as unread_count
        FROM conversations c
        INNER JOIN conversation_participants cp ON c.id = cp.conversation_id
        INNER JOIN conversation_participants other_cp ON c.id = other_cp.conversation_id 
            AND other_cp.user_id != $1
        INNER JOIN users other_user ON other_cp.user_id = other_user.user_id
        LEFT JOIN LATERAL (
            SELECT message_text 
            FROM messages 
            WHERE conversation_id = c.id 
              AND deleted_at IS NULL
            ORDER BY created_at DESC 
            LIMIT 1
        ) last_msg ON true
        WHERE cp.user_id = $1
        ORDER BY c.updated_at DESC
    `

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user conversations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var conv models.ConversationWithDetails
		err := rows.Scan(
			&conv.ID,
			&conv.CreatedAt,
			&conv.UpdatedAt,
			&conv.LastMessageAt,
			&conv.OtherUserID,
			&conv.OtherUserName,
			&conv.OtherUserAvatar,
			&conv.LastMessageText,
			&conv.UnreadCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, conv)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating conversations: %w", err)
	}

	return conversations, nil
}

// MarkConversationAsRead updates the last_read_at timestamp for a user in a conversation
func (r *MessageRepository) MarkConversationAsRead(conversationID, userID uuid.UUID) error {
	query := `
        UPDATE conversation_participants
        SET last_read_at = NOW()
        WHERE conversation_id = $1 AND user_id = $2
    `

	result, err := r.db.Exec(query, conversationID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark conversation as read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("conversation participant not found")
	}

	return nil
}

// GetConversationByID retrieves a conversation by ID
func (r *MessageRepository) GetConversationByID(conversationID uuid.UUID) (*models.Conversation, error) {
	conversation := &models.Conversation{}

	query := `
        SELECT id, created_at, updated_at, last_message_at
        FROM conversations
        WHERE id = $1
    `

	err := r.db.QueryRow(query, conversationID).Scan(
		&conversation.ID,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
		&conversation.LastMessageAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("conversation not found")
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return conversation, nil
}

// VerifyUserInConversation checks if a user is a participant in a conversation
func (r *MessageRepository) VerifyUserInConversation(conversationID, userID uuid.UUID) (bool, error) {
	var exists bool

	query := `
        SELECT EXISTS(
            SELECT 1 
            FROM conversation_participants 
            WHERE conversation_id = $1 AND user_id = $2
        )
    `

	err := r.db.QueryRow(query, conversationID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to verify user in conversation: %w", err)
	}

	return exists, nil
}

// GetOtherParticipantID gets the other user's ID in a 1-on-1 conversation
func (r *MessageRepository) GetOtherParticipantID(conversationID, currentUserID uuid.UUID) (uuid.UUID, error) {
	var otherUserID uuid.UUID

	query := `
        SELECT user_id
        FROM conversation_participants
        WHERE conversation_id = $1 AND user_id != $2
        LIMIT 1
    `

	err := r.db.QueryRow(query, conversationID, currentUserID).Scan(&otherUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, fmt.Errorf("other participant not found")
		}
		return uuid.Nil, fmt.Errorf("failed to get other participant: %w", err)
	}

	return otherUserID, nil
}

// DeleteMessage soft deletes a message (sets deleted_at)
func (r *MessageRepository) DeleteMessage(messageID, userID uuid.UUID) error {
	query := `
        UPDATE messages
        SET deleted_at = NOW()
        WHERE id = $1 AND sender_id = $2 AND deleted_at IS NULL
    `

	result, err := r.db.Exec(query, messageID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("message not found or already deleted")
	}

	return nil
}
