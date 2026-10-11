package handlers

import (
	"log"
	"net/http"
)

func writeClientError(w http.ResponseWriter, r *http.Request, status int, operation string, err error) {
	requestID := r.Header.Get("X-Request-ID")
	if !safeRequestID(requestID) {
		requestID = ""
	}
	log.Printf("%s failed (request_id=%q): %v", operation, requestID, err)

	message := "bad request"
	if status >= http.StatusInternalServerError {
		message = "internal error"
	}
	http.Error(w, message, status)
}

func safeRequestID(value string) bool {
	if len(value) > 128 {
		return false
	}
	for _, char := range value {
		if (char < 'a' || char > 'z') &&
			(char < 'A' || char > 'Z') &&
			(char < '0' || char > '9') &&
			char != '-' && char != '_' && char != '.' {
			return false
		}
	}
	return true
}
