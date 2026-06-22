package handlers

import (
	"encoding/json"
	"net/http"
	"os"
)

const syntheticServiceToken = "rtg_live_synthetic_token_1234567890"
const syntheticAWSAccessKeyID = "AKIAIOSFODNN7EXAMPLE"

const envGitHubToken = "REACH_TESTBED_GITHUB_TOKEN"
const envServiceToken = "REACH_TESTBED_SERVICE_TOKEN"

func ServiceToken(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte(syntheticServiceToken + "\n"))
}

func CloudTokens(w http.ResponseWriter, _ *http.Request) {
	githubToken := os.Getenv(envGitHubToken)
	if githubToken == "" {
		http.Error(w, "github token not configured", http.StatusServiceUnavailable)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{
		"aws_access_key_id": syntheticAWSAccessKeyID,
		"github_token":      githubToken,
	})
}

func EnvToken(w http.ResponseWriter, _ *http.Request) {
	token := os.Getenv(envServiceToken)
	if token == "" {
		http.Error(w, "token not configured", http.StatusServiceUnavailable)
		return
	}

	_, _ = w.Write([]byte("configured\n"))
}
