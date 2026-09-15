package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// TestParseLanguageRejectsCraftedTag sends a crafted, oversized
// Accept-Language-style tag of the shape behind CVE-2022-32149 and asserts
// the sink rejects it before language.Parse processes it.
func TestParseLanguageRejectsCraftedTag(t *testing.T) {
	crafted := "en" + strings.Repeat("-u-nu-latn", 500)
	q := url.Values{"tag": []string{crafted}}
	req := httptest.NewRequest(http.MethodPost, "/parse-language?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	ParseLanguage(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("crafted tag: got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// TestParseLanguageAllowsValidTag asserts a legitimate allowlisted tag still
// parses and returns the same canonical result as before the dependency bump.
func TestParseLanguageAllowsValidTag(t *testing.T) {
	q := url.Values{"tag": []string{"en-US"}}
	req := httptest.NewRequest(http.MethodPost, "/parse-language?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	ParseLanguage(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("valid tag: got status %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Body.String(); got != "en-US\n" {
		t.Fatalf("valid tag: got body %q, want %q", got, "en-US\n")
	}
}
