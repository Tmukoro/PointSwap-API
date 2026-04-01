package handlers

import (
	"net/http"
	"postswapapi/config"
	"postswapapi/models"
	"postswapapi/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RegisterPushToken(c *gin.Context) {
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

	var req models.RegisterPushTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Upsert token (insert or update if exists)
	query := `
		INSERT INTO push_tokens (id, user_id, token, device_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (user_id, token) 
		DO UPDATE SET updated_at = NOW()
	`

	_, err := config.DB.Exec(query, uuid.New(), user.User_ID, req.Token, req.DeviceType)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to register push token")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "push token registered successfully", nil)
}
