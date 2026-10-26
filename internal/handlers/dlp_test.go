package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSupportExportRedactsPII(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/support/export", nil)
	rec := httptest.NewRecorder()

	SupportExport(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("export: got status %d, want %d", rec.Code, http.StatusOK)
	}
	body, err := io.ReadAll(io.LimitReader(rec.Body, 1<<20))
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	text := string(body)
	for _, leak := range []string{"123-45-6789", "1978-04-23", "4111111111111111", "avery@example.invalid"} {
		if strings.Contains(text, leak) {
			t.Fatalf("export leaked PII %q in body %q", leak, text)
		}
	}
}

func TestSupportExportKeepsCSVContract(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/support/export", nil)
	rec := httptest.NewRecorder()

	SupportExport(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("export: got status %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/csv" {
		t.Fatalf("export: content-type %q, want text/csv", ct)
	}
	body, err := io.ReadAll(io.LimitReader(rec.Body, 1<<20))
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	text := string(body)
	if !strings.HasPrefix(text, "name,email,ssn,phone,card_number,last4\n") {
		t.Fatalf("export: missing CSV header: %q", text)
	}
	if !strings.Contains(text, "Avery Example") || !strings.Contains(text, "4242") {
		t.Fatalf("export: missing legitimate non-sensitive fields: %q", text)
	}
}
