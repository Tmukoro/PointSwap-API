package handlers

import (
	"database/sql"
	"net/http"
	"postswapapi/config"
	"postswapapi/models"
	"postswapapi/services"
	"postswapapi/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//Function to create the conversation

func CreateConversation(ctx *gin.Context) {
	var req models.CreateConversationRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	presentUser, exists := ctx.Get("User")

	if !exists {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User Not Authenticated")
		return
	}

	user, ok := presentUser.(models.Users)

	if !ok {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "User Not Valid")
		return
	}

	//Prevents users from messaging themselves

	if req.User_ID == user.User_ID {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "User cannot message themself")
		return
	}

	//Checks if the user you're messaging exists

	var otherUser models.Users

	err := config.DB.QueryRow(`
     SELECT user_id, username, display_name, email, phone_number, avatar_url
	 FROM users WHERE user_id = $1
   `, req.User_ID).Scan(&otherUser.User_ID, &otherUser.Username, &otherUser.Display_name, &otherUser.Email, &otherUser.Phone_number,
		&otherUser.Avatar_url)

	if err == sql.ErrNoRows {
		utils.ErrorResponse(ctx, http.StatusNotFound, "User not found")
		return
	}

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Database Error")
		return
	}

	//Check if product is still available during feed view

	var product *models.ProductWithSeller

	if req.Product_ID != nil {
		var p models.ProductWithSeller

		err := config.DB.QueryRow(`
		  SELECT p.product_id, p.title, p.category, p.estimated_size, p.status, p.created_at, p.updated_at,
		  u.user_id, u.username, u.display_name, u.email, u.phone_number, u.avatar_url FROM products p JOIN 
		  users u ON p.seller_id = u.user_id
		  WHERE product_id = $1
		`, *req.Product_ID).Scan(&p.Product_ID, &p.Title, &p.Category, &p.Estimated_size, &p.Status, &p.Created_at,
			&p.Updated_at, &p.Seller.User_ID, &p.Seller.Username, &p.Seller.Display_name, &p.Seller.Email,
			&p.Seller.Phone_number, &p.Seller.Avatar_url)

		if err == sql.ErrNoRows {
			utils.ErrorResponse(ctx, http.StatusNotFound, "Product not availabe")
			return
		}

		if err != nil {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, "Database error")
			return
		}

		product = &p

	}

	//Check if a conversation is already ongoing for that product

	var existingConvoID uuid.UUID
	var checkQuery string

	if req.Product_ID != nil {

		checkQuery = `
	  SELECT c.conversation_id FROM conversations c 
	  JOIN conversation_participants cp1 ON c.conversation_id = cp1.conversation_id 
	  JOIN conversation_participants cp2 ON c.conversation_id = cp2.conversation_id 
	  WHERE c.product_id = $1 AND cp1.user_id = $2
	  AND cp2.user_id = $3 AND cp1.left_at IS NULL AND cp2.left_at is NULL
	`
		err = config.DB.QueryRow(checkQuery, *req.Product_ID, user.User_ID, req.User_ID).Scan(&existingConvoID)

		if err != sql.ErrNoRows {

			//if the conversation does exist then return it
			conversations, err := services.GetConversationWithDetails(existingConvoID, user.User_ID)

			if err != nil {
				utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch existing conversation")
				return
			}

			utils.SuccessResponse(ctx, http.StatusOK, "Conversation Already Exists", conversations)
			return

		}
	}

	checkQuery = `
    SELECT c.conversation_id FROM conversations c 
	JOIN conversation_participants cp1 ON c.conversation_id = cp1.conversation_id 
	JOIN conversation_participants cp2 ON c.conversation_id = cp2.conversation_id 
	WHERE c.product_id IS NULL AND cp1.user_id = $1 AND cp2.user_id = $2 AND cp1.left_at IS NULL AND
	cp2.left_at IS NULL
  `
	err = config.DB.QueryRow(checkQuery, user.User_ID, req.User_ID).Scan(&existingConvoID)

	if err != sql.ErrNoRows {

		if err != nil {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, "Database error")
			return
		}

	}

	//Creating the new conversation

	tx, err := config.DB.Begin()
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to start transaction")
		return
	}

	defer tx.Rollback()

	conversation := models.Conversations{
		Conversation_ID:   uuid.New(),
		Conversation_type: "direct",
		Product_ID:        product.Product_ID,
		Created_by:        user.User_ID,
		Created_at:        time.Now(),
		Updated_at:        time.Now(),
		Last_message_at:   time.Now(),
	}

	_, err = config.DB.Exec(`
     INSERT INTO conversations (conversation_id, conversation_type, product_id, created_by, created_at, updated_at,
	 last_message_at) VALUES ($1, $2, $3, $4, $5, $6, $7)
  `, conversation.Conversation_ID, conversation.Conversation_type, conversation.Product_ID, conversation.Created_by,
		conversation.Created_at, conversation.Updated_at, conversation.Last_message_at)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create conversation")
		return
	}

	//Add participants to the conversation

	participants := []uuid.UUID{user.User_ID, req.User_ID}

	for _, participantID := range participants {
		_, err = tx.Exec(`
	     INSERT INTO conversation_participants (participant_id, conversation_id, user_id, joined_at, last_read_message_id)
		 VALUES ($1, $2, $3, $4, $5)	
		`, uuid.New(), conversation.Conversation_ID, participantID, time.Now(), participantID)

		if err != nil {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
			return
		}
	}

	//commit the transcation

	if err = tx.Commit(); err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to commit transcation")
		return
	}

	//prepare response

	response := models.ConversationWithDetails{
		Conversation_ID:   conversation.Conversation_ID,
		Conversation_type: conversation.Conversation_type,
		Product:           product,
		Participants:      []models.Users{user, otherUser},
		Last_message:      nil,
		Unread_count:      0,
		Created_at:        conversation.Created_at,
		Updated_at:        conversation.Updated_at,
		Last_message_at:   conversation.Last_message_at,
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Conversation Successfully Created", response)
}

