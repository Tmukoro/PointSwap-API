package models

import (
	"time"

	"github.com/google/uuid"
)

type Messages struct {
	Message_ID          uuid.UUID  `json:"message_id" db:"message_id"`
	Conversation_ID     uuid.UUID  `json:"conversation_id" db:"conversation_id"`
	Sender_ID           uuid.UUID  `json:"sender_id" db:"sender_id"`
	Message_type        string     `json:"message_type" db:"message_type"`
	Content             string     `json:"content" db:"content"`
	File_url            string     `json:"file_url" db:"file_url"`
	Reply_to_message_ID uuid.UUID  `json:"reply_to_message_id" db:"reply_to_message_id"`
	Created_at          time.Time  `json:"created_at" db:"created_at"`
	Updated_at          time.Time  `json:"updated_at" db:"updated_at"`
	Edited_at           *time.Time `json:"edited_at" db:"updated_at"`
	Is_deleted          bool       `json:"is_deleted" db:"is_deleted"`
}

type SendMessageRequest struct {
	Conversation_ID     uuid.UUID  `json:"conversation_id"`
	Message_type        string     `json:"message_type"`
	Content             *string    `json:"content"`
	File_url            *string    `json:"file_url"`
	Reply_to_message_ID *uuid.UUID `json:"reply_to_message_id"`
}

type MesageWithSender struct {
	Message_ID       uuid.UUID  `json:"message_id"`
	Conversation_ID  uuid.UUID  `json:"conversation_id"`
	Sender           Users      `json:"sender"`
	Message_type     string     `json:"message_type"`
	Content          *string    `json:"content"`
	File_url         *string    `json:"file_url"`
	Reply_to_message *Messages  `json:"reply_to_message,omitempty"`
	Created_at       time.Time  `json:"created_at"`
	Updated_at       time.Time  `json:"updated_at"`
	Edited_at        *time.Time `json:"edited_at"`
	Is_deleted       bool       `json:"is_deleted"`
}
