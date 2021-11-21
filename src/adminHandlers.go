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
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func adminPanelRouters(h *http.ServeMux) {
	fs := http.FileServer(http.Dir(panelDir))
	h.Handle("/admin/", http.StripPrefix("/admin/", fs))

	// h.HandleFunc("/admin/api/signup", adminSignUp)
	h.HandleFunc("/admin/api/changepassword", adminChangePassword)
	h.HandleFunc("/admin/api/newreftoken", refTokenHandler)
	h.HandleFunc("/admin/api/newacctoken", accTokenHandler)

	h.HandleFunc("/admin/api/reported", returnReported)
	h.HandleFunc("/admin/api/votedelete", voteDelete)

}

/* func adminSignUp(w http.ResponseWriter, r *http.Request) {

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
	logger.Println(getRequestId(r) + " " + input.Username + " signed up as admin")
	w.WriteHeader(http.StatusOK)
} */

func adminChangePassword(w http.ResponseWriter, r *http.Request) {

	type Input struct {
		Username    string `json:"username"`
		OldPassword string `json:"oldpassword"`
		NewPassword string `json:"newpassword"`
	}

	var input Input

	// decode input
	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	exists, hashedpassword := dbAdminCredsQuery(input.Username)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("user not found"))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedpassword), []byte(input.OldPassword))

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		logger.Println(getRequestId(r) + " failed to change password for " + input.Username + " [wrong old password]")
		return
	}
	// Hashing the password with the default cost of 10
	hashednewPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dbAdminPasswordChange(input.Username, string(hashednewPassword))
	logger.Println(getRequestId(r) + " " + input.Username + " changed their password")

	dbAdminRefTokenInsert(input.Username, "")
	logger.Println(getRequestId(r) + " revoked refresh token for \"" + input.Username + "\" [reason: password changed]")

	w.WriteHeader(http.StatusOK)
}

func returnReported(w http.ResponseWriter, r *http.Request) {
	_, valid := parseAuthHeader(r)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err := strconv.Atoi(r.FormValue("page"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	jsonData, err := dbQueryReported(p)
	if err != nil {
		logger.Println(getRequestId(r) + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func voteDelete(w http.ResponseWriter, r *http.Request) {
	username, valid := parseAuthHeader(r)
	if !valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	type Input struct {
		Domain string `json:"domain"`
		Path   string `json:"path"`
	}

	var input Input

	// decode input
	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	exists, _, votedfordeletion := dbQuery(input.Domain, input.Path)

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("domain and path not found"))
		return
	}

	if votedfordeletion != 0 {
		logger.Println(votedfordeletion)
		err = dbAdminSoftDelete(input.Domain, input.Path)
		if err != nil {
			logger.Println(getRequestId(r) + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger.Println(getRequestId(r) + " deleted" + input.Domain + input.Path)
		w.WriteHeader(http.StatusAccepted)
		return
	}

	dbAdminVoteDelete(username, input.Domain, input.Path)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
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

	logger.Println(getRequestId(r) + " generated refresh token as " + input.Username)

	type Response struct {
		RefToken string `json:"reftoken"`
	}

	w.Header().Set("Content-Type", "application/json")
	resp := Response{
		RefToken: refTokenString,
	}
	j, _ := json.MarshalIndent(resp, "", "  ")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
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

	username, accTokenString, err := genAccessToken(input.RefToken)

	if err != nil {
		logger.Println(getRequestId(r) + " failed to generate access token, " + err.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	logger.Println(getRequestId(r) + " generated an access token for " + username)

	type Response struct {
		AccToken string `json:"acctoken"`
	}

	w.Header().Set("Content-Type", "application/json")
	resp := Response{
		AccToken: accTokenString,
	}
	j, _ := json.MarshalIndent(resp, "", "  ")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
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
		logger.Println(getRequestId(r) + " failed to authenticate [invalid access token in header] ")
		return "", false
	}
	logger.Println(getRequestId(r) + " authenticated as \"" + username + "\" using an access token")
	return username, true
}
