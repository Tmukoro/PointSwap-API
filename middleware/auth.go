package middleware

import (
	"fmt"
	"net/http"
	"postswapapi/config"
	"postswapapi/models"
	"postswapapi/services"
	"postswapapi/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

//Middleware basically serves as what gives the user authorisation to be in the app when signed in

func AuthMiddleWare() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")

		if authHeader == "" {
			utils.ErrorResponse(ctx, http.StatusUnauthorized, "Authorization Header Required")
			ctx.Abort()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		fmt.Println(bearerToken)

		if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
			utils.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid Authorization Header Format")
			ctx.Abort()
			return
		}

		claims, err := services.ValidateToken(bearerToken[1])
		if err != nil {
			utils.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid Token")
			ctx.Abort()
			return
		}

		var user models.Users
		err = config.DB.QueryRow(`
		SELECT user_id, email, created_at, updated_at FROM users
	    WHERE user_id = $1
		`, claims.UserID).Scan(&user.User_ID, &user.Email, &user.Created_at, &user.Updated_at)

		if err != nil {
			utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not found")
			ctx.Abort()
			return
		}

		ctx.Set("User", user)
		ctx.Next()
	}
}

func OptionalAuthMiddleWare() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.Next()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || strings.ToLower(bearerToken[0]) != "bearer" {
			ctx.Next()
			return
		}

		claims, err := services.ValidateToken(bearerToken[1])
		if err != nil {
			ctx.Next()
			return
		}

		var user models.Users

		err = config.DB.QueryRow(`
		   SELECT user_id, email, created_at, updated_at FROM users
	       WHERE user_id = $1
		`, claims.UserID).Scan(
			&user.User_ID, &user.Email, &user.Created_at, &user.Updated_at,
		)

		if err == nil {
			ctx.Set("User", user)
		}

		ctx.Next()

	}
}
