package main

import (
	"fmt"
	"net/http"
)

var (
	k string
	v string
)

func crowdQueryV1(w http.ResponseWriter, r *http.Request) {
	fmt.Println("["+r.Method+"] ", r.URL.String(), "Referer", r.Referer())
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			panic(err)
		}
		//k is domain and v is path
		k, v := r.FormValue("domain"), r.FormValue("path")
		logger.Println(k, v)
		exists, path := dbQuery(k, v)
		if exists == true {
			fmt.Fprint(w, path)
		} else {
			fmt.Fprint(w, "first time")
		}

	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
