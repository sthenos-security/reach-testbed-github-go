package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSupportExportRedactsPIIAndSkipsOutbound(t *testing.T) {
	prev := http.DefaultTransport
	http.DefaultTransport = rejectOutboundTransport{}
	t.Cleanup(func() { http.DefaultTransport = prev })

	t.Run("attack_style_pii_absent", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/support/export", nil)
		rr := httptest.NewRecorder()
		SupportExport(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		body, err := io.ReadAll(io.LimitReader(rr.Body, 1<<20))
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		text := string(body)
		for _, needle := range []string{"123-45-6789", "1978-04-23", "avery@example.invalid", "4111111111111111", "+1-415-555-0199"} {
			if strings.Contains(text, needle) {
				t.Fatalf("response still contains PII %q: %s", needle, text)
			}
		}
	})

	t.Run("legitimate_export_shape", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/support/export", nil)
		rr := httptest.NewRecorder()
		SupportExport(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		body, err := io.ReadAll(io.LimitReader(rr.Body, 1<<20))
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		text := string(body)
		if !strings.Contains(text, "name,email,ssn,phone,card_number,last4") {
			t.Fatalf("missing CSV header: %s", text)
		}
		if !strings.Contains(text, "Avery Example") || !strings.Contains(text, "4242") {
			t.Fatalf("missing redacted legitimate row: %s", text)
		}
		if !strings.Contains(text, "[REDACTED]") {
			t.Fatalf("expected redacted markers: %s", text)
		}
	})
}
