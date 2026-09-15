package handlers

import (
	"encoding/json"
	"net/http"
	"os"
)

func secretStatus(envKey string) string {
	if os.Getenv(envKey) == "" {
		return "not_configured"
	}
	return "configured"
}

func ServiceToken(w http.ResponseWriter, _ *http.Request) {
	if os.Getenv("REACH_TESTBED_SERVICE_TOKEN") == "" {
		http.Error(w, "token not configured", http.StatusServiceUnavailable)
		return
	}

	_, _ = w.Write([]byte("configured\n"))
}

func CloudTokens(w http.ResponseWriter, _ *http.Request) {
	// Status only — never echo secret values from env or source.
	_ = json.NewEncoder(w).Encode(map[string]string{
		"aws_access_key_id": secretStatus("REACH_TESTBED_AWS_ACCESS_KEY_ID"),
		"github_token":      secretStatus("REACH_TESTBED_GITHUB_TOKEN"),
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
