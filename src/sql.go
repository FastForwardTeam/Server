/*
Copyright 2021, 2022 NotAProton

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
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	db       *sql.DB
	errnoEnt error = errors.New("no entries")
)

func connectDb() error {
	var err error

	creds := fmt.Sprintf("%s:%s@tcp(db:3306)/%s", dbUser, dbPassword, dbName)

	db, err = sql.Open("mysql", creds)
	if err != nil {
		return err
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	return nil
}

// returns (error, if exists, destination if exists, votedfordeletion 0/1)
func dbQuery(domain string, path string) (bool, string, int, error) {
	stmt, err := db.Prepare("SELECT destination, votedfordeletion FROM links WHERE domain = ? AND path = ?")
	if err != nil {
		return false, "", 0, err
	}
	defer stmt.Close()
	var dest string
	var votedfordeletion int
	switch err = stmt.QueryRow(domain, path).Scan(&dest, &votedfordeletion); err {
	case sql.ErrNoRows:
		return false, "", 0, nil
	case nil:
		return true, dest, votedfordeletion, nil
	default:
		return false, "", 0, err
	}

}

func dbAdminVoteQuery(domain string, path string) (votedfordeletion int, votedBy string, err error) {
	stmt, err := db.Prepare("SELECT votedfordeletion, voted_by FROM links WHERE domain = ? AND path = ?")
	if err != nil {
		return 0, "", err
	}
	defer stmt.Close()
	var voted_by_safe sql.NullString //Allows null values
	switch err = stmt.QueryRow(domain, path).Scan(&votedfordeletion, &voted_by_safe); err {
	case sql.ErrNoRows:
		return 0, "", errnoEnt
	case nil:
		if voted_by_safe.Valid {
			return votedfordeletion, voted_by_safe.String, nil //If the value is valid(not null), return it
		}
		return votedfordeletion, "", nil //otherwise return empty string
	default:
		return 0, "", err
	}
}

func dbInsert(domain string, path string, target string, hashedIP string) error {

	stmt, err := db.Prepare("INSERT INTO links (domain, path, destination, hashed_IP) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(domain, path, target, hashedIP)

	return err

}

//Increases times_reported by 1
func dbReport(domain string, path string) error {

	stmt, err := db.Prepare("UPDATE links SET times_reported = times_reported + 1 WHERE domain = ? AND path = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(domain, path)

	return err
}

// Admin stuff

func dbAdminSoftDelete(domain string, path string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	{
		stmt, err := tx.Prepare(`INSERT INTO recycle_bin (id, domain, path, destination, times_reported, hashed_IP, votedfordeletion, voted_by)
							SELECT *
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
	logger.Println("deleted")

	return tx.Commit()
}

func dbAdminVoteDelete(username string, domain string, path string) error {

	stmt, err := db.Prepare("UPDATE links SET votedfordeletion = 1, voted_by = ? WHERE domain = ? AND path = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, domain, path)

	return err
}

/*
func dbAdminCredsInsert(username string, password string) error {

	stmt, err := db.Prepare("INSERT INTO admin_creds (username, password) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, password)

	return err

}
*/

func dbAdminPasswordChange(username string, newPassword string) error {

	stmt, err := db.Prepare("UPDATE admin_creds SET password = ? WHERE username = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(newPassword, username)

	return err
}

func dbAdminRefTokenInsert(username string, uuid string) error {

	stmt, err := db.Prepare("UPDATE admin_creds SET token_id = ? WHERE username = ?;")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(uuid, username)

	return err

}

// Returns (user exists, password if exists, error)
func dbAdminCredsQuery(username string) (bool, string, error) {

	stmt, err := db.Prepare("SELECT password FROM admin_creds WHERE username = ?")
	if err != nil {
		return false, "", err
	}
	defer stmt.Close()
	var password string
	switch err = stmt.QueryRow(username).Scan(&password); err {
	case sql.ErrNoRows:
		return false, "", nil
	case nil:
		return true, password, nil
	default:
		return false, "", nil
	}

}

// Returns (user exists) (Ref token if exists) (error)
func dbAdminRefTokenQuery(username string) (bool, string, error) {

	stmt, err := db.Prepare("SELECT token_id FROM admin_creds WHERE username = ?")
	if err != nil {
		logger.Fatal(err)
	}
	defer stmt.Close()
	var token string
	switch err = stmt.QueryRow(username).Scan(&token); err {
	case sql.ErrNoRows:
		return false, "", nil
	case nil:
		return true, token, nil
	default:
		return false, "", err
	}

}

// Return reported links as json array with id, domain, path, destination, times_reported, hashed_ip, votedfordeletion, voted_by
func dbQueryReported(page int) ([]byte, error) {

	stmt, err := db.Prepare("SELECT * FROM links WHERE times_reported > 0 LIMIT ? OFFSET ?")
	if err != nil {
		logger.Fatal(err)
	}
	defer stmt.Close()
	page -= 1
	rows, err := stmt.Query(20, page*20)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type Entry struct {
		Id               int    `json:"id"`
		Domain           string `json:"domain"`
		Path             string `json:"path"`
		Destination      string `json:"destination"`
		Times_reported   int    `json:"times_reported"`
		Hashed_IP        string `json:"hashed_IP"`
		Votedfordeletion bool   `json:"votedfordeletion"`
		Voted_by         string `json:"voted_by"`
	}

	var entries []Entry

	for rows.Next() {
		var id, times_reported, tempVotedfordeletion int
		var domain, path, destination, hashed_IP, voted_by string
		var votedfordeletion bool

		rows.Scan(&id, &domain, &path, &destination, &times_reported, &hashed_IP, &tempVotedfordeletion, &voted_by)
		// sorry :p
		if tempVotedfordeletion == 0 {
			votedfordeletion = false
		} else {
			votedfordeletion = true
		}

		// (hopefully) prevent xss
		domain = bmStrict.Sanitize(domain)
		path = bmStrict.Sanitize(path)
		destination = bmStrict.Sanitize(destination)
		hashed_IP = bmStrict.Sanitize(hashed_IP)
		voted_by = bmStrict.Sanitize(voted_by)

		entries = append(entries, Entry{id, domain, path, destination, times_reported, hashed_IP, votedfordeletion, voted_by})
	}
	if len(entries) == 0 {
		return nil, errnoEnt
	}
	entryBytes, _ := json.MarshalIndent(&entries, "", "  ")

	return entryBytes, err
}
