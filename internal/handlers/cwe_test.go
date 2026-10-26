package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestDiagnosticPingRejectsShellMetacharacters(t *testing.T) {
	q := url.Values{"host": []string{"127.0.0.1;id"}}
	req := httptest.NewRequest(http.MethodGet, "/diagnostics/ping?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	DiagnosticPing(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("injection host: got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
	body, err := io.ReadAll(io.LimitReader(rec.Body, 1<<20))
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if strings.Contains(string(body), "uid=") {
		t.Fatalf("injection host: command output leaked: %q", body)
	}
}

func TestDiagnosticPingAllowsValidatedHost(t *testing.T) {
	q := url.Values{"host": []string{"127.0.0.1"}}
	req := httptest.NewRequest(http.MethodGet, "/diagnostics/ping?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	DiagnosticPing(rec, req)

	// Validated hosts reach the non-shell ping argv path. Ping may succeed or
	// fail depending on the local environment; either is acceptable so long as
	// the request is not rejected as invalid input.
	if rec.Code == http.StatusBadRequest {
		t.Fatalf("valid host: got status %d, want non-400", rec.Code)
	}
	if rec.Code == http.StatusBadGateway {
		body, err := io.ReadAll(io.LimitReader(rec.Body, 1<<20))
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if text := string(body); strings.Contains(text, "ping:") || strings.Contains(text, "Error") {
			t.Fatalf("valid host: client saw internal diagnostic detail: %q", text)
		}
	}
}
