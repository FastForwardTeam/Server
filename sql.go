//TODO: Return errors for all funcs

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

// returns (wether exists), (path if it exists)
func dbQuery(domain string, path string) (bool, string) {

	stmt, err := db.Prepare("SELECT destination FROM links WHERE domain = ? AND path = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	var dest string
	switch err = stmt.QueryRow(domain, path).Scan(&dest); err {
	case sql.ErrNoRows:
		return false, ""
	case nil:
		return true, dest
	default:
		panic(err)
	}

}

func dbInsert(domain string, path string, target string, hashedIP string) {

	stmt, err := db.Prepare("INSERT INTO links (domain, path, destination, hashed_IP) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(domain, path, target, hashedIP)

	panic(err)
}

// Admin stuff

func dbSoftDelete(domain string, path string) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	{
		stmt, err := tx.Prepare(`INSERT INTO recycle_bin (id, domain, path, destination, hashed_IP, times_reported)
							SELECT (id, domain, path, destination, hashed_IP, times_reported)
							FROM links
							WHERE domain = ? AND path = ?;`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		if _, err := stmt.Exec(domain, path); err != nil {
			tx.Rollback()
			return err
		}
	}

	{
		stmt, err := tx.Prepare(`DELETE FROM links
							WHERE domain = ? AND path = ?;`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		if _, err := stmt.Exec(domain, path); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func dbAdminCredsInsert(username string, password string) bool {

	stmt, err := db.Prepare("INSERT INTO admin_creds (username, password) VALUES (?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(username, password)
	if err == nil {
		return true
	}
	panic(err)
}

// Returns (user exists) (password if exists)
func dbAdminCredsQuery(username string) (bool, string) {

	stmt, err := db.Prepare("SELECT password FROM admin_creds WHERE username = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	var password string
	switch err = stmt.QueryRow(username).Scan(&password); err {
	case sql.ErrNoRows:
		return false, ""
	case nil:
		return true, password
	default:
		panic(err)
	}

}

func dbReport(domain string, path string) {

}
