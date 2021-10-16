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

func parseEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	usrdir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	port = os.Getenv("port")
	logFile = strings.ReplaceAll(os.Getenv("log_file"), "USER_DIR", usrdir)
	ipList = strings.Split(os.Getenv("banned_ip_list"), "\n")
	dbName = os.Getenv("dbName")
	dbUser = os.Getenv("dbUser")
	dbPassword = os.Getenv("dbPassword")
	panelDir = strings.ReplaceAll(os.Getenv("panel_dir"), "USER_DIR", usrdir)
	log.Println(panelDir)
}
