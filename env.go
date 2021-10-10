package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var port string
var dbUser string
var dbPassword string
var dbName string

func parseEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port = os.Getenv("port")
	dbUser = os.Getenv("dbUser")
	dbPassword = os.Getenv("dbPassword")
	dbName = os.Getenv("dbName")
}
