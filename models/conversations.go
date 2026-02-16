package models

import (
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID            uuid.UUID `json:"id" db:"id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	LastMessageAt time.Time `json:"last_message_at" db:"last_message_at"`
}

// links users to conversations
type ConversationParticipant struct {
	ID             uuid.UUID `json:"id" db:"id"`
	ConversationID uuid.UUID `json:"conversation_id" db:"conversation_id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	JoinedAt       time.Time `json:"joined_at" db:"joined_at"`
	LastReadAt     time.Time `json:"last_read_at" db:"last_read_at"`
}

// Basically message in a convo
type Message struct {
	ID             uuid.UUID `json:"id" db:"id"`
	ConversationID uuid.UUID `json:"conversation_id" db:"conversation_id"`
	SenderID       uuid.UUID `json:"sender_id" db:"sender_id"`
	MessageText    string    `json:"message_text" db:"message_text"`
	IsRead         bool      `json:"is_read" db:"is_read"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	DeletedAt      time.Time `json:"deleted_at" db:"deleted_at"`
}

// The particpants info with their last message display
type ConversationWithDetails struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	LastMessageAt   *time.Time `json:"last_message_at,omitempty" db:"last_message_at"`
	OtherUserID     uuid.UUID  `json:"other_user_id" db:"other_user_id"`
	OtherUserName   string     `json:"other_user_name" db:"other_user_name"`
	OtherUserAvatar *string    `json:"other_user_avatar,omitempty" db:"other_user_avatar"`
	LastMessageText *string    `json:"last_message_text,omitempty" db:"last_message_text"`
	UnreadCount     int        `json:"unread_count" db:"unread_count"`
}

// Includes senders info
type MessageWithSender struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	ConversationID uuid.UUID  `json:"conversation_id" db:"conversation_id"`
	SenderID       uuid.UUID  `json:"sender_id" db:"sender_id"`
	SenderName     string     `json:"sender_name" db:"sender_name"`
	SenderAvatar   *string    `json:"sender_avatar,omitempty" db:"sender_avatar"`
	MessageText    string     `json:"message_text" db:"message_text"`
	IsRead         bool       `json:"is_read" db:"is_read"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Creating a message request for Ably endpoint
type CreateMessageRequest struct {
	RecipientID uuid.UUID `json:"recipient_id" binding:"required"`
	MessageText string    `json:"message_text" binding:"required,min=1,max=5000"`
}

// SendMessageRequest for sending to existing conversation
type SendMessageRequest struct {
	MessageText string `json:"message_text" binding:"required,min=1,max=5000"`
}
