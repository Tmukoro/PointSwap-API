package services

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type MatchingService struct {
	db *sql.DB
}

type ProductMatch struct {
	MatchedProductID    uuid.UUID
	MatchedUserID       uuid.UUID
	MatchedUserName     string
	YourProductID       uuid.UUID
	YourProductTitle    string
	MatchedProductTitle string
	MatchedSize         string
	YourSize            string
	Category            string
}

func NewMatchingService(db *sql.DB) *MatchingService {
	return &MatchingService{db: db}
}

// FindMatches finds complementary matches when a product is uploaded
func (s *MatchingService) FindMatches(productID, sellerID uuid.UUID) ([]ProductMatch, error) {
	var matches []ProductMatch

	// Get the uploaded product details
	var uploadedProduct struct {
		Title    string
		Category string
		Size     string
	}

	query := `SELECT title, category, estimated_size FROM products WHERE product_id = $1`
	err := s.db.QueryRow(query, productID).Scan(&uploadedProduct.Title, &uploadedProduct.Category, &uploadedProduct.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Get seller's wants (what the seller is looking for)
	sellerWantsQuery := `
		SELECT wanted_category, wanted_size, product_id
		FROM product_wants
		WHERE want_user_id = $1
	`

	sellerWantsRows, err := s.db.Query(sellerWantsQuery, sellerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get seller wants: %w", err)
	}
	defer sellerWantsRows.Close()

	var sellerWants []struct {
		WantedCategory string
		WantedSize     string
		ProductID      uuid.UUID
	}

	for sellerWantsRows.Next() {
		var want struct {
			WantedCategory string
			WantedSize     string
			ProductID      uuid.UUID
		}
		if err := sellerWantsRows.Scan(&want.WantedCategory, &want.WantedSize, &want.ProductID); err != nil {
			continue
		}
		sellerWants = append(sellerWants, want)
	}

	// For each seller want, find matching products
	for _, sellerWant := range sellerWants {
		// Find products that match what seller wants
		// AND check if those product owners want what seller is offering
		matchQuery := `
			SELECT DISTINCT
				p.product_id,
				p.seller_id,
				p.title,
				p.estimated_size,
				CONCAT(u.first_name, ' ', u.last_name) as owner_name
			FROM products p
			INNER JOIN product_wants pw ON pw.want_user_id = p.seller_id
			INNER JOIN users u ON u.user_id = p.seller_id
			WHERE p.category = $1
			  AND p.estimated_size = $2
			  AND p.seller_id != $3
			  AND p.status = 'active'
			  AND pw.wanted_category = $4
			  AND pw.wanted_size = $5
		`

		matchRows, err := s.db.Query(
			matchQuery,
			sellerWant.WantedCategory, // Products matching seller's want category
			sellerWant.WantedSize,     // Products matching seller's want size
			sellerID,                  // Exclude seller's own products
			uploadedProduct.Category,  // Other users want this category
			uploadedProduct.Size,      // Other users want this size
		)

		if err != nil {
			continue
		}

		for matchRows.Next() {
			var match ProductMatch
			err := matchRows.Scan(
				&match.MatchedProductID,
				&match.MatchedUserID,
				&match.MatchedProductTitle,
				&match.MatchedSize,
				&match.MatchedUserName,
			)
			if err != nil {
				continue
			}

			match.YourProductID = productID
			match.YourProductTitle = uploadedProduct.Title
			match.YourSize = uploadedProduct.Size
			match.Category = uploadedProduct.Category

			matches = append(matches, match)
		}
		matchRows.Close()
	}

	return matches, nil
}

// CreateMatchNotifications creates notifications for both users in a match
func (s *MatchingService) CreateMatchNotifications(match ProductMatch, uploaderID uuid.UUID) error {
	// Notification for uploader
	uploaderNotification := `
		INSERT INTO notifications (
			id, user_id, notification_type, title, body, 
			related_product_id, related_user_id, is_read, is_pushed, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, false, false, NOW())
	`

	uploaderTitle := "Perfect Match Found! 🎯"
	uploaderBody := fmt.Sprintf("%s has a %s (size %s) and wants your %s (size %s)!",
		match.MatchedUserName,
		match.MatchedProductTitle,
		match.MatchedSize,
		match.YourProductTitle,
		match.YourSize,
	)

	_, err := s.db.Exec(
		uploaderNotification,
		uuid.New(),
		uploaderID,
		"product_match",
		uploaderTitle,
		uploaderBody,
		match.MatchedProductID, // Link to the other user's product
		match.MatchedUserID,
	)

	if err != nil {
		return fmt.Errorf("failed to create uploader notification: %w", err)
	}

	// Notification for matched user
	matchedUserNotification := `
		INSERT INTO notifications (
			id, user_id, notification_type, title, body, 
			related_product_id, related_user_id, is_read, is_pushed, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, false, false, NOW())
	`

	matchedUserTitle := "Perfect Match Found! 🎯"
	matchedUserBody := fmt.Sprintf("Someone has a %s (size %s) and wants your %s (size %s)!",
		match.YourProductTitle,
		match.YourSize,
		match.MatchedProductTitle,
		match.MatchedSize,
	)

	_, err = s.db.Exec(
		matchedUserNotification,
		uuid.New(),
		match.MatchedUserID,
		"product_match",
		matchedUserTitle,
		matchedUserBody,
		match.YourProductID, // Link to uploader's product
		uploaderID,
	)

	if err != nil {
		return fmt.Errorf("failed to create matched user notification: %w", err)
	}

	return nil
}
