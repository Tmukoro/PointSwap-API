package routes

import (
	"postswapapi/handlers"
	"postswapapi/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	api := r.Group("/pointSwapApi/v1")

	api.GET("/userProfile", middleware.AuthMiddleWare(), handlers.UserProfile)

	api.POST("/register", handlers.Register)

	api.POST("/profileSetUp", middleware.AuthMiddleWare(), handlers.UserProfileSetUp)

	api.POST("/login", handlers.Login)

	api.POST("/location", middleware.AuthMiddleWare(), handlers.GetLocation)

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

	notification := api.Group("/notifications")
	{
		notification.GET("", middleware.OptionalAuthMiddleWare(), handlers.GetMyNotifications)
		notification.PATCH("/:notification_id", middleware.AuthMiddleWare(), handlers.MarkNotifcationAsRead)
	}

	api.GET("/me", middleware.AuthMiddleWare())

	return r

}
