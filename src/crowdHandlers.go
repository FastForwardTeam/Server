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
	"crypto/sha256"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
)

func hashSha256(s string) string {
	sum := sha256.Sum256([]byte(s))
	hexstring := fmt.Sprintf("%x", sum)
	return hexstring
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
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	n := rand.Intn(15)
	if n == 1 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	d, p := r.FormValue("domain"), r.FormValue("path")
	if len(d) >= 254 || len(p) >= 2047 {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}
	d, p = sanitizeDomain(d), sanitizePath(p)
	exists, path, votedfordeletion, err := dbQuery(d, p)

	if err != nil {
		logger.Println(getRequestId(r) + " Error: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if exists && votedfordeletion == 0 {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "https://"+path)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}

}

func crowdContributeV1(w http.ResponseWriter, r *http.Request) {

	hip := hashSha256(getUserIP(r))
	if isIPblacklisted(hip) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

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
	if len(d) >= 254 || len(p) >= 2047 || len(t) >= 4095 {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		logger.Println(getRequestId(r), " Request too large")
		return
	}
	d, p = sanitizeDomain(d), sanitizePath(p)
	tURL, err := url.Parse(t)
	if tURL.Scheme != "https" && tURL.Scheme != "http" || tURL.Host == "" || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Println(getRequestId(r), " Bad target URL")
		return
	}
	targetSlice := []string{sanitizeDomain(tURL.Host), sanitizePath(tURL.Path)}
	if tURL.Port() != "" {
		targetSlice[0] += ":" + tURL.Port()
	}
	if tURL.Fragment != "" {
		targetSlice = append(targetSlice, "#", sanitizePath(tURL.Fragment))
	}

	t = strings.Join(targetSlice, "")

	if d == "" || p == "" || t == "" {
		w.WriteHeader(http.StatusBadRequest)
		logger.Println(getRequestId(r) + " rejected crowd contribution [Illegal characters] ")
		return
	}
	exists, destination, _, err := dbQuery(d, p)
	if err != nil {
		logger.Println(getRequestId(r) + " Error: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if exists {
		if t != destination {
			err = dbReport(d, p)
			if err != nil {
				logger.Println(getRequestId(r) + " Error: " + err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusCreated)
		return
	}

	err = dbInsert(d, p, t, hip)
	if err != nil {
		logger.Println(getRequestId(r) + " Error: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