//Basically your chat feed showing all users you're currently chating with

func GetMyConversations(ctx *gin.Context) {

	presentUser, exists := ctx.Get("User")

	if !exists {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, ok := presentUser.(models.Users)

	if !ok {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Invalid User")
		return
	}

	//parse pagination (basically scroll function and all)

	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)

	if err != nil || offset < 0 {
		offset = 0
	}

	//Get the conversations with details

	query := `
	SELECT c.conversation_id, c.conversation_type, c.product_id, c.created_at, c.updated_at, c.last_message_at,
		   m.message_id, m.content, m.message_type, m.created_at as message_created_at,
		   sender.user_id as sender_id, sender.username as sender_username, sender.display_name as sender_display_name
	FROM conversations c
	JOIN conversation_participants cp ON c.conversation_id = cp.conversation_id
	LEFT JOIN messages m ON c.conversation_id = m.conversation_id 
		AND m.created_at = (
			SELECT MAX(created_at) FROM messages 
			WHERE conversation_id = c.conversation_id AND is_deleted = false
		)
	LEFT JOIN users sender ON m.sender_id = sender.user_id
	WHERE cp.user_id = $1 AND cp.left_at IS NULL
	ORDER BY c.last_message_at DESC
	LIMIT $2 OFFSET $3
`

	rows, err := config.DB.Query(query, user.User_ID, limit+1, offset)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch conversations")
		return
	}

	defer rows.Close()

	var conversationIDs []uuid.UUID

	conversationMap := make(map[uuid.UUID]*models.ConversationWithDetails)

	for rows.Next() {
		var convID uuid.UUID
		var convType string
		var productID *uuid.UUID
		var createdAt, updatedAt, lastMessageAt time.Time
		var messageID *uuid.UUID
		var messageContent *string
		var messageType *string
		var messageCreatedAt *time.Time
		var senderID *uuid.UUID
		var senderUsername, senderDisplayName *string

		err = rows.Scan(&convID, &convType, &productID, &createdAt, &updatedAt, &lastMessageAt, &messageID,
			&messageContent, &messageType, &messageCreatedAt, &senderID, &senderUsername, &senderDisplayName)

		if err != nil {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to parse Conversations")
			return
		}

		//check if conversation has already been processed

		if _, exists := conversationMap[convID]; !exists {
			conversationMap[convID] = &models.ConversationWithDetails{
				Conversation_ID:   convID,
				Conversation_type: convType,
				Created_at:        createdAt,
				Updated_at:        updatedAt,
				Last_message_at:   lastMessageAt,
			}

			conversationIDs = append(conversationIDs, convID)
		}

		if messageID != nil {
			conversationMap[convID].Last_message = &models.Messages{
				Message_ID:      *messageID,
				Conversation_ID: convID,
				Sender_ID:       *senderID,
				Message_type:    *messageType,
				Content:         *messageContent,
				Created_at:      *messageCreatedAt,
			}
		}

	}

	//Check pagination

	hasMore := len(conversationIDs) > limit

	if hasMore {
		conversationIDs = conversationIDs[:limit]
	}

	//Get additional details for each convo (such as unread notification badge)

	var conversations []models.ConversationWithDetails

	for _, convID := range conversationIDs {
		conv := conversationMap[convID]

		//Get Participants

		participants, err := services.GetConversationParticipants(convID, user.User_ID)

		if err != nil {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch participants")
			return
		}

		conv.Participants = participants

		//Get unread count

		unreadCount, err := services.GetUnreadMessageCount(convID, user.User_ID)

		if err != nil {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch unread count")
			return
		}

		conv.Unread_count = unreadCount

		conversations = append(conversations, *conv)
	}

	//calculating the next offset

	var nextOffset *int

	if hasMore {
		next := offset + limit
		nextOffset = &next
	}

	response := models.InfiniteScrollData{
		Items: conversations,
		Meta: models.PaginationMeta{
			Limit:       limit,
			Offset:      offset,
			Has_more:    hasMore,
			Next_offset: nextOffset,
		},
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Chat Successfully fetched", response)

}

//Get the conversation when you click on the chat

func GetConversationByID(ctx *gin.Context) {
	conversationIDStr := ctx.Param("conversation_id")
	conversationID, err := uuid.Parse(conversationIDStr)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	presentUser, exists := ctx.Get("User")

	if !exists {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, ok := presentUser.(models.Users)

	if !ok {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Invalid User")
		return
	}

	//Check if the user is present in this conversation

	var participantCount int

	err = config.DB.QueryRow(`
	   SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = $1 AND user_id = $2 AND
	   left_at IS NULL
	`, conversationID, user.User_ID).Scan(&participantCount)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Database Error")
		return
	}

	if participantCount == 0 {
		utils.ErrorResponse(ctx, http.StatusForbidden, "Access Denied")
		return
	}

	//Get the conversation details

	conversation, err := services.GetConversationWithDetails(conversationID, user.User_ID)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch conversation details")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Conversation Successfully Fetched", conversation)

	//Used to auto mark as read once clicked on

	// _, err = config.DB.Exec(`
	//     UPDATE conversation_participants SET last_read_message_id = (
	// 	  SELECT message_id FROM messages WHERE conversation_id = $1 AND
	// 	  is_deleted = false ORDER BY created_at DESC LIMIT 1
	// 	)
	// 	  WHERE conversation_id = $1 AND user_id = $2
	// `, conversationID, user.User_ID)

}
