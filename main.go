package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

func getEnv() string {
	retryAfter, exists := os.LookupEnv("RETRY_AFTER")
	if exists {
		return retryAfter
	}
	return ""
}

func isValidRetryAfter(data string) bool {
	_, err := time.Parse("Mon, 02 Jan 2006 15:04:05 GMT", data)
	return err == nil
}

func getRetryAfter() (string, error) {
	envVal := getEnv()
	if isValidRetryAfter(envVal) {
		return envVal, nil
	}
	return "", errors.New("unknown Retry-After")
}

func handler(w http.ResponseWriter, r *http.Request) {
	retryAfter, err := getRetryAfter()
	if err == nil {
		w.Header().Add("Retry-After", retryAfter)
	}
	w.WriteHeader(http.StatusServiceUnavailable)
	fmt.Fprint(w, "Sorry. We're under maintenance.")
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
