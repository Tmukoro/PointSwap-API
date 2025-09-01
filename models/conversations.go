package models

import (
	"time"

	"github.com/google/uuid"
)

type Conversations struct {
	Conversation_ID   uuid.UUID `json:"conversation_id" db:"conversation_id"`
	Conversation_type string    `json:"conversation_type" db:"conversation_type"`
	Created_by        uuid.UUID `json:"created_by" db:"created_by"`
	Created_at        time.Time `json:"created_at" db:"created_at"`
	Updated_at        time.Time `json:"updated_at" db:"updated_at"`
	Last_message_at   time.Time `json:"last_message_at" db:"last_message_at"`
	Product_ID        uuid.UUID `json:"product_id" db:"product_id"`
}

type ConversationParticipants struct {
	Participant_ID       uuid.UUID `json:"participant_id" db:"participant_id"`
	Conversation_ID      uuid.UUID `json:"conversation_id" db:"conversation_id"`
	User_ID              uuid.UUID `json:"user_id" db:"user_ud"`
	Joined_at            time.Time `json:"joined_at" db:"joined_at"`
	Left_at              time.Time `json:"left_at" db:"left_at"`
	Last_read_message_ID uuid.UUID `json:"last_read_message_id" db:"last_read_message_id"`
	Is_muted             bool      `json:"is_muted" db:"is_muted"`
}

type CreateConversationRequest struct {
	Product_ID *uuid.UUID `json:"product_id"`
	User_ID    uuid.UUID  `json:"user_id" binding:"required"`
}

// This is basically the chat page to show the list of convos you're having
type ConversationWithDetails struct {
	Conversation_ID   uuid.UUID          `json:"conversation_id"`
	Conversation_type string             `json:"conversation_type"`
	Product           *ProductWithSeller `json:"product,omitempty"`
	Participants      []Users            `json:"participants"`
	Last_message      *Messages          `json:"last_message,omitempty"`
	Unread_count      int                `json:"unread_count"`
	Created_at        time.Time          `json:"created_at"`
	Updated_at        time.Time          `json:"updated_at"`
	Last_message_at   time.Time          `json:"last_message_at"`
}
