package main

import (
	"log"
	"os"
	"postswapapi/config"
	"postswapapi/routes"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

//@title PointSwap API
//@version 1.0

//@

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Print("env file not found")
	}

	config.ConnectToDb()

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Println("Server running on port ", port)

	r := routes.SetupRouter()
	r.Run(":" + port)

}
