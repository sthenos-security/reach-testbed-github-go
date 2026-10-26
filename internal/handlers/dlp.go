package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func SupportExport(w http.ResponseWriter, _ *http.Request) {
	// Local audit only — no raw PII constants, logs, or outbound analytics POST.
	log.Printf("Processing patient export request")

	w.Header().Set("Content-Type", "text/csv")
	_, _ = w.Write([]byte(redactedSupportExportCSV()))
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

func redactedSupportExportCSV() string {
	var b strings.Builder
	b.WriteString("name,email,ssn,phone,card_number,last4\n")
	b.WriteString("Avery Example,[REDACTED],[REDACTED],[REDACTED],[REDACTED],4242\n")
	return b.String()
}
