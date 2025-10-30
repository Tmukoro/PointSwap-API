package handlers

import (
	"database/sql"
	"fmt"
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

	//Convert the password inputed by the user into a hash password
	hashedPassword, err := services.GenerateHashPassword(req.Password)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Couldn't hash password")
		return
	}

	//Create a user variable of the type Users models

	user := models.Users{
		User_ID:    uuid.New(),
		Email:      req.Email,
		Created_at: time.Now(),
		Updated_at: time.Now(),
	}

	//Insert into the db the credentials from the registration model

	_, err = config.DB.Exec(`INSERT INTO users (user_id, email, password_hash, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
		user.User_ID, user.Email, hashedPassword, user.Created_at, user.Updated_at)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
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

func UserProfileSetUp(ctx *gin.Context) {
	var req models.UserProfileSetUpRequest

	if err := ctx.ShouldBind(&req); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		fmt.Println(err)
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

	_, err := config.DB.Exec(`
	   UPDATE USERS
	   SET first_name = $1, last_name = $2, avatar_url = $3
	   WHERE user_id = $4
	`, req.First_Name, req.Last_Name, req.Avatar_url, user.User_ID)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	SetUp := models.Users{
		First_Name: req.First_Name,
		Last_Name:  req.Last_Name,
		Avatar_url: &req.Avatar_url,
	}

	utils.SuccessResponse(ctx, http.StatusAccepted, "Profile Successfully Set Up", gin.H{
		"profile_Update": SetUp,
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
	  SELECT user_id, email, password_hash, created_at, updated_at FROM users
	  WHERE email = $1
	`, req.Email).Scan(
		&user.User_ID, &user.Email, &passwordHash,
		&user.Created_at, &user.Updated_at,
	)

	//Basically returns an error if the input you put in is wrong

	if err == sql.ErrNoRows {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid Credentials")
		return
	}

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
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

//Get location of the user

func GetLocation(ctx *gin.Context) {
	var req models.UserLocationRequest

	if err := ctx.ShouldBind(&req); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	presentUser, exists := ctx.Get("User")

	if !exists {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not authentiticated")
		return
	}

	user, ok := presentUser.(models.Users)

	if !ok {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Invalid User")
		return
	}

	location := models.Users{
		Location: req.Location,
	}

	_, err := config.DB.Exec(`
	   UPDATE users
	   SET location = $1 
	   WHERE user_id = $2
	`, req.Location, user.User_ID)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to add location")
		return
	}

	utils.SuccessResponse(ctx, http.StatusCreated, "location Added", gin.H{
		"location": location,
		"user":     user,
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

func UserProfile(ctx *gin.Context) {

	presentUser, exists := ctx.Get("User")

	if !exists {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, ok := presentUser.(models.Users)

	if !ok {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "User not found")
		return
	}

	var User models.Users

	err := config.DB.QueryRow(`
	    SELECT user_id, email, first_name, last_name, avatar_url FROM users
		WHERE user_id = $1 
	`, user.User_ID).Scan(&User.User_ID, &User.Email, &User.First_Name, &User.Last_Name, &User.Avatar_url)

	if err == sql.ErrNoRows {
		utils.ErrorResponse(ctx, http.StatusNotFound, "User not found")
		return
	}

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch User Details")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "User Details Fetched", User)

}
