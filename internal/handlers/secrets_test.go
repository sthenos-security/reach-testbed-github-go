package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCloudTokensUsesEnvironmentGitHubToken(t *testing.T) {
	t.Setenv(githubTokenEnvVar, "runtime-token")

	req := httptest.NewRequest(http.MethodGet, "/cloud-tokens", nil)
	rec := httptest.NewRecorder()

	CloudTokens(rec, req)

	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if got["aws_access_key_id"] != syntheticAWSAccessKeyID {
		t.Fatalf("aws_access_key_id = %q, want %q", got["aws_access_key_id"], syntheticAWSAccessKeyID)
	}

	if got["github_token"] != "runtime-token" {
		t.Fatalf("github_token = %q, want %q", got["github_token"], "runtime-token")
	}
}

func TestCloudTokensReturnsEmptyGitHubTokenWhenUnset(t *testing.T) {
	t.Setenv(githubTokenEnvVar, "")

	req := httptest.NewRequest(http.MethodGet, "/cloud-tokens", nil)
	rec := httptest.NewRecorder()

	CloudTokens(rec, req)

	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if got["github_token"] != "" {
		t.Fatalf("github_token = %q, want empty string", got["github_token"])
	}
}
