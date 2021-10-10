package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func connectDb() {
	var err error

	creds := fmt.Sprintf("%s:%s@/%s", dbUser, dbPassword, dbName)

	db, err = sql.Open("mysql", creds)
	if err != nil {
		panic(err)
	}
}

func dbQuery() string {
	return "test"
}
