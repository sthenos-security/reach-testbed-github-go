package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
)

func SupportExport(w http.ResponseWriter, _ *http.Request) {
	log.Printf("Processing patient record export")

	w.Header().Set("Content-Type", "text/csv")
	_, _ = w.Write([]byte("name,email_hash,record_id\n"))
	emailHash := sha256.Sum256([]byte("avery@example.invalid"))
	_, _ = w.Write([]byte("Avery Example," + hex.EncodeToString(emailHash[:]) + ",patient-123\n"))
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
