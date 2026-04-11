package handlers

import (
	"database/sql"
	"net/http"
	"postswapapi/config"
	"postswapapi/models"
	"postswapapi/services"
	"postswapapi/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func OAuthLogin(ctx *gin.Context) {
	var req models.OAuthLoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// TODO: Verify Firebase token (we'll add this next)
	// For now, assume token is valid and extract email
	email := extractEmailFromFirebaseToken(req.FirebaseToken)

	if email == "" {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "invalid token")
		return
	}

	// Check if user exists
	var user models.Users
	err := config.DB.QueryRow(`
		SELECT user_id, email, first_name, last_name, avatar_url, created_at
		FROM users
		WHERE email = $1
	`, email).Scan(&user.User_ID, &user.Email, &user.First_Name, &user.Last_Name, &user.Avatar_url, &user.Created_at)

	profileComplete := false

	// User doesn't exist - create new user
	if err == sql.ErrNoRows {
		user.User_ID = uuid.New()
		user.Email = email
		user.Created_at = time.Now()

		_, err = config.DB.Exec(`
			INSERT INTO users (user_id, email, created_at)
			VALUES ($1, $2, $3)
		`, user.User_ID, user.Email, user.Created_at)

		if err != nil {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, "failed to create user")
			return
		}

		profileComplete = false
	} else if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "database error")
		return
	} else {
		// User exists - check if profile is complete
		profileComplete = user.First_Name != "" && user.Last_Name != ""
	}

	// Generate JWT token
	token, err := services.GenerateToken(&user)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "failed to generate token")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "login successful", models.OAuthLoginResponse{
		Token:           token,
		UserID:          user.User_ID.String(),
		Email:           user.Email,
		ProfileComplete: profileComplete,
	})
}

// Placeholder - we'll implement Firebase verification next
func extractEmailFromFirebaseToken(token string) string {
	// TODO: Verify Firebase token and extract email
	return "test@example.com" // Temporary
}
