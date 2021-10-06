package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var port string

func parseEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port = os.Getenv("port")
}
