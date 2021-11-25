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
	"crypto/sha1"
	"encoding/hex"
	"io"
	"math/rand"
	"net/http"

	anyascii "github.com/anyascii/go"
)

func sanitize(s ...*string) {
	for _, i := range s {
		*i = anyascii.Transliterate(*i)
		*i = reg.ReplaceAllString(*i, "")
	}
}

func sha256(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
func getRequestId(r *http.Request) string {
	requestID, ok := r.Context().Value(requestIDKey).(string)
	if !ok {
		requestID = "unknown"
	}
	return requestID
}

func getUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}

func isIPblacklisted(hash string) bool {
	for _, b := range ipList {
		if b == hash {
			return true
		}
	}
	return false
}

func crowdQueryV1(w http.ResponseWriter, r *http.Request) {
	ref := r.Referer()
	sanitize(&ref)

	logger.Println(getRequestId(r), "["+r.Method+"] ", r.URL.String(), "Referer", ref)

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	n := rand.Intn(10)
	if n == 1 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	err := r.ParseForm()
	if err != nil {
		logger.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	d, p := r.FormValue("domain"), r.FormValue("path")
	sanitize(&d, &p)
	exists, path, votedfordeletion := dbQuery(d, p)
	if exists && votedfordeletion == 0 {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, path)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}

}

func crowdContributeV1(w http.ResponseWriter, r *http.Request) {

	hip := sha256(getUserIP(r))
	if isIPblacklisted(hip) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	logger.Println("["+r.Method+"] ", r.URL.String(), "Referer", r.Referer())
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	d, p, t := r.FormValue("domain"), r.FormValue("path"), r.FormValue("target")
	sanitize(&d, &p, &t)
	if d == "" || p == "" || t == "" {
		w.WriteHeader(http.StatusBadRequest)
		logger.Println(getRequestId(r) + " rejected crowd contribution [Illegal characters] ")
		return
	}
	exists, destination, _ := dbQuery(d, p)
	if exists {
		if t != destination {
			dbReport(d, p)
		}
		w.WriteHeader(http.StatusCreated)
		return
	}

	dbInsert(d, p, t, hip)
	w.WriteHeader(http.StatusCreated)
}
