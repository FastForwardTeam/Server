package main

import (
	"database/sql"
	"fmt"
	"log"

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

// returns bool and string
func dbQuery(domain string, path string) (bool, string) {

	stmt, err := db.Prepare("SELECT destination FROM links WHERE domain = ? AND path = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	var dest string
	switch err = stmt.QueryRow(domain, path).Scan(&dest); err {
	case sql.ErrNoRows:
		return false, "not found in database"
	case nil:
		return true, dest
	default:
		panic(err)
	}

}
