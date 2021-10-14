package main

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var port string
var dbUser string
var dbPassword string
var dbName string
var logFile string
var ipList []string

func parseEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port = os.Getenv("port")
	usrdir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	logFile = strings.ReplaceAll(os.Getenv("log_file"), "USER_DIR", usrdir)
	ipList = strings.Split(os.Getenv("banned_ip_list"), "\n")
	dbName = os.Getenv("dbName")
	dbUser = os.Getenv("dbUser")
	dbPassword = os.Getenv("dbPassword")
}
