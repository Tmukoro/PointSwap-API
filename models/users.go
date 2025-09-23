package models

import (
	"time"

	"github.com/google/uuid"
)

type Users struct {
	User_ID       uuid.UUID  `json:"user_id" db:"user_id"`
	Username      string     `json:"username" db:"username"`
	Email         string     `json:"email" db:"email"`
	Password_Hash string     `json:"password_hash" db:"password_hash"`
	Display_name  string     `json:"display_name" db:"display_name"`
	Phone_number  string     `json:"phone_number" db:"phone_number"`
	Avatar_url    *string    `json:"avatar_url" db:"avatar_url"`
	FCM_token     *string    `json:"fcm_token" db:"fcm_token"`
	Created_at    time.Time  `json:"created_at" db:"created_at"`
	Updated_at    time.Time  `json:"updated_at" db:"updated_at"`
	Last_seen     *time.Time `json:"last_seen" db:"last_seen"`
	Is_online     bool       `json:"is_online" db:"is_online"`
}

type UserRegistrationRequest struct {
	Username     string `binding:"required"`
	Display_name string `binding:"required"`
	Phone_number string `binding:"required,min=10,max=20"`
	Email        string `binding:"required,email"`
	Password     string `binding:"required,min=10"`
}

type UserLoginRequest struct {
	Email    string ` json:"email" binding:"required,email"`
	Password string ` json:"password" binding:"required,min=10"`
}

type UserBlocks struct {
	Block_ID   uuid.UUID `json:"block_id" db:"block_id"`
	Blokcer_ID uuid.UUID `json:"blocker_id" db:"blocker_id"`
	Blocked_ID uuid.UUID `json:"blocked_id" db:"blocked_id"`
	Created_at time.Time `json:"created_at" db:"created_at"`
}
