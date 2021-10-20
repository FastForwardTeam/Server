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
var panelDir string
var USER_DIR string

func parseEnv() {
	err := godotenv.Load("/etc/fastforward/.env")
	if err != nil {
		err = godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
	USER_DIR = os.Getenv("USER_DIR")
	port = os.Getenv("port")
	logFile = strings.ReplaceAll(os.Getenv("log_file"), "USER_DIR", USER_DIR)
	ipList = strings.Split(os.Getenv("banned_ip_list"), "\n")
	dbName = os.Getenv("dbName")
	dbUser = os.Getenv("dbUser")
	dbPassword = os.Getenv("dbPassword")
	panelDir = strings.ReplaceAll(os.Getenv("panel_dir"), "USER_DIR", USER_DIR)
}
