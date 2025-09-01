package models

import (
	"time"

	"github.com/google/uuid"
)

type Categories struct {
	Category_ID   uuid.UUID `json:"category_id" db:"category_id"`
	Name          string    `json:"name" db:"name"`
	Display_name  string    `json:"display_name" db:"display_name"`
	Display_order int       `json:"display_order" db:"display_order"`
	Is_active     bool      `json:"is_active" db:"is_active"`
	Created_at    time.Time `json:"created_at" db:"created_at"`
}

type SizeOptions struct {
	Size_ID       uuid.UUID `json:"size_id" db:"size_id"`
	Category_ID   uuid.UUID `json:"category_id" db:"category_id"`
	Size_value    string    `json:"size_value" db:"size_value"`
	Display_order int       `json:"display_order" db:"display_order"`
	Created_at    time.Time `json:"created_at" db:"created_at"`
}
