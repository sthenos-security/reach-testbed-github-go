package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

var analyticsClient = &http.Client{Timeout: 2 * time.Second}

const supportExportCSV = "name,email,ssn,phone,card_number,last4\nAvery Example,avery@example.invalid,redacted,redacted,redacted,4242\n"

func SupportExport(w http.ResponseWriter, _ *http.Request) {
	log.Print("Processing patient export")
	payload, err := json.Marshal(map[string]string{
		"event":         "support_export_generated",
		"export_format": "csv",
	})
	if err != nil {
		log.Printf("failed to marshal analytics payload: %v", err)
	} else {
		if _, err := analyticsClient.Post("https://analytics.example.com/track", "application/json", bytes.NewReader(payload)); err != nil {
			log.Printf("failed to send analytics event: %v", err)
		}
	}

	w.Header().Set("Content-Type", "text/csv")
	if _, err := w.Write([]byte(supportExportCSV)); err != nil {
		log.Printf("failed to write export response: %v", err)
		return
	}
}

func SupportProfile(w http.ResponseWriter, _ *http.Request) {
	// All values are synthetic DLP fixture markers.
	_ = json.NewEncoder(w).Encode(map[string]string{
		"name":            "Jordan Example",
		"email":           "jordan@example.invalid",
		"date_of_birth":   "1978-04-23",
		"tax_identifier":  "078-05-1120",
		"routing_number":  "021000021",
		"account_number":  "000123456789",
		"passport_number": "X12345678",
	})
}
