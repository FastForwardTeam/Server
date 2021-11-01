/*
Copyright 2021 NotAProton

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
		logger.Fatalln(err)
	}
}

// returns (if exists), (destination if exists)
func dbQuery(domain string, path string) (bool, string) {
	stmt, err := db.Prepare("SELECT destination FROM links WHERE domain = ? AND path = ?")
	if err != nil {
		logger.Fatalln(err)
	}
	defer stmt.Close()
	var dest string
	switch err = stmt.QueryRow(domain, path).Scan(&dest); err {
	case sql.ErrNoRows:
		return false, ""
	case nil:
		return true, dest
	default:
		logger.Fatalln(err)
		return false, "" //Fatal stops the program, return is unnecessary
	}

}

func dbInsert(domain string, path string, target string, hashedIP string) {

	stmt, err := db.Prepare("INSERT INTO links (domain, path, destination, hashed_IP) VALUES (?, ?, ?, ?)")
	if err != nil {
		logger.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(domain, path, target, hashedIP)

	if err != nil {
		logger.Fatalln(err)
	}
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

func dbAdminCredsInsert(username string, password string) {

	stmt, err := db.Prepare("INSERT INTO admin_creds (username, password) VALUES (?, ?)")
	if err != nil {
		logger.Fatalln(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, password)

	if err != nil {
		logger.Fatalln(err)
	}

}

func dbAdminRefTokenInsert(username string, uuid string) {

	stmt, err := db.Prepare("UPDATE admin_creds SET token_id = ? WHERE username = ?;")
	if err != nil {
		logger.Fatalln(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(uuid, username)

	if err != nil {
		logger.Fatalln(err)
	}

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
		logger.Fatalln(err)
		return false, ""
	}

}

// Returns (user exists) (Ref token if exists)
func dbAdminRefTokenQuery(username string) (bool, string) {

	stmt, err := db.Prepare("SELECT token_id FROM admin_creds WHERE username = ?")
	if err != nil {
		logger.Fatal(err)
	}
	defer stmt.Close()
	var token string
	switch err = stmt.QueryRow(username).Scan(&token); err {
	case sql.ErrNoRows:
		return false, ""
	case nil:
		return true, token
	default:
		logger.Fatalln(err)
		return false, ""
	}

}

// Returns (user exists) (password if exists)

func dbQueryReported() ([]byte, error) {
	rows, err := db.Query("SELECT * FROM links WHERE times_reported > 0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type Entry struct {
		Id             int    `json:"id"`
		Domain         string `json:"domian"`
		Path           string `json:"path"`
		Destination    string `json:"destination"`
		Times_reported int    `json:"times_reported"`
		Hashed_IP      string `json:"hashed_IP"`
	}

	var users []Entry

	for rows.Next() {
		var id, times_reported int
		var domain, path, destination, hashed_IP string

		rows.Scan(&id, &domain, &path, &destination, &times_reported, &hashed_IP)
		users = append(users, Entry{id, domain, path, destination, times_reported, hashed_IP})
	}

	entryBytes, _ := json.MarshalIndent(&users, "", "  ")

	return entryBytes, nil
}

//Increases times_reported 1
func dbReport(domain string, path string) {

	stmt, err := db.Prepare("UPDATE links SET times_reported = times_reported + 1 WHERE domain = ? AND path = ?")
	if err != nil {
		logger.Fatalln(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(domain, path)

	if err != nil {
		logger.Fatalln(err)
	}
}
