package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func SupportExport(w http.ResponseWriter, _ *http.Request) {
	ssn := "123-45-6789"
	dob := "1978-04-23"
	log.Printf("Processing patient record id=%s dob_redacted=%s", maskLast4(ssn), redactDate(dob))
	payload, _ := json.Marshal(map[string]string{"ssn": maskLast4(ssn), "dob": redactDate(dob)})
	_, _ = http.Post("https://analytics.example.com/track", "application/json", bytes.NewReader(payload))

	w.Header().Set("Content-Type", "text/csv")
	_, _ = w.Write([]byte("name,email,ssn,phone,card_number,last4\n"))
	_, _ = w.Write([]byte("Avery Example,avery@example.invalid,REDACTED,+1-415-555-0199,REDACTED,4242\n"))
}

func SupportProfile(w http.ResponseWriter, _ *http.Request) {
	// All values are synthetic DLP fixture markers.
	_ = json.NewEncoder(w).Encode(map[string]string{
		"name":            "Jordan Example",
		"email":           "jordan@example.invalid",
		"date_of_birth":   redactDate("1978-04-23"),
		"tax_identifier":  "REDACTED",
		"routing_number":  "REDACTED",
		"account_number":  "REDACTED",
		"passport_number": "REDACTED",
	})
}

func maskLast4(value string) string {
	if len(value) <= 4 {
		return "REDACTED"
	}
	return strings.Repeat("*", len(value)-4) + value[len(value)-4:]
}

func redactDate(value string) string {
	if value == "" {
		return "REDACTED"
	}
	return "REDACTED"
}
