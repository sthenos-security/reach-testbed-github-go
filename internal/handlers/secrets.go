package handlers

import (
	"encoding/json"
	"net/http"
	"os"
)

func ServiceToken(w http.ResponseWriter, r *http.Request) {
	EnvToken(w, r)
}

func CloudTokens(w http.ResponseWriter, _ *http.Request) {
	_ = json.NewEncoder(w).Encode(map[string]string{
		"aws_access_key_id_configured": configured(os.Getenv("AWS_ACCESS_KEY_ID")),
		"github_token_configured":      configured(os.Getenv("GITHUB_TOKEN")),
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

func configured(value string) string {
	if value == "" {
		return "false"
	}
	return "true"
}
