package services

import (
	"database/sql"
	"fmt"
	"postswapapi/models"
	"time"

	"github.com/google/uuid"
	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
	// "postswapapi/models"
)

type NotificationService struct {
	db         *sql.DB
	expoClient *expo.PushClient
}

func NewNotificationService(db *sql.DB) *NotificationService {
	return &NotificationService{
		db:         db,
		expoClient: expo.NewPushClient(nil),
	}
}

// GetUserPushTokens retrieves all push tokens for a user
func (s *NotificationService) GetUserPushTokens(userID uuid.UUID) ([]string, error) {
	query := `SELECT token FROM push_tokens WHERE user_id = $1`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			continue
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

// SendMessageNotification sends a push notification for a new message
func (s *NotificationService) SendMessageNotification(recipientID uuid.UUID, senderName, messageText string, conversationID uuid.UUID) error {
	// Get recipient's push tokens
	tokens, err := s.GetUserPushTokens(recipientID)
	if err != nil {
		return fmt.Errorf("failed to get push tokens: %w", err)
	}

	if len(tokens) == 0 {
		fmt.Println("No push tokens found for user:", recipientID)
		return nil // Not an error, user just doesn't have notifications enabled
	}

	// Prepare notification content
	title := senderName
	body := messageText

	// Truncate long messages
	if len(body) > 100 {
		body = body[:97] + "..."
	}

	// Handle empty messages (image/audio only)
	if body == "" {
		body = "Sent you a message"
	}

	// Create push messages
	var pushTokens []expo.ExponentPushToken
	for _, token := range tokens {
		pushToken, err := expo.NewExponentPushToken(token)
		if err != nil {
			fmt.Printf("Invalid push token %s: %v\n", token, err)
			continue
		}
		pushTokens = append(pushTokens, pushToken)
	}

	if len(pushTokens) == 0 {
		return fmt.Errorf("no valid push tokens")
	}

	// Build notification
	messages := make([]expo.PushMessage, len(pushTokens))
	for i, pushToken := range pushTokens {
		messages[i] = expo.PushMessage{
			To:       []expo.ExponentPushToken{pushToken},
			Body:     body,
			Title:    title,
			Sound:    "default",
			Priority: expo.DefaultPriority,
			Data: map[string]string{
				"type":            "new_message",
				"conversation_id": conversationID.String(),
				"sender_id":       "", // You can add sender ID if needed
			},
		}
	}

	// Send notifications
	response, err := s.expoClient.Publish(&messages[0])
	if err != nil {
		return fmt.Errorf("failed to send push notification: %w", err)
	}

	// Check for errors in response
	if response.ValidateResponse() != nil {
		fmt.Printf("Push notification errors: %v\n", response.PushMessage)
	}

	fmt.Printf("Successfully sent push notification to %d devices\n", len(pushTokens))
	return nil
}

// CreateProductMatchNotification creates a notification for product match
func (s *NotificationService) CreateProductMatchNotification(userID, otherUserID, productID uuid.UUID, otherUserName, productTitle string) error {
	_, err := s.db.Exec(`
		INSERT INTO notifications (
			notification_id, user_id, notification_type, title, message, 
			related_product_id, related_user_id, is_read, is_pushed, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`,
		uuid.New(),
		userID,
		"product_match",
		"Perfect Match Found! 🎉",
		fmt.Sprintf("You and %s might just have what each other needs!", otherUserName),
		productID,
		otherUserID,
		false,
		false,
		time.Now(),
	)

	if err != nil {
		fmt.Printf("ERROR inserting notification: %v\n", err)
	}

	return err
}

// GetUserNotifications retrieves all notifications for a user with details
func (s *NotificationService) GetUserNotifications(userID uuid.UUID) ([]models.NotificationWithUserDetails, error) {
	query := `
		SELECT 
			n.notification_id, 
			n.user_id, 
			n.notification_type, 
			n.title, 
			n.message,
			n.related_product_id, 
			n.related_user_id, 
			n.related_conversation_id,
			n.is_read, 
			n.is_pushed, 
			n.created_at, 
			n.read_at,
			COALESCE(CONCAT(u.first_name, ' ', u.last_name), '') as other_user_name,
			u.avatar_url as other_user_avatar,
			p.title as product_title
		FROM notifications n
		LEFT JOIN users u ON n.related_user_id = u.user_id
		LEFT JOIN products p ON n.related_product_id = p.product_id
		WHERE n.user_id = $1
		ORDER BY n.created_at DESC
		LIMIT 100
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	var notifications []models.NotificationWithUserDetails
	for rows.Next() {
		var n models.NotificationWithUserDetails
		err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.NotificationType,
			&n.Title,
			&n.Body,
			&n.RelatedProductID,
			&n.RelatedUserID,
			&n.RelatedConversationID,
			&n.IsRead,
			&n.IsPushed,
			&n.CreatedAt,
			&n.ReadAt,
			&n.OtherUserName,
			&n.OtherUserAvatar,
			&n.ProductTitle,
		)
		if err != nil {
			fmt.Printf("Scan error: %v\n", err)
			continue
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(notificationID string, userID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE notifications
		SET is_read = true, read_at = $1
		WHERE notification_id = $2 AND user_id = $3
	`

	_, err := s.db.Exec(query, now, notificationID, userID)
	return err
}

// GetUnreadCount gets count of unread notifications
func (s *NotificationService) GetUnreadCount(userID uuid.UUID) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1 AND is_read = false
	`

	err := s.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// FindAndNotifyProductMatches finds matching products when a new product is created
func (s *NotificationService) FindAndNotifyProductMatches(newProductID uuid.UUID) error {
	// Get the newly created product details
	var newProduct struct {
		ProductID uuid.UUID
		SellerID  uuid.UUID
		Category  string
		Size      string
		Title     string
	}

	err := s.db.QueryRow(`
		SELECT product_id, seller_id, title, category, estimated_size
		FROM products
		WHERE product_id = $1
	`, newProductID).Scan(&newProduct.ProductID, &newProduct.SellerID, &newProduct.Title, &newProduct.Category, &newProduct.Size)

	if err != nil {
		return fmt.Errorf("failed to get new product: %w", err)
	}

	// Get what the new product owner wants
	var newProductWants struct {
		WantedCategory string
		WantedSize     string
	}

	err = s.db.QueryRow(`
		SELECT wanted_category, wanted_size
		FROM product_wants
		WHERE product_id = $1
	`, newProductID).Scan(&newProductWants.WantedCategory, &newProductWants.WantedSize)

	if err == sql.ErrNoRows {
		// No wants set yet, skip matching
		fmt.Println("No wants set for product:", newProductID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get product wants: %w", err)
	}

	// Find matching products
	query := `
		SELECT 
			p.product_id,
			p.seller_id,
			p.title,
			u.first_name,
			u.last_name
		FROM products p
		INNER JOIN product_wants pw ON p.product_id = pw.product_id
		INNER JOIN users u ON p.seller_id = u.user_id
		WHERE p.seller_id != $1
		  AND p.status = 'active'
		  AND p.category = $2
		  AND p.estimated_size = $3
		  AND pw.wanted_category = $4
		  AND pw.wanted_size = $5
	`

	rows, err := s.db.Query(query,
		newProduct.SellerID,
		newProductWants.WantedCategory,
		newProductWants.WantedSize,
		newProduct.Category,
		newProduct.Size,
	)

	if err != nil {
		return fmt.Errorf("failed to find matches: %w", err)
	}
	defer rows.Close()

	matchCount := 0

	for rows.Next() {
		var match struct {
			ProductID uuid.UUID
			SellerID  uuid.UUID
			Title     string
			FirstName string
			LastName  string
		}

		err := rows.Scan(&match.ProductID, &match.SellerID, &match.Title, &match.FirstName, &match.LastName)
		if err != nil {
			continue
		}

		matchCount++
		matchedUserName := fmt.Sprintf("%s %s", match.FirstName, match.LastName)

		// Get new product owner's name
		var newOwnerFirstName, newOwnerLastName string
		s.db.QueryRow(`SELECT first_name, last_name FROM users WHERE user_id = $1`, newProduct.SellerID).Scan(&newOwnerFirstName, &newOwnerLastName)
		newOwnerName := fmt.Sprintf("%s %s", newOwnerFirstName, newOwnerLastName)

		// Create notification for the new product owner
		s.CreateProductMatchNotification(
			newProduct.SellerID,
			match.SellerID,
			match.ProductID,
			matchedUserName,
			match.Title,
		)

		// Create notification for the matched product owner
		s.CreateProductMatchNotification(
			match.SellerID,
			newProduct.SellerID,
			newProduct.ProductID,
			newOwnerName,
			newProduct.Title,
		)
	}

	fmt.Printf("Found %d matches for product %s\n", matchCount, newProductID)
	return nil
}
