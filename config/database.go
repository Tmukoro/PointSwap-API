package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

var DB *sql.DB

func ConnectToDb() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dataSource := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error

	DB, err = sql.Open("postgres", dataSource)

	if err != nil {
		log.Fatal("failed to connect to DB: ", err)
	}

	log.Print("Successfully connected to DB")
}
