package handlers

import (
	"fmt"
	"net/http"
	"postswapapi/models"
	"postswapapi/services"
	"postswapapi/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MessageHandler struct {
	service *services.MessageService
}

func NewMessageHandler(service *services.MessageService) *MessageHandler {
	return &MessageHandler{
		service: service,
	}
}

// SendMessage creates a new conversation and sends first message
// POST /api/messages
func (h *MessageHandler) SendMessage(c *gin.Context) {
	// Get authenticated user ID from context (set by your auth middleware)
	senderID, err := getUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Prevent sending message to yourself
	if senderID == req.RecipientID {
		utils.ErrorResponse(c, http.StatusBadRequest, "cannot send message to yourself")
		return
	}

	message, err := h.service.SendMessage(senderID, req.RecipientID, req.MessageText)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to send message")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "message sent successfully", gin.H{
		"message":         message,
		"conversation_id": message.ConversationID,
	})
}

// SendMessageToConversation sends message to existing conversation
// POST /api/conversations/:conversation_id/messages
func (h *MessageHandler) SendMessageToConversation(c *gin.Context) {
	senderID, err := getUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	conversationID, err := uuid.Parse(c.Param("conversation_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid conversation ID")
		return
	}

	var req models.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	message, err := h.service.SendMessageToConversation(conversationID, senderID, req.MessageText)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "message sent successfully", gin.H{
		"message": message,
	})
}

// GetConversationMessages retrieves messages for a conversation
// GET /api/conversations/:conversation_id/messages?limit=50&offset=0
func (h *MessageHandler) GetConversationMessages(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	conversationID, err := uuid.Parse(c.Param("conversation_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid conversation ID")
		return
	}

	// Parse pagination params
	limit := 50 // default
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	messages, err := h.service.GetConversationMessages(conversationID, userID, limit, offset)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "messages retrieved successfully", gin.H{
		"messages": messages,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetUserConversations retrieves all conversations for the authenticated user
// GET /api/conversations
func (h *MessageHandler) GetUserConversations(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	conversations, err := h.service.GetUserConversations(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to get conversations")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "conversations retrieved successfully", gin.H{
		"conversations": conversations,
	})
}

// MarkConversationAsRead marks all messages in a conversation as read
// PUT /api/conversations/:conversation_id/read
func (h *MessageHandler) MarkConversationAsRead(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	conversationID, err := uuid.Parse(c.Param("conversation_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid conversation ID")
		return
	}

	err = h.service.MarkConversationAsRead(conversationID, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "conversation marked as read", nil)
}

// DeleteMessage soft deletes a message
// DELETE /api/messages/:message_id
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	messageID, err := uuid.Parse(c.Param("message_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid message ID")
		return
	}

	err = h.service.DeleteMessage(messageID, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "message deleted successfully", nil)
}

// GetAblyToken generates an Ably token for the authenticated user
// GET /api/messages/ably-token
func (h *MessageHandler) GetAblyToken(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	token, err := h.service.GenerateAblyTokenForUser(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to generate token")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "token generated successfully", gin.H{
		"token": token,
	})
}

// Helper function to get user ID from Gin context
func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	presentUser, exists := c.Get("User")
	if !exists {
		return uuid.Nil, fmt.Errorf("user not authenticated")
	}

	user, ok := presentUser.(models.Users)
	if !ok {
		return uuid.Nil, fmt.Errorf("user not found")
	}

	return user.User_ID, nil
}
