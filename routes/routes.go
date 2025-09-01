package routes

import (
	"postswapapi/handlers"
	"postswapapi/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	api := r.Group("/pointSwapApi/v1")

	api.POST("/register", handlers.Register)

	api.POST("/login", handlers.Login)

	product := api.Group("/products")
	{
		product.POST("", middleware.AuthMiddleWare(), handlers.CreateProduct)
		product.GET("", middleware.OptionalAuthMiddleWare(), handlers.GetProducts)
		product.GET("/me", middleware.AuthMiddleWare(), handlers.GetMyProducts)
		product.GET("/:product_id", middleware.OptionalAuthMiddleWare(), handlers.GetProductById)
		product.POST("/:product_id", middleware.AuthMiddleWare(), handlers.UpdateProductStatus)
		product.DELETE("/:product_id", middleware.AuthMiddleWare(), handlers.DeleteProduct)

	}

	productWant := api.Group("/product_wants")
	{
		productWant.POST("/:product_id/want", middleware.AuthMiddleWare(), handlers.CreateProductWant)
		productWant.GET("/:product_id/want", middleware.OptionalAuthMiddleWare(), handlers.GetProductWant)
		productWant.PUT("/:product_id/want", middleware.AuthMiddleWare(), handlers.UpdateProductWant)
	}

	conversation := api.Group("/conversations")
	{
		conversation.POST("", middleware.AuthMiddleWare(), handlers.CreateConversation)
		conversation.GET("", middleware.OptionalAuthMiddleWare(), handlers.GetMyConversations)
		conversation.GET("/:conversation_id", middleware.OptionalAuthMiddleWare(), handlers.GetConversationByID)
	}

	message := api.Group("/messages")
	{
		message.POST("", middleware.AuthMiddleWare(), handlers.SendMessage)
		message.GET("/:conversation_id", middleware.OptionalAuthMiddleWare(), handlers.GetMessages)
		message.PUT("/:message_id", middleware.AuthMiddleWare(), handlers.EditMessage)
		message.DELETE("/:message_id", middleware.AuthMiddleWare(), handlers.DeleteMessage)
	}

	notification := api.Group("/notifications")
	{
		notification.GET("", middleware.OptionalAuthMiddleWare(), handlers.GetMyNotifications)
		notification.PATCH("/:notification_id", middleware.AuthMiddleWare(), handlers.MarkNotifcationAsRead)
	}

	api.GET("/me", middleware.AuthMiddleWare())

	return r

}
