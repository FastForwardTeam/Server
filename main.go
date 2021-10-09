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
)

func bypassed(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Query())
	w.Header().Set("Location", r.URL.Query().Get("target"))
	w.Header().Set("Referer", r.URL.Query().Get("referer"))
	w.WriteHeader(http.StatusPermanentRedirect)
}

func all(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

//Request will be application/x-www-form-urlencoded

func crowdQueryV1(w http.ResponseWriter, r *http.Request) {
	fmt.Println("["+r.Method+"] ", r.URL.String(), "Referer", r.Referer())
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			panic(err)
		}
		for k, v := range r.Form {
			// domain, path
			fmt.Printf("\t %s = %s\n", k, v)
		}
		// TODO fetch target from db
		// test example: that was read from db
		fmt.Fprint(w, dbQuery())
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
		for k, v := range r.Form {
			// TODO upsert to db
			// path, domain, target
			fmt.Printf("\t %s = %s\n", k, v)
		}
		fmt.Println()
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	parseEnv()
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)

	logger.Println("FastForward server")
	logger.Println("Version:", Version)
	logger.Println("Server is starting...")

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

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					requestID = "unknown"
				}
				logger.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
