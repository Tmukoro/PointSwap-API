package models

import (
	"time"

	"github.com/google/uuid"
)

type Products struct {
	Product_ID     uuid.UUID `json:"product_id" db:"product_id"`
	Seller_ID      uuid.UUID `json:"seller_id" db:"seller_id"`
	Title          string    `json:"title" db:"title"`
	Category       string    `json:"category" db:"category"`
	Estimated_size *string   `json:"estimated_size" db:"estimated_size"`
	Status         string    `json:"status" db:"status"`
	Created_at     time.Time `json:"created_at" db:"created_at"`
	Updated_at     time.Time `json:"updated_at" db:"updated_at"`
}

type ProductPhotos struct {
	Photo_ID      uuid.UUID `json:"photo_id" db:"photo_id"`
	Product_ID    uuid.UUID `json:"product_id" db:"product_id"`
	Image_Url     string    `json:"image_url" db:"image_url"`
	Display_order int       `json:"display_order" db:"display_order"`
	Created_at    time.Time `json:"created_at" db:"created_at"`
}

//Model for when you click on a product and want to preview it in full detail

type ProductWithSeller struct {
	Product_ID     uuid.UUID       `json:"product_id"`
	Seller         Users           `json:"sellers"`
	Title          string          `json:"title"`
	Category       string          `json:"category"`
	Estimated_size string          `json:"estimated_size"`
	Status         string          `json:"status"`
	Photos         []ProductPhotos `json:"photos"`
	Created_at     time.Time       `json:"created_at"`
	Updated_at     time.Time       `json:"updated_at"`
}

// Models required for creating a product upload request
type CreateProductRequest struct {
	Category       string   `json:"category" binding:"required"`
	Image_Urls     []string `json:"image_urls"  binding:"required,min=1"`
	Title          string   ` json:"title" binding:"required"`
	Estimated_size string   ` json:"estimated_size" binding:"required"`
}

// (Inactive)

//To preview your product before post

// type ProductPreviewRequest struct {
// 	Image_Url      string    `json:"image_url"`
// 	Title          string    `json:"title"`
// 	Estimated_size string    `json:"estimated_size"`
// 	Seller_ID      uuid.UUID `json:"seller_id"`
// }
