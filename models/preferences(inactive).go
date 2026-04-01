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

//API request to create and update the product you want

type CreateProductWantRequest struct {
	WantedCategory string  `json:"wanted_category"`
	WantedSize     *string `json:"wanted_size" binding:"required"`
}

type UpdateProductWantRequest struct {
	WantedCategory string  `json:"wanted_category"`
	WantedSize     *string `json:"wanted_size" binding:"required"`
}

type ProductWantWithProduct struct {
	WantID         uuid.UUID         `json:"want_id"`
	Product        ProductWithSeller `json:"product"`
	WantedCategory string            `json:"wanted_category"`
	WantedSize     *string           `json:"wanted_size"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}
