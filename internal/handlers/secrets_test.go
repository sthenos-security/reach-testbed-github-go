package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCloudTokensDoesNotExposeConfiguredGitHubToken(t *testing.T) {
	const token = "configured-test-token"
	t.Setenv("REACH_TESTBED_GITHUB_TOKEN", token)

	response := httptest.NewRecorder()
	CloudTokens(response, httptest.NewRequest("GET", "/cloud-tokens", nil))

	if response.Code != 200 {
		t.Fatalf("CloudTokens() status = %d, want 200", response.Code)
	}
	if strings.Contains(response.Body.String(), token) {
		t.Fatal("CloudTokens() exposed the configured GitHub token")
	}
	if !strings.Contains(response.Body.String(), `"github_token":"configured"`) {
		t.Fatalf("CloudTokens() response = %s, want configured status", response.Body.String())
	}
}

func TestCloudTokensReportsGitHubTokenNotConfigured(t *testing.T) {
	t.Setenv("REACH_TESTBED_GITHUB_TOKEN", "")

	response := httptest.NewRecorder()
	CloudTokens(response, httptest.NewRequest("GET", "/cloud-tokens", nil))

	if response.Code != 200 {
		t.Fatalf("CloudTokens() status = %d, want 200", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"github_token":"not configured"`) {
		t.Fatalf("CloudTokens() response = %s, want not-configured status", response.Body.String())
	}
}
