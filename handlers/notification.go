package handlers

import (
	"net/http"
	"postswapapi/models"
	"postswapapi/services"
	"postswapapi/utils"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationService *services.NotificationService
}

func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// GetUserNotifications retrieves all notifications for a user
func (h *NotificationHandler) GetUserNotifications(c *gin.Context) {
	presentUser, exists := c.Get("User")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, ok := presentUser.(models.Users)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid User")
		return
	}

	notifications, err := h.notificationService.GetUserNotifications(user.User_ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to get notifications")
		return
	}

	// Ensure we always return an array, even if empty
	if notifications == nil {
		notifications = []models.NotificationWithUserDetails{}
	}

	utils.SuccessResponse(c, http.StatusOK, "notifications retrieved successfully", gin.H{
		"notifications": notifications,
	})
}

// MarkNotificationAsRead marks a notification as read
func (h *NotificationHandler) MarkNotificationAsRead(c *gin.Context) {
	presentUser, exists := c.Get("User")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, ok := presentUser.(models.Users)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid User")
		return
	}

	notificationID := c.Param("id")

	err := h.notificationService.MarkAsRead(notificationID, user.User_ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to mark notification as read")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "notification marked as read", nil)
}

// GetUnreadCount gets the count of unread notifications
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	presentUser, exists := c.Get("User")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, ok := presentUser.(models.Users)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid User")
		return
	}

	count, err := h.notificationService.GetUnreadCount(user.User_ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to get unread count")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "unread count retrieved", gin.H{
		"unread_count": count,
	})
}
