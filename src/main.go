/*
Copyright 2021, 2022 NotAProton, mockuser404

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
	"context"
	"fmt"
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
	version string = "1.4.5"
)

func bypassed(w http.ResponseWriter, r *http.Request) {
	// redirect to error page
	http.Redirect(w, r, "https://fastforward.team/crowd-bypassed", http.StatusSeeOther)
}

func all(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}

}

func main() {

	parseEnv()

	logStart()
	err := connectDb()
	if err != nil {
		logger.Fatalln(err)
	}

	//check connection to db
	err = db.Ping()
	if err != nil {
		logger.Fatalln(err)
	}
	logger.Println("Connected to database")

	RSAprivateKey, RSApublicKey, err = loadRSAKeys()
	if err != nil {
		logger.Fatalln(err)
	}
	logger.Println("Loaded RSA keys")

	router := http.NewServeMux()
	router.HandleFunc("/", all)
	router.HandleFunc("/bypassed", bypassed)
	router.HandleFunc("/navigate", bypassed)
	router.HandleFunc("/crowd-bypassed", bypassed)
	router.HandleFunc("/crowd/query_v1", crowdQueryV1)
	router.HandleFunc("/crowd/contribute_v1", crowdContributeV1)

	adminPanelRouters(router)

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
			logger.Fatalln(err)
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
