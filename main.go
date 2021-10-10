/////////////////////////////////////////////////////////////////////////
/*
/*
/*
/*
/*
*/ /////////////////////////////////////////////////////////////////////////

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"time"
)

type key int

const (
	requestIDKey key = 0
)

var (
	Version string = "0.0.1"
	logger         = log.New(os.Stdout, "http: ", log.LstdFlags)
)

func bypassed(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Query())
	w.Header().Set("Location", r.URL.Query().Get("target"))
	w.Header().Set("Referer", r.URL.Query().Get("referer"))
	w.WriteHeader(http.StatusPermanentRedirect)
}

func all(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}

}

//Request will be application/x-www-form-urlencoded

func crowdContributeV1(w http.ResponseWriter, r *http.Request) {
	fmt.Println("["+r.Method+"] ", r.URL.String(), "Referer", r.Referer())
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			panic(err)
		}
		for k, v := range r.Form {
			// TODO upsert to db
			// path, domain, target
			fmt.Printf("\t %s = %s\n", k, v)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	parseEnv()

	logStart()
	connectDb()

	//check connection
	err := db.Ping()
	if err != nil {
		panic(err.Error())
	}
	logger.Println("Connected to database")

	router := http.NewServeMux()
	router.HandleFunc("/", all)
	router.HandleFunc("/firstrun", all)
	router.HandleFunc("/options", all)
	router.HandleFunc("/bypassed", bypassed)
	router.HandleFunc("/navigate", bypassed)
	router.HandleFunc("/crowd-bypassed", bypassed)
	router.HandleFunc("/crowd/query_v1", crowdQueryV1)
	router.HandleFunc("/crowd/contribute_v1", crowdContributeV1)
	router.Handle("/healthz", healthz())

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	server := &http.Server{
		Addr:         port,
		Handler:      tracing(nextRequestID)(logging(logger)(router)),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Println("Server is shutting down...")
		atomic.StoreInt32(&healthy, 0)

		err := db.Close()
		if err != nil {
			panic(err.Error())
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	logger.Println("Server is ready to handle requests at", port)
	atomic.StoreInt32(&healthy, 1)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", port, err)
	}

	<-done
	logger.Println("Server stopped")
}
