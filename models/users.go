package models

import (
	"time"

	"github.com/google/uuid"
)

type Users struct {
	User_ID       uuid.UUID  `json:"user_id" db:"user_id"`
	Email         string     `json:"email" db:"email"`
	Password_Hash string     `json:"password_hash" db:"password_hash"`
	First_Name    string     `json:"first_name" db:"first_name"`
	Last_Name     string     `json:"last_name" db:"last_name"`
	Avatar_url    *string    `json:"avatar_url" db:"avatar_url"`
	Location      string     `json:"location" db:"location"`
	FCM_token     *string    `json:"fcm_token" db:"fcm_token"`
	Created_at    time.Time  `json:"created_at" db:"created_at"`
	Updated_at    time.Time  `json:"updated_at" db:"updated_at"`
	Last_seen     *time.Time `json:"last_seen" db:"last_seen"`
	Is_online     bool       `json:"is_online" db:"is_online"`
}

type UserRegistrationRequest struct {
	Email    string `binding:"required,email"`
	Password string `binding:"required,min=10"`
}

type UserProfileSetUpRequest struct {
	First_Name string `binding:"required"`
	Last_Name  string `binding:"required"`
	Avatar_url string `binding:"required"`
}
type UserLoginRequest struct {
	Email    string ` json:"email" binding:"required,email"`
	Password string ` json:"password" binding:"required,min=10"`
}

type UserLocationRequest struct {
	Location string `json:"location" binding:"required"`
}
type UserBlocks struct {
	Block_ID   uuid.UUID `json:"block_id" db:"block_id"`
	Blokcer_ID uuid.UUID `json:"blocker_id" db:"blocker_id"`
	Blocked_ID uuid.UUID `json:"blocked_id" db:"blocked_id"`
	Created_at time.Time `json:"created_at" db:"created_at"`
}
