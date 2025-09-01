package handlers

import (
	"database/sql"
	"net/http"
	"postswapapi/config"
	"postswapapi/models"
	"postswapapi/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//function to send a message

func SendMessage(ctx *gin.Context) {
	var req models.SendMessageRequest

	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
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

	//validate message type

	if req.Message_type != "text" && req.Message_type != "image" && req.Message_type != "file" {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid message type")
		return
	}

	//validate content based on message type(basically ensuring if youre sending any type something must be present)

	if req.Message_type == "test" && (req.Content == nil || *req.Content == "") {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "A text must be present to be sent")
		return
	}

	if (req.Message_type == "image" || req.Message_type == "file") && (req.Content == nil || *req.Content == "") {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "image/file must have a url")
		return
	}

	//check if user is a participant in this conversation

	var participantCount int

	err := config.DB.QueryRow(`
	   SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = $1 AND user_id = $2
	   AND left_at IS NULL 
	`, req.Conversation_ID, user.User_ID).Scan(&participantCount)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Database Error")
		return
	}

	if participantCount == 0 {
		utils.ErrorResponse(ctx, http.StatusForbidden, "Access denied to this conversation")
		return
	}

	//validate reply to message if provided

	if req.Reply_to_message_ID != &uuid.Nil {
		var replyMessageExist int

		err := config.DB.QueryRow(`
		   SELECT COUNT(*) FROM messages WHERE message_id = $1 AND conversation_id = $2 AND is_deleted = false
		`, *req.Reply_to_message_ID, req.Conversation_ID).Scan(&replyMessageExist)

		if err != nil {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, "Database Error")
			return
		}

		if replyMessageExist == 0 {
			utils.ErrorResponse(ctx, http.StatusBadRequest, "Reply to message not found")
			return
		}
	}

	//Start transaction

	tx, err := config.DB.Begin()

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to start transcation")
		return
	}

	tx.Rollback()

	//Create the message

	message := models.Messages{
		Message_ID:          uuid.New(),
		Conversation_ID:     req.Conversation_ID,
		Sender_ID:           user.User_ID,
		Message_type:        req.Message_type,
		Content:             "",
		File_url:            "",
		Reply_to_message_ID: uuid.Nil,
		Created_at:          time.Now(),
		Updated_at:          time.Now(),
		Edited_at:           nil,
		Is_deleted:          false,
	}

	if req.Content != nil {
		message.Content = *req.Content
	}

	if req.File_url != nil {
		message.File_url = *req.File_url
	}

	if req.Reply_to_message_ID != &uuid.Nil {
		message.Reply_to_message_ID = *req.Reply_to_message_ID
	}

	//insert the message to the db

	_, err = tx.Exec(`
	    INSERT INTO messages (message_id, conversation_id, sender_id, message_type, content,
		reply_to_message_id, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, message.Message_ID, message.Conversation_ID, message.Sender_ID, message.Message_type,
		message.Content, message.Reply_to_message_ID, message.Created_at, message.Updated_at,
		message.Is_deleted)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to send message")
		return
	}

	//Update conversations last message

	_, err = tx.Exec(`
	    UPDATE conversations SET last_message_at = $1, updated_at = $1
		WHERE conversation_id = $2
	`, time.Now(), req.Conversation_ID)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to updated conversation")
		return
	}

	//commit transaction

	if err = tx.Commit(); err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	//create a response with sender details

	response := models.MesageWithSender{
		Message_ID:      message.Message_ID,
		Conversation_ID: message.Conversation_ID,
		Sender:          user,
		Message_type:    message.Message_type,
		Content:         &message.Content,
		File_url:        &message.File_url,
		Created_at:      message.Created_at,
		Updated_at:      message.Updated_at,
		Edited_at:       message.Edited_at,
		Is_deleted:      message.Is_deleted,
	}

	//Get reply message details if its a reply

	if req.Reply_to_message_ID != &uuid.Nil {
		var replyMessage models.Messages
		err := config.DB.QueryRow(`
		   SELECT message_id, conversation_id, sender_id, message_type, content, file_url, reply_to message_id,
		   created_at, updated_at, edited_at, is_deleted FROM messages WHERE message_id = $1
		`, *req.Reply_to_message_ID).Scan(&message.Message_ID, &message.Conversation_ID, &message.Sender_ID,
			&message.Message_type, &message.Content, &message.File_url, &message.Reply_to_message_ID,
			&message.Created_at, &message.Updated_at, &message.Edited_at, &message.Is_deleted)

		if err == nil {
			response.Reply_to_message = &replyMessage
		}
	}

	utils.SuccessResponse(ctx, http.StatusCreated, "Message sent successfully", response)

}

//Get messages inside the chat pagination

func GetMessages(ctx *gin.Context) {
	conversationIDStr := ctx.Param("conversation_id")
	conversationID, err := uuid.Parse(conversationIDStr)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid Conversation ID")
		return
	}

	presentUser, exists := ctx.Get("User")

	if !exists {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not Authenticated")
		return
	}

	user, ok := presentUser.(models.Users)

	if !ok {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Invalid User")
		return
	}

	//Check if user is a participant in this conversation

	var participantCount int

	err = config.DB.QueryRow(`
	   SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = $1 AND user_id = $2 AND left_at IS NULL
	`, conversationID, user.User_ID).Scan(&participantCount)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Database error")
		return
	}

	if participantCount == 0 {
		utils.ErrorResponse(ctx, http.StatusForbidden, "Access denied to this conversation")
		return
	}

	//parse pagination parameters

	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)

	if err != nil || offset < 0 {
		offset = 0
	}

	// Get messages with sender detais

	query := `
	   SELECT m.message_id, m.conversation_id, m.message_type, m.content, m.file_url, m.reply_to_message_id,
	   m.created_at, m.updated_at, m.edited_at, m.is_deleted, u.user_id, u.username, u.display_name, u.email,
	   u.phone_number, u.avatat_url FROM messages m JOIN users u ON m.sender_id = u.user_id
	   WHERE m.conversation_id = $1 and m.is_deleted = false ORDER BY m.created_at DESC
	   LIMIT $2 OFFSET $3
	`

	rows, err := config.DB.Query(query, conversationID, limit+1, offset)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch messages")
		return
	}

	defer rows.Close()

	var messages []models.MesageWithSender

	for rows.Next() {
		var message models.MesageWithSender
		var replyToMessageID sql.NullString

		err := rows.Scan(&message.Message_ID, &message.Conversation_ID, &message.Message_type, &message.Content,
			&message.File_url, &replyToMessageID, &message.Created_at, &message.Updated_at, &message.Edited_at,
			&message.Edited_at, &message.Is_deleted, &message.Sender.User_ID, &message.Sender.Username,
			&message.Sender.Display_name, &message.Sender.Email, &message.Sender.Phone_number,
			&message.Sender.Avatar_url)

		if err != nil {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to parse messages")
			return
		}

		//handle reply messages if present

		if replyToMessageID.Valid {

			replyId, err := uuid.Parse(replyToMessageID.String)

			if err == nil {
				var replyMessage models.Messages

				err := config.DB.QueryRow(`
				       SELECT message_id, conversation_id, sender_id, message_type, content, file_url,
					   reply_to_message_id, created_at, updated_at, edited_at, is_deleted FROM messages
					   WHERE message_id = $1
				`, replyId).Scan(&replyMessage.Message_ID, &replyMessage.Conversation_ID, &replyMessage.Sender_ID,
					&replyMessage.Message_type, &replyMessage.Content, &replyMessage.File_url,
					&replyMessage.Reply_to_message_ID, &replyMessage.Created_at, &replyMessage.Updated_at,
					&replyMessage.Edited_at, &replyMessage.Is_deleted)

				if err == nil {
					message.Reply_to_message = &replyMessage
				}
			}

		}

		messages = append(messages, message)

	}

	//check pagination

	hasMore := len(messages) > limit

	if hasMore {
		messages = messages[:limit]
	}

	//calculate next offset

	var nexOffset *int

	if hasMore {
		next := offset + limit
		nexOffset = &next
	}

	response := models.InfiniteScrollData{
		Items: messages,
		Meta: models.PaginationMeta{
			Limit:       limit,
			Offset:      offset,
			Has_more:    hasMore,
			Next_offset: nexOffset,
		},
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Messages Successfully Retrieved", response)

}

//function to edit the message

func EditMessage(ctx *gin.Context) {
	messageIDStr := ctx.Param("message_id")

	messageID, err := uuid.Parse(messageIDStr)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid Message ID")
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err = ctx.ShouldBindBodyWithJSON(&req); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	presentUser, exists := ctx.Get("User")

	if !exists {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, ok := presentUser.(models.Users)

	if !ok {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Invalud User")
		return
	}

	//check if message exists and belongs to the user

	var senderID uuid.UUID

	var messageType string

	err = config.DB.QueryRow(`
	   SELECT sender_id, message_type FRROM messages
	   WHERE message_id = $1 AND is_deleted = false
	`, messageID).Scan(&senderID, &messageType)

	if err == sql.ErrNoRows {
		utils.ErrorResponse(ctx, http.StatusNotFound, "Message not found")
		return
	}

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Database error")
		return
	}

	if senderID != user.User_ID {
		utils.ErrorResponse(ctx, http.StatusForbidden, "You can only edit your own messages")
		return
	}

	if messageType != "text" {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Can only edit text")
		return
	}

	//updating the message

	now := time.Now()

	_, err = config.DB.Exec(`
	   UPDATE messages
	   SET content = $1, updated_at = $2, edited_at = $3
	   WHERE message_id = $4
	`, req.Content, now, now, messageID)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to edit message")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Message successfully edited", gin.H{
		"message_id": messageID,
		"content":    req.Content,
		"edited_at":  now,
	})
}

//Delete a message (temporary-soft delete)

func DeleteMessage(ctx *gin.Context) {
	messageIDStr := ctx.Param("message_id")
	messageID, err := uuid.Parse(messageIDStr)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid message id")
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

	//check if the message exists and belongs to the uer

	var senderID uuid.UUID

	err = config.DB.QueryRow(`
	     SELECT sender_id from messages WHERE message_id = $1 AND is_deleted = false
	`, messageID).Scan(&senderID)

	if err == sql.ErrNoRows {
		utils.ErrorResponse(ctx, http.StatusNotFound, "Message not found")
		return
	}

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Database Error")
		return
	}

	if senderID != user.User_ID {
		utils.ErrorResponse(ctx, http.StatusForbidden, "Only user can delete message")
		return
	}

	//soft delete the message

	_, err = config.DB.Exec(`
	   UPDATE messages SET is_deleted = true, updated_at = $1
	   WHERE message_id = $2
	`, time.Now(), messageID)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete message")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Message deleted successfully", nil)

}
