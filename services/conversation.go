package services

import (
	"postswapapi/config"
	"postswapapi/models"
	"time"

	"github.com/google/uuid"
)

//This is to get the two users who are going to be talking with each other

func GetConversationParticipants(conversationID, excludeUserID uuid.UUID) ([]models.Users, error) {
	rows, err := config.DB.Query(`
       SELECT user_id, username, display_name, email, phone_number, avatar_url FROM users u
	   JOIN conversation_participants cp ON u.user_id = cp.user_id
	   WHERE cp.conversation_id = $1 AND cp.left_at IS NULL AND u.user_id != $2	
	`, conversationID, excludeUserID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var participants []models.Users

	for rows.Next() {
		var participant models.Users
		err := rows.Scan(&participant.User_ID, &participant.Username, &participant.Display_name,
			&participant.Email, &participant.Phone_number, &participant.Avatar_url,
		)

		if err != nil {
			return nil, err
		}

		participants = append(participants, participant)
	}

	return participants, nil
}

// This is to fetch details about the product when hitting the message about product
func GetProductForConversation(productID uuid.UUID) (*models.ProductWithSeller, error) {
	var product models.ProductWithSeller

	err := config.DB.QueryRow(`
	   SELECT p.product_id, p.title, p.category, p.estimated_size, p.status, p.created_at, p.updated_at
	   u.user_id, u.username, u.display_name, u.email, u.phone_number, u.avatar_url
	   FROM products p JOIN users u ON p.seller_id = u.user_id
	   WHERE p.prodcut_id = $1
	`, productID).Scan(&product.Product_ID, &product.Title, &product.Category, &product.Estimated_size, &product.Status,
		&product.Created_at, &product.Updated_at, &product.Seller.User_ID, &product.Seller.Username,
		&product.Seller.Display_name, &product.Seller.Email, &product.Seller.Phone_number, &product.Seller.Avatar_url)

	if err != nil {
		return nil, err
	}

	return &product, nil
}

//to show the last message in the preview box of the chat in the chat list of a user

func GetLastMessage(conversationID uuid.UUID) (*models.Messages, error) {
	var message models.Messages

	err := config.DB.QueryRow(`
	  SELECT message_id, conversation_id, sender_id, message_type, content, file_url, reply_to_message_id,
	  created_at, updated_at, edited_at, is_deleted FROM messages WHERE conversation_id = $1 AND is_deleted = false
	  ORDER BY created_at DESC
	  LIMIT 1
	`, conversationID).Scan(&message.Message_ID, &message.Conversation_ID, &message.Sender_ID, &message.Message_type,
		&message.Content, &message.File_url, &message.Reply_to_message_ID, &message.Created_at, &message.Updated_at,
		&message.Edited_at, &message.Is_deleted)

	if err != nil {
		return nil, err
	}

	return &message, nil
}

func GetUnreadMessageCount(conversationID, userID uuid.UUID) (int, error) {
	var count int

	var lastReadMessageID *uuid.UUID

	err := config.DB.QueryRow(`
	   SELECT last_read_message_id FROM conversation_participants 
	   WHERE conversation_id = $1 AND user_id = $2
	`, conversationID, userID).Scan(&lastReadMessageID)

	if err != nil {
		return 0, err
	}

	if lastReadMessageID == nil {
		err = config.DB.QueryRow(`
	      SELECT COUNT(*) FROM messages 
		  WHERE conversation_id = $1 AND sender_id = $2 AND is_deleted = false	
		`, conversationID, userID).Scan(&count)
	}

	var lastReadTime time.Time
	err = config.DB.QueryRow(`
	   SELECT created_at FROM messages WHERE message_id = $1
	`, &lastReadMessageID).Scan(&lastReadTime)

	if err != nil {
		return 0, err
	}

	err = config.DB.QueryRow(`
	  SELECT COUNT(*) FROM messages
	  WHERE conversation_id = $1 AND sender_id = $2 AND is_deleted = false AND created_at > $3
	`, conversationID, userID, lastReadTime).Scan(&count)

	if err != nil {
		return 0, nil
	}

	return count, nil
}

func GetConversationWithDetails(conversationID, currentUserID uuid.UUID) (*models.ConversationWithDetails, error) {
	var conv models.ConversationWithDetails

	err := config.DB.QueryRow(`
	   SELECT conversation_id, conversation_type, product_id, created_at, updated_at, last_message_at
	   FROM conversations WHERE conversation_id = $1
	`, conversationID).Scan(&conv.Conversation_ID, &conv.Conversation_type, &conv.Product.Product_ID, &conv.Created_at,
		&conv.Updated_at, &conv.Last_message_at,
	)

	if err != nil {
		return nil, err
	}

	//Get participants

	participants, err := GetConversationParticipants(conversationID, currentUserID)

	if err != nil {
		return nil, err
	}

	conv.Participants = participants

	//Get product details if it exists(basically checks if the product still exists before initiating the chat)

	if conv.Product.Product_ID != uuid.Nil {
		product, err := GetProductForConversation(conv.Product.Product_ID)
		if err != nil {
			conv.Product = product
		}
	}

	//Get last message

	unreadCount, err := GetUnreadMessageCount(conversationID, currentUserID)

	if err != nil {
		conv.Unread_count = unreadCount
	}

	return &conv, nil

}
