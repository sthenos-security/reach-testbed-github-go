package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestDiagnosticPingRejectsCommandInjection(t *testing.T) {
	query := url.Values{"host": {"127.0.0.1; touch /tmp/owned"}}
	req := httptest.NewRequest(http.MethodGet, "/diagnostics/ping?"+query.Encode(), nil)
	rec := httptest.NewRecorder()

	DiagnosticPing(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if got := rec.Body.String(); got != "invalid host\n" {
		t.Fatalf("body = %q, want %q", got, "invalid host\n")
	}
}

func TestDiagnosticPingReturnsStatusForValidHost(t *testing.T) {
	query := url.Values{"host": {"example.com"}}
	req := httptest.NewRequest(http.MethodGet, "/diagnostics/ping?"+query.Encode(), nil)
	rec := httptest.NewRecorder()

	DiagnosticPing(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Body.String(); got != "ping diagnostics are disabled\n" {
		t.Fatalf("body = %q, want %q", got, "ping diagnostics are disabled\n")
	}
}
