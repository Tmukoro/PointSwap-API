package models

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	UserID                uuid.UUID  `json:"user_id" db:"user_id"`
	NotificationType      string     `json:"notification_type" db:"notification_type"`
	Title                 string     `json:"title" db:"title"`
	Body                  string     `json:"body" db:"body"`
	RelatedProductID      *uuid.UUID `json:"related_product_id,omitempty" db:"related_product_id"`
	RelatedUserID         *uuid.UUID `json:"related_user_id,omitempty" db:"related_user_id"`
	RelatedConversationID *uuid.UUID `json:"related_conversation_id,omitempty" db:"related_conversation_id"`
	IsRead                bool       `json:"is_read" db:"is_read"`
	IsPushed              bool       `json:"is_pushed" db:"is_pushed"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	ReadAt                *time.Time `json:"read_at,omitempty" db:"read_at"`
}

// NotificationWithUserDetails includes user info for display
type NotificationWithUserDetails struct {
	Notification
	OtherUserName   *string `json:"other_user_name,omitempty"`
	OtherUserAvatar *string `json:"other_user_avatar,omitempty"`
	ProductTitle    *string `json:"product_title,omitempty"`
}
