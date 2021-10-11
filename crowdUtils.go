package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
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

func crowdQueryV1(w http.ResponseWriter, r *http.Request) {
	fmt.Println("["+r.Method+"] ", r.URL.String(), "Referer", r.Referer())
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			panic(err)
		}
		d, p := r.FormValue("domain"), r.FormValue("path")
		logger.Println(d, p)
		exists, path := dbQuery(d, p)
		if exists {
			fmt.Fprint(w, path)
		} else {
			fmt.Fprint(w, "first time")
		}

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func crowdContributeV1(w http.ResponseWriter, r *http.Request) {
	fmt.Println("["+r.Method+"] ", r.URL.String(), "Referer", r.Referer())
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			panic(err)
		}

		hip := sha256(getUserIP(r))

		d, p, t := r.FormValue("domain"), r.FormValue("path"), r.FormValue("target")
		exists, _ := dbQuery(d, p)
		if exists {
			w.WriteHeader(http.StatusConflict)
		} else {
			dbInsert(d, p, t, hip)
			w.WriteHeader(http.StatusCreated)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
