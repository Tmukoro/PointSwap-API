package models

import (
	"time"

	"github.com/google/uuid"
)

type PushToken struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	Token      string    `json:"token" db:"token"`
	DeviceType string    `json:"device_type" db:"device_type"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type RegisterPushTokenRequest struct {
	Token      string `json:"token" binding:"required"`
	DeviceType string `json:"device_type" binding:"required"`
}
