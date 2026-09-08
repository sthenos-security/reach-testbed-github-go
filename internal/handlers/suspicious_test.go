package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFetchToolAllowlistAndLocalAdapter(t *testing.T) {
	prev := http.DefaultTransport
	http.DefaultTransport = rejectOutboundTransport{}
	t.Cleanup(func() { http.DefaultTransport = prev })

	t.Run("attack_url_rejected", func(t *testing.T) {
		q := url.Values{}
		q.Set("url", "http://169.254.169.254/latest/meta-data/")
		req := httptest.NewRequest(http.MethodPost, "/admin/fetch-tool?"+q.Encode(), nil)
		rr := httptest.NewRecorder()
		FetchTool(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusBadRequest, rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), "url not allowed") {
			t.Fatalf("body = %q, want url not allowed", rr.Body.String())
		}
	})

	t.Run("legitimate_allowlisted_local_stage", func(t *testing.T) {
		q := url.Values{}
		q.Set("url", "https://example.invalid/tool")
		req := httptest.NewRequest(http.MethodPost, "/admin/fetch-tool?"+q.Encode(), nil)
		rr := httptest.NewRecorder()
		FetchTool(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%s", rr.Code, http.StatusOK, rr.Body.String())
		}
		target := strings.TrimSpace(rr.Body.String())
		want := filepath.Join(os.TempDir(), "reach-testbed-tool.bin")
		if target != want {
			t.Fatalf("target = %q, want %q", target, want)
		}
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read staged tool: %v", err)
		}
		if !strings.Contains(string(data), "local-safe-tool:example-tool") {
			t.Fatalf("staged content = %q", data)
		}
	})
}
