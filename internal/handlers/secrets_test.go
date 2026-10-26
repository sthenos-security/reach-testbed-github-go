package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestServiceTokenNeverEchoesSecret(t *testing.T) {
	t.Setenv("REACH_TESTBED_SERVICE_TOKEN", "runtime-test-value")
	req := httptest.NewRequest(http.MethodGet, "/token", nil)
	rec := httptest.NewRecorder()

	ServiceToken(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("configured token: got status %d, want %d", rec.Code, http.StatusOK)
	}
	body, err := io.ReadAll(io.LimitReader(rec.Body, 1<<20))
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	text := string(body)
	if strings.Contains(text, "runtime-test-value") {
		t.Fatalf("configured token: secret echoed in response: %q", text)
	}
	if text != "configured\n" {
		t.Fatalf("configured token: got body %q, want configured status", text)
	}
}

func TestCloudTokensNeverEchoesSecrets(t *testing.T) {
	t.Setenv("REACH_TESTBED_AWS_ACCESS_KEY_ID", "runtime-test-value")
	t.Setenv("REACH_TESTBED_GITHUB_TOKEN", "runtime-test-value")
	req := httptest.NewRequest(http.MethodGet, "/cloud-tokens", nil)
	rec := httptest.NewRecorder()

	CloudTokens(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("cloud tokens: got status %d, want %d", rec.Code, http.StatusOK)
	}
	body, err := io.ReadAll(io.LimitReader(rec.Body, 1<<20))
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	text := string(body)
	for _, leak := range []string{"runtime-test-value", "ghp_", "AKIA"} {
		if strings.Contains(text, leak) {
			t.Fatalf("cloud tokens: leaked %q in body %q", leak, text)
		}
	}
	var got map[string]string
	if err := json.NewDecoder(strings.NewReader(text)).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got["aws_access_key_id"] != "configured" || got["github_token"] != "configured" {
		t.Fatalf("cloud tokens: unexpected status payload: %#v", got)
	}
}

func TestCloudTokensReportsNotConfigured(t *testing.T) {
	_ = os.Unsetenv("REACH_TESTBED_AWS_ACCESS_KEY_ID")
	_ = os.Unsetenv("REACH_TESTBED_GITHUB_TOKEN")
	req := httptest.NewRequest(http.MethodGet, "/cloud-tokens", nil)
	rec := httptest.NewRecorder()

	CloudTokens(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("cloud tokens: got status %d, want %d", rec.Code, http.StatusOK)
	}
	var got map[string]string
	if err := json.NewDecoder(io.LimitReader(rec.Body, 1<<20)).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got["aws_access_key_id"] != "not_configured" || got["github_token"] != "not_configured" {
		t.Fatalf("cloud tokens: unexpected status payload: %#v", got)
	}
}
