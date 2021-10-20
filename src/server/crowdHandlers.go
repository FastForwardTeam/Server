package main

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"net/http"
)

var (
	k string
	v string
)

func sha256(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
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

	logger.Println("["+r.Method+"] ", r.URL.String(), "Referer", r.Referer())

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	/* Randomly ask(force) user to verify
	n := rand.Intn(10)
	if n == 1 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	*/
	err := r.ParseForm()
	if err != nil {
		logger.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	d, p := r.FormValue("domain"), r.FormValue("path")
	logger.Println(d, p)
	exists, path := dbQuery(d, p)
	if exists {
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

	// Return 201 anyway
	exists, path := dbQuery(d, p)
	if exists {
		if p != path {
			dbReport(d, p)
		}
		w.WriteHeader(http.StatusCreated)
		return
	}

	dbInsert(d, p, t, hip)
	w.WriteHeader(http.StatusCreated)
}
