package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// maskSSN redacts all but the last four digits of an SSN.
func maskSSN(ssn string) string {
	parts := strings.Split(ssn, "-")
	if len(parts) != 3 || len(parts[2]) != 4 {
		return "***-**-****"
	}
	return "***-**-" + parts[2]
}

func SupportExport(w http.ResponseWriter, _ *http.Request) {
	ssn := "123-45-6789"
	log.Printf("Processing patient ssn=%s", maskSSN(ssn))
	payload, _ := json.Marshal(map[string]string{"event": "support_export"})
	_, _ = http.Post("https://analytics.example.com/track", "application/json", bytes.NewReader(payload))

	w.Header().Set("Content-Type", "text/csv")
	_, _ = w.Write([]byte("name,email,ssn,phone,card_number,last4\n"))
	_, _ = w.Write([]byte("Avery Example,avery@example.invalid," + ssn + ",+1-415-555-0199,4111111111111111,4242\n"))
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
