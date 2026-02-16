package main

import (
	"log"
	"os"
	"postswapapi/config"
	"postswapapi/handlers"
	"postswapapi/repository"
	"postswapapi/routes"
	"postswapapi/services"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Print("env file not found")
	}

	config.ConnectToDb()

	// Initialize message components
	messageRepo := repository.NewMessageRepository(config.DB)
	messageService, err := services.NewMessageService(messageRepo)
	if err != nil {
		log.Fatal("Failed to initialize message service:", err)
	}
	defer messageService.Close()

	messageHandler := handlers.NewMessageHandler(messageService)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Println("Server running on port ", port)

	r := routes.SetupRouter(messageHandler)
	r.Run(":" + port)

}
