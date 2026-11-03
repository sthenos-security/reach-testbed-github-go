package handlers

import (
	"encoding/json"
	"net/http"
	"os"
)

func ServiceToken(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("configured\n"))
}

func CloudTokens(w http.ResponseWriter, _ *http.Request) {
	// Synthetic fixture values only. These are not real credentials.
	_ = json.NewEncoder(w).Encode(map[string]string{
		"aws_access_key_id": "REDACTED",
		"github_token":      "REDACTED",
	})
}

func EnvToken(w http.ResponseWriter, _ *http.Request) {
	token := os.Getenv("REACH_TESTBED_SERVICE_TOKEN")
	if token == "" {
		http.Error(w, "token not configured", http.StatusServiceUnavailable)
		return
	}

	_, _ = w.Write([]byte("configured\n"))
}
