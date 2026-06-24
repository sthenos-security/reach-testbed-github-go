package handlers

import (
	"log"
	"net/http"
)

func logHandlerError(operation string, err error) {
	if err != nil {
		log.Printf("%s: %v", operation, err)
	}
}

func writeBadRequest(w http.ResponseWriter, operation string, err error) {
	logHandlerError(operation, err)
	http.Error(w, "bad request", http.StatusBadRequest)
}

func writeInternalError(w http.ResponseWriter, operation string, err error) {
	logHandlerError(operation, err)
	http.Error(w, "internal error", http.StatusInternalServerError)
}
