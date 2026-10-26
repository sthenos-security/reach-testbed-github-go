package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFetchToolRejectsUntrustedURL(t *testing.T) {
	q := url.Values{"url": []string{"http://127.0.0.1:9/metadata"}}
	req := httptest.NewRequest(http.MethodPost, "/admin/fetch-tool?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	FetchTool(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("untrusted url: got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestFetchToolStagesLocalAllowlistedSource(t *testing.T) {
	q := url.Values{"url": []string{"https://tools.example.invalid/tool.bin"}}
	req := httptest.NewRequest(http.MethodPost, "/admin/fetch-tool?"+q.Encode(), nil)
	rec := httptest.NewRecorder()

	FetchTool(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("allowlisted url: got status %d, want %d", rec.Code, http.StatusOK)
	}
	body, err := io.ReadAll(io.LimitReader(rec.Body, 1<<20))
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	target := strings.TrimSpace(string(body))
	want := filepath.Join(os.TempDir(), "reach-testbed-tool.bin")
	if target != want {
		t.Fatalf("allowlisted url: got path %q, want %q", target, want)
	}
	staged, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read staged tool: %v", err)
	}
	if !strings.Contains(string(staged), "reach-testbed-local-tool-stub") {
		t.Fatalf("allowlisted url: staged content unexpected: %q", staged)
	}
}
