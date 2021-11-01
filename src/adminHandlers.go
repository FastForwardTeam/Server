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

//TODO: add env var to check if signup is allowed
package main

//NOT READY YET

import (
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func adminPanelRouters(h *http.ServeMux) {
	fs := http.FileServer(http.Dir(panelDir))
	h.Handle("/admin/", http.StripPrefix("/admin/", fs))

	h.HandleFunc("/admin/api/signup", signUp)
	h.HandleFunc("/admin/api/reported", returnReported)
	h.HandleFunc("/admin/api/newreftoken", refTokenHandler)
	h.HandleFunc("/admin/api/newacctoken", accTokenHandler)

}

func signUp(w http.ResponseWriter, r *http.Request) {

	type Input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var input Input

	// decode input
	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	alreadyExists, _ := dbAdminCredsQuery(input.Username)
	if alreadyExists {
		w.WriteHeader(http.StatusConflict)
		return
	}
	// Hashing the password with the default cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dbAdminCredsInsert(input.Username, string(hashedPassword))
	logger.Println(input.Username + " signed up as admin")
	w.WriteHeader(http.StatusOK)
}

func returnReported(w http.ResponseWriter, r *http.Request) {
	username, valid := parseAuthHeader(r)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	logger.Println(username + " accessed /admin/api/reported")
	jsonData, err := dbQueryReported()
	if err != nil {
		panic(err)
	}

	w.Write(jsonData)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

}

// takes username and password, returns refresh token
func refTokenHandler(w http.ResponseWriter, r *http.Request) {

	type Input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var input Input

	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	exists, hashedpassword := dbAdminCredsQuery(input.Username)

	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedpassword), []byte(input.Password))

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	refTokenString := genRefToken(input.Username)

	type Response struct {
		RefToken string `json:"reftoken"`
	}

	w.Header().Set("Content-Type", "application/json")
	resp := Response{
		RefToken: refTokenString,
	}
	j, _ := json.MarshalIndent(resp, "", "  ")
	w.Write(j)
	w.WriteHeader(http.StatusOK)
}

// takes refresh token, returns access token
func accTokenHandler(w http.ResponseWriter, r *http.Request) {
	type Input struct {
		RefToken string `json:"reftoken"`
	}

	var input Input

	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	accTokenString, err := genAccessToken(input.RefToken)

	if err != nil {
		if err != nil {
			logger.Println(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	type Response struct {
		AccToken string `json:"acctoken"`
	}

	w.Header().Set("Content-Type", "application/json")
	resp := Response{
		AccToken: accTokenString,
	}
	j, _ := json.MarshalIndent(resp, "", "  ")
	w.Write(j)
	w.WriteHeader(http.StatusOK)
}

func parseAuthHeader(r *http.Request) (username string, ok bool) {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) != 2 {
		return "", false
	}
	reqToken = splitToken[1]
	username, err := pasreAccessToken(reqToken)
	if err != nil {
		return "", false
	}
	return username, true
}
