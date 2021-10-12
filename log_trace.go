package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger *log.Logger
)

//Shows server starting message
func logStart() {
	logger.Println("FastForward server")
	logger.Println("Version:", version)
	logger.Println("Server is starting...")

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
func createLogFile() {
	err := os.MkdirAll(logFile, os.ModePerm)
	if err != nil {
		panic(err)
	}

	logger = log.New(&lumberjack.Logger{
		Filename:   filepath.Join(logFile, "server.log"),
		MaxSize:    500,
		MaxBackups: 10,
		MaxAge:     28,
	}, "", log.LstdFlags)

}
