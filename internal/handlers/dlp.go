package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func SupportExport(w http.ResponseWriter, _ *http.Request) {
	log.Printf("Processing support export request")

	w.Header().Set("Content-Type", "text/csv")
	_, _ = w.Write([]byte("name,email,ssn,phone,card_number,last4\n"))
	_, _ = w.Write([]byte("Avery Example,redacted@example.invalid,***-**-****,+1-***-***-****,****,****\n"))
}

func SupportProfile(w http.ResponseWriter, _ *http.Request) {
	// Values are redacted to avoid emitting PII in response bodies.
	_ = json.NewEncoder(w).Encode(map[string]string{
		"name":            "Jordan Example",
		"email":           "redacted@example.invalid",
		"date_of_birth":   "redacted",
		"tax_identifier":  "redacted",
		"routing_number":  "redacted",
		"account_number":  "redacted",
		"passport_number": "redacted",
	})
}
