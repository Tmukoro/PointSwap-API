package models

import (
	"time"

	"github.com/google/uuid"
)

type Notifications struct {
	Notification_ID         uuid.UUID  `json:"notification_id" db:"notification_id"`
	User_ID                 uuid.UUID  `json:"user_id" db:"user_id"`
	Notification_type       string     `json:"notification_type" db:"notification_type"`
	Title                   string     `json:"title" db:"title"`
	Message                 string     `json:"message" db:"message"`
	Related_conversation_ID *uuid.UUID `json:"related_conversation_id"`
	Related_product_ID      *uuid.UUID `json:"related_product_id" db:"related_product_id"`
	Related_user_ID         *uuid.UUID `json:"related_user_id" db:"related_user_id"`
	Is_Read                 bool       `json:"is_read" db:"is_read"`
	Is_Pushed               bool       `json:"is_pushed" db:"is_pushed"`
	Created_at              time.Time  `json:"created_at" db:"created_at"`
	Read_at                 time.Time  `json:"read_at" db:"read_at"`
}
