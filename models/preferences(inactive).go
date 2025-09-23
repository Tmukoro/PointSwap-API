package models

import (
	"time"

	"github.com/google/uuid"
)

//this model is for when the user posts and supplies what they need to be swapped

type ProductWants struct {
	WantID         uuid.UUID `json:"want_id" db:"want_id"`
	ProductID      uuid.UUID `json:"product_id" db:"product_id"`
	WantUserID     uuid.UUID `json:"want_user_id" db:"want_user_id"`
	WantedCategory string    `json:"wanted_category" db:"wanted_category"`
	WantedSize     *string   `json:"wanted_size" db:"wanted_size"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

//This is what the users wants in general

type UserPreferences struct {
	PreferenceID uuid.UUID `json:"preference_id" db:"preference_id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	Category     string    `json:"category" db:"category"`
	Size         *string   `json:"size" db:"size"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

//tracks matches found and shows to users

type PotentialMatches struct {
	MatchID        uuid.UUID  `json:"match_id" db:"match_id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	TheirProductID uuid.UUID  `json:"their_product_id" db:"their_product_id"`
	MyProductID    *uuid.UUID `json:"my_product_id" db:"my_product"`
	MatchType      string     `json:"match_type" db:"match_type"`
	IsDismissed    bool       `json:"is_dismissed" db:"is_dismissed"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

//API request to create and update the product you want

type CreateProductWantRequest struct {
	WantedCategory string  `json:"wanted_category" binding:"required"`
	WantedSize     *string `json:"wanted_size"`
}

type UpdateProductWantRequest struct {
	WantedCategory string  `json:"wanted_category" binding:"required"`
	WantedSize     *string `json:"wanted_size"`
}

//Wishlist section for you to be notified when someone uploads what you want

type CreateUserPreferenceRequest struct {
	Category string  `json:"category" binding:"required"`
	Size     *string `json:"size"`
}

type UpdateUserPreferenceRequest struct {
	Category string `json:"category" binding:"required"`
	Size     string `json:"size"`
	IsActive bool   `json:"is_active"`
}

//response models

type ProductWantWithProduct struct {
	WantID         uuid.UUID         `json:"want_id"`
	Product        ProductWithSeller `json:"product"`
	WantedCategory string            `json:"wanted_category"`
	WantedSize     *string           `json:"wanted_size"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type PotentialMatchWithDetails struct {
	MatchID      uuid.UUID          `json:"match_id"`
	UserID       uuid.UUID          `json:"user_id"`
	TheirProduct ProductWithSeller  `json:"their_product"`
	MyProduct    *ProductWithSeller `json:"my_product"`
	MatchType    string             `json:"match_type"`
	MatchReason  string             `json:"match_reason"`
	IsDismissed  bool               `json:"is_dismissed"`
	CreatedAt    time.Time          `json:"created_at"`
}

type MatchSuggestion struct {
	SuggestionID  string             `json:"suggestion_id"`
	TheirProduct  ProductWithSeller  `json:"their_product"`
	MyProduct     *ProductWithSeller `json:"my_product"`
	MatchReason   string             `json:"match_reason"`
	Compatibility float64            `json:"compatibility"`
	CanInitiate   bool               `json:"can_initiate"`
}

//Dismiss multiple matches at once

type BulkDismissRequest struct {
	MatchIDS []uuid.UUID `json:"match_ids" binding:"required, min=1"`
}
