package handlers

import (
	"database/sql"
	"net/http"
	"postswapapi/config"
	"postswapapi/models"
	"postswapapi/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//update this is now scratched this kept for future purposes and currently unasable so anything invovling this handler is unnecessary as of now

/*
this handler is for a wishlist function where the backend scans the db for users who have what you want regardless

	of if you have what they have (probably this will serve as a source of negotiation)
*/
func CreateUserPreference(ctx *gin.Context) {
	var req models.CreateUserPreferenceRequest

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
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Invald User")
		return
	}

	//check if similar preference already exists

	var existingID string

	err := config.DB.QueryRow(`
	   SELECT preference_id FROM user_preferences WHERE user_id = $1 AND category = $2 AND
	    (size = $3 OR (size is NULL AND $3 IS NULL)) AND is_active = true
	`, user.User_ID, req.Category, req.Size).Scan(&existingID)

	if err == nil {
		utils.ErrorResponse(ctx, http.StatusConflict, "You already have this preference in your wishlist")
		return
	}

	if err != sql.ErrNoRows {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Database error")
		return
	}

	//create the new preference

	preference := models.UserPreferences{
		PreferenceID: uuid.New(),
		UserID:       user.User_ID,
		Category:     req.Category,
		Size:         req.Size,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = config.DB.Exec(`
        INSERT INTO user_preferences (preference_id, user_id, category, size, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)	
	`, preference.PreferenceID, preference.UserID, preference.Category, preference.Size, preference.IsActive,
		preference.CreatedAt, preference.UpdatedAt)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to save user preference")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Wishlist successfully created", gin.H{
		"preference_id": preference.PreferenceID,
	})

}

//Classic get users wishlist items on the wishlist feed

func GetUserPreferences(ctx *gin.Context) {
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

	//Get active preferences only

	rows, err := config.DB.Query(`
       SELECT preference_id, user_id, category, size, is_active, created_at, updated_at
	   FROM user_preferences WHERE user_id = $1 AND is_active = "true"
	   ORDER BY created_at DESC	
	`, user.User_ID)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Database Error")
		return
	}

	defer rows.Close()

	var preferences []models.UserPreferences

	for rows.Next() {
		var preference models.UserPreferences

		err := rows.Scan(&preference.PreferenceID, &preference.UserID, &preference.Category, &preference.Size,
			&preference.IsActive, &preference.CreatedAt, &preference.UpdatedAt)

		if err != nil {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to scan preferences")
			return
		}

		preferences = append(preferences, preference)
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved Preferences", preferences)
}

//update user preference in wishlist

func UpdateUserPreference(ctx *gin.Context) {
	var req models.UpdateUserPreferenceRequest

	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	preferenceID := ctx.Param("preference_id")

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

	//verify ownership and update

	result, err := config.DB.Exec(`
	   UPDATE user_preferences SET category = $3, size = $4, is_active = $5, updated_at = $6 WHERE
	   preference_id = $1 AND user_id = $2
	`, preferenceID, user.User_ID, req.Category, req.Size, req.IsActive, time.Now())

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update preference")
		return
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		utils.ErrorResponse(ctx, http.StatusNotFound, "Preference not found or you dont own it")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Preference Successfully Updated", nil)

}

// Delete Preference

func DeletePreference(ctx *gin.Context) {

	preferenceID := ctx.Param("preference_id")
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

	result, err := config.DB.Exec(`
       UPDATE user_preferences
	   SET is_active = false, updated_at = $3
	   WHERE preference_id = $1 AND user_id = $2	
	`, preferenceID, user.User_ID, time.Now())

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete preference")
		return
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		utils.ErrorResponse(ctx, http.StatusNotFound, "Preference Not found")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Preference Successfully Removed", nil)
}
