package handlers

import (
	"fmt"
	"net/http"
	"postswapapi/config"
	"postswapapi/models"
	"postswapapi/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//used to create new notifications for an action

func CreateNotification(userID uuid.UUID, notificationType, title, message string, relatedProductID, relatedUserID *uuid.UUID) {

	notificationID := uuid.New()

	config.DB.Exec(`
	   INSERT INTO notifications (notification_id, user_id, notification_type, title, message, related_product_id,
	   related_user_id, is_read, is_pushed, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, notificationID, userID, notificationType, title, message, relatedProductID, relatedUserID, false, false, time.Now())
}

/* This is to trigger notification for mutual matches where user a has what user b wants and user b has what user a wants
   So basically it searches the db and matches the records sending a notification
*/

func TriggerNotificationsForNewWant(productID, requestingUserID uuid.UUID, wantedCategory string, wantedSize *string) {

	//Get the requesting users product info

	var myCategory, myProductTitle, myUsername, mySize string

	err := config.DB.QueryRow(`
	   SELECT p.category, p.estimated_size, p.title, u.username FROM products p
	   JOIN users u on p.seller_id = u.user_id
	   WHERE p.product_id = $1
	`, productID).Scan(&myCategory, &mySize, &myProductTitle, &myUsername)

	if err != nil {
		return //this wont be able to proceed without product info
	}

	//query for the basic description of this function

	var query string

	var args []any

	if wantedSize != nil && *wantedSize != "" {
		query = `
	     SELECT p.product_id, p.seller_id, p.title, u.username, pw.wanted_category FROM products p
		 JOIN users u ON p.seller_id = u.user_id
		 JOIN product_wants pw ON p.product_id = pw.product_id
		 WHERE p.category = $1 AND p.estimated_size = $2 AND pw.wanted_category = $3 AND 
		 (pw.wanted_size = $4 OR pw.wanted_size IS NULL) AND p.status = 'acitve' AND p.seller_id != $5
		 LIMIT 5	
		`
		args = []any{wantedCategory, *wantedSize, myCategory, mySize, requestingUserID}
	}

	rows, err := config.DB.Query(query, args...)

	if err != nil {
		return
	}

	defer rows.Close()

	//creating mutual interest notification

	for rows.Next() {
		var theirProductID, ownerID uuid.UUID

		var theirProductTitle, ownerUsername, theirWantedCategory string

		err := rows.Scan(&theirProductID, &ownerID, &theirProductTitle, &ownerUsername, &theirWantedCategory)

		if err != nil {
			continue
		}

		//notify the other user about mutual interest

		title := "Mutual Swap Interest"

		message := fmt.Sprintf("Hey %s has %s and wants %s - you also have %s and want %s, you guys have a perfect match!",
			myUsername, myProductTitle, wantedCategory, theirProductTitle, theirWantedCategory)

		CreateNotification(ownerID, "Mutual Match", title, message, &theirProductID, &requestingUserID)

		//notify the current user about mutual interest

		myTitle := "Mutual Swap Interest"

		myMessage := fmt.Sprintf("Hey %s has %s and wants %s - you also have %s and want %s, you guys have a perfect match!",
			ownerID, theirProductTitle, theirWantedCategory, myProductTitle, wantedCategory)

		CreateNotification(requestingUserID, "Mutual interest", myTitle, myMessage, &productID, &ownerID)
	}

}

//classic get all notificaitons present

func GetMyNotifications(ctx *gin.Context) {
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

	//parse pagination parameters

	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")
	unreadOnly := ctx.DefaultQuery("unread_only", "false")

	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)

	if err != nil || offset < 0 {
		offset = 0
	}

	//build query

	var query string

	var args []any

	argIndex := 1

	query = `
	    SELECT notification_id, notification_type, title, message, related_conversation_id, related_product_id,
		related_user_id, is_read, is_pushed, created_at, updated_at FROM notifications
		WHERE user_id = $1
	`

	args = append(args, user.User_ID)

	//filter unread messages only if requested

	if unreadOnly == "true" {
		query += "AND is_read = false"
	}

	query += " ORDER BY created_at DESC"
	query += " LIMIT $" + strconv.Itoa(argIndex)

	args = append(args, limit+1)
	argIndex++

	query += " OFFSET $" + strconv.Itoa(argIndex)
	args = append(args, offset)

	rows, err := config.DB.Query(query, args...)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch notifcations")
		return
	}

	defer rows.Close()

	var notficiations []models.Notifications

	for rows.Next() {
		var notification models.Notifications

		err := rows.Scan(&notification.Notification_ID, &notification.User_ID, &notification.Notification_type, &notification.Title,
			&notification.Message, &notification.Related_conversation_ID, &notification.Related_product_ID, &notification.Related_user_ID,
			&notification.Is_Read, &notification.Is_Pushed, &notification.Created_at, &notification.Read_at)

		if err != nil {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to parse notification")
			return
		}

		notification.User_ID = user.User_ID

		notficiations = append(notficiations, notification)
	}

	//chech pagination

	hasMore := len(notficiations) > limit

	if hasMore {
		notficiations = notficiations[:limit]
	}

	//celculate next offset

	var nextOffset *int

	if hasMore {
		next := offset + limit
		nextOffset = &next
	}

	response := models.InfiniteScrollData{
		Items: notficiations,
		Meta: models.PaginationMeta{
			Limit:       limit,
			Offset:      offset,
			Has_more:    hasMore,
			Next_offset: nextOffset,
		},
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Notificaitons successfully fetched", response)
}

func MarkNotifcationAsRead(ctx *gin.Context) {
	notificationIDStr := ctx.Param("notification_id")
	notificationID, err := uuid.Parse(notificationIDStr)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid notification ID")
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

	//verify notifcation belongs to user and update

	result, err := config.DB.Exec(`
       UPDATE notifcations
	   SET is_read = true, read_at = $1
	   WHERE notifcation_id = $2 AND user_id = $3 AND is_read = false	
	`, time.Now(), notificationID, user.User_ID)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to mark notification as read")
		return
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Database error")
		return
	}

	if rowsAffected == 0 {
		utils.ErrorResponse(ctx, http.StatusNotFound, "Notification not found")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Notifcation successfully marked as read", nil)
}

func MarkAllNotificationsAsRead(ctx *gin.Context) {
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

	//mark all notification as read

	_, err := config.DB.Exec(`
       UPDATE notifications
	   SET is_read = true, read_at = $1
	   WHERE user_id = $2 AND is_read = false	
	`, time.Now(), user.User_ID)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to mark notifications as read")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "All notifications marked as read", nil)
}

func GetUnreadNotificationCount(ctx *gin.Context) {
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

	var count int

	err := config.DB.QueryRow(`
	   SELECT COUNT(*) FROM notifications
	   WHERE user_id =  $1 AND is_read = false
	`, user.User_ID).Scan(&count)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get count")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Unread Count recieved", gin.H{
		"unread_count": count,
	})
}
