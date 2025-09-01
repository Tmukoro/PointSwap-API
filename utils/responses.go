package utils

import "github.com/gin-gonic/gin"

//model for response properties

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}


//Function to relay if the operation was a success
func SuccessResponse(ctx *gin.Context, status int, message string, data any) {
	ctx.JSON(status, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

//function to relay if there was an error
func ErrorResponse(ctx *gin.Context, status int, message string) {
	ctx.JSON(status, Response{
		Success: false,
		Message: message,
	})
}
