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

//Handler to handle the registration process within the db

func Register(ctx *gin.Context) {
	var req models.UserRegistrationRequest

	if err := ctx.ShouldBind(&req); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	//Check if the user exists in the db when registering
	var existingID uuid.UUID

	err := config.DB.QueryRow("SELECT user_id FROM users WHERE email = $1",
		req.Email).Scan(&existingID)

	if err != sql.ErrNoRows {
		utils.ErrorResponse(ctx, http.StatusConflict, err.Error())
		return
	}

	//Check for existing phone number

	err = config.DB.QueryRow(`SELECT user_id FROM users WHERE phone_number = $1`, req.Phone_number).Scan(&existingID)

	if err != sql.ErrNoRows {
		utils.ErrorResponse(ctx, http.StatusConflict, "Phone number already in use by another account")
		return
	}

	//Convert the password inputed by the user into a hash password
	hashedPassword, err := services.GenerateHashPassword(req.Password)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Couldn't hash password")
		return
	}

	//Create a user variable of the type Users models

	user := models.Users{
		User_ID:      uuid.New(),
		Username:     req.Username,
		Display_name: req.Display_name,
		Phone_number: req.Phone_number,
		Email:        req.Email,
		Created_at:   time.Now(),
		Updated_at:   time.Now(),
	}

	//Insert into the db the credentials from the registration model

	_, err = config.DB.Exec(`INSERT INTO users (user_id, username, display_name, phone_number, email, password_hash, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		user.User_ID, user.Username, user.Display_name, user.Phone_number, user.Email, hashedPassword, user.Created_at, user.Updated_at)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create user")
	}

	token, err := services.GenerateToken(&user)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.SuccessResponse(ctx, http.StatusCreated, "User Created Successfully", gin.H{
		"user":  user,
		"token": token,
	})

}

// Handler to handle the login process within the db
func Login(ctx *gin.Context) {
	var req models.UserLoginRequest

	if err := ctx.ShouldBind(&req); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	var user models.Users
	var passwordHash string

	//Goes into the db to return a row that contains the relation between the inputed credentials and the one in the db

	err := config.DB.QueryRow(`
	  SELECT user_id, username, display_name, phone_number, email, password_hash, created_at, updated_at FROM users
	  WHERE email = $1
	`, req.Email).Scan(
		&user.User_ID, &user.Username, &user.Display_name, &user.Phone_number, &user.Email, &passwordHash,
		&user.Created_at, &user.Updated_at,
	)

	//Basically returns an error if the input you put in is wrong

	if err == sql.ErrNoRows {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid Credentials")
		return
	}

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Database error")
		return
	}

	if !services.CheckHashPassword(req.Password, passwordHash) {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid Credentials")
		return
	}

	token, err := services.GenerateToken(&user)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Login Successful", gin.H{
		"user":  user,
		"token": token,
	})

}

//Handler to check if the user exists

func GetUser(ctx *gin.Context) {
	user, exist := ctx.Get("User")

	if !exist {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not authenticated")
	}

	utils.SuccessResponse(ctx, http.StatusOK, "User Found", user)
}
