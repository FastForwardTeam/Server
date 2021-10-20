//TODO: Return errors for all funcs

package main

import (
	"database/sql"
	"encoding/json"
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

// returns (wether exists), (path if it exists)
func dbQuery(domain string, path string) (bool, string) {

	stmt, err := db.Prepare("SELECT destination FROM links WHERE domain = ? AND path = ?")
	if err != nil {
		logger.Fatal(err)
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
		logger.Fatal(err)
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
		logger.Fatal(err)
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
		logger.Fatal(err)
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

// Returns (user exists) (password if exists)

func dbQueryReported() ([]byte, error) {
	rows, err := db.Query("SELECT * FROM links WHERE reports > 0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type Entry struct {
		Id          int    `json:"id"`
		Domain      string `json:"domian"`
		Path        string `json:"path"`
		Destination string `json:"destination"`
		Hashed_IP   string `json:"hashed_IP"`
		Reports     int    `json:"reports"`
	}

	var users []Entry

	for rows.Next() {
		var id, reports int
		var domain, path, destination, hashed_IP string

		rows.Scan(&id, &domain, &path, &destination, &hashed_IP, &reports)
		users = append(users, Entry{id, domain, path, destination, hashed_IP, reports})
	}

	entryBytes, _ := json.Marshal(&users)

	return entryBytes, nil
}

func dbReport(domain string, path string) {

}
