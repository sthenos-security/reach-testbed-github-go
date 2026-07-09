package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func SupportExport(w http.ResponseWriter, _ *http.Request) {
	log.Printf("SupportExport requested; exporting redacted synthetic data")

	w.Header().Set("Content-Type", "text/csv")
	_, _ = w.Write([]byte("name,email,ssn,phone,card_number,last4\n"))
	_, _ = w.Write([]byte("Avery Example,redacted@example.invalid,***-**-****,***-***-****,****-****-****-1111,4242\n"))
}

func SupportProfile(w http.ResponseWriter, _ *http.Request) {
	// All values are synthetic DLP fixture markers.
	_ = json.NewEncoder(w).Encode(map[string]string{
		"name":            "redacted",
		"email":           "redacted@example.invalid",
		"date_of_birth":   "redacted",
		"tax_identifier":  "redacted",
		"routing_number":  "redacted",
		"account_number":  "redacted",
		"passport_number": "redacted",
	})
}
