package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeAIAnswerInvalidJSONReturnsGenericError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/ai/safe-answer", strings.NewReader("{"))
	rec := httptest.NewRecorder()

	SafeAIAnswer(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if got := rec.Body.String(); got != "invalid request body\n" {
		t.Fatalf("expected generic error, got %q", got)
	}
}

func TestFetchToolRejectsNonHTTPSURL(t *testing.T) {
	t.Setenv("REACH_FETCH_TOOL_ALLOWED_URLS", "https://example.invalid/tool")
	req := httptest.NewRequest(http.MethodPost, "/admin/fetch-tool?url=http://example.invalid/tool", nil)
	rec := httptest.NewRecorder()

	FetchTool(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if got := rec.Body.String(); got != "invalid source url\n" {
		t.Fatalf("expected invalid source response, got %q", got)
	}
}

func TestFetchToolRejectsDisallowedHost(t *testing.T) {
	t.Setenv("REACH_FETCH_TOOL_ALLOWED_URLS", "https://example.invalid/tool")
	req := httptest.NewRequest(http.MethodPost, "/admin/fetch-tool?url=https://not-allowed.invalid/tool", nil)
	rec := httptest.NewRecorder()

	FetchTool(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if got := rec.Body.String(); got != "invalid source url\n" {
		t.Fatalf("expected invalid source response, got %q", got)
	}
}

func TestFetchToolRejectsNonAllowlistedPath(t *testing.T) {
	t.Setenv("REACH_FETCH_TOOL_ALLOWED_URLS", "https://example.invalid/tool")
	req := httptest.NewRequest(http.MethodPost, "/admin/fetch-tool?url=https://example.invalid/tool/subpath", nil)
	rec := httptest.NewRecorder()

	FetchTool(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if got := rec.Body.String(); got != "invalid source url\n" {
		t.Fatalf("expected invalid source response, got %q", got)
	}
}

func TestFetchToolFetchFailureReturnsGenericError(t *testing.T) {
	t.Setenv("REACH_FETCH_TOOL_ALLOWED_URLS", "https://127.0.0.1:1/tool")
	req := httptest.NewRequest(http.MethodPost, "/admin/fetch-tool?url=https://127.0.0.1:1/tool", nil)
	rec := httptest.NewRecorder()

	FetchTool(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, rec.Code)
	}
	body := rec.Body.String()
	if body != "unable to fetch tool\n" {
		t.Fatalf("expected generic fetch error, got %q", body)
	}
	if strings.Contains(strings.ToLower(body), "dial tcp") {
		t.Fatalf("response leaked internal details: %q", body)
	}
}

func TestFetchToolAllowsApprovedHTTPSHost(t *testing.T) {
	target := filepath.Join(os.TempDir(), "reach-testbed-tool.bin")
	_ = os.Remove(target)
	t.Cleanup(func() { _ = os.Remove(target) })

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("tool-bytes"))
	}))
	defer server.Close()

	t.Setenv("REACH_FETCH_TOOL_ALLOWED_URLS", server.URL+"/tool")
	originalTransport := http.DefaultTransport
	http.DefaultTransport = server.Client().Transport
	t.Cleanup(func() { http.DefaultTransport = originalTransport })

	req := httptest.NewRequest(http.MethodPost, "/admin/fetch-tool?url="+server.URL+"/tool", nil)
	rec := httptest.NewRecorder()
	FetchTool(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if rec.Body.String() != target+"\n" {
		t.Fatalf("expected response to include target path %q, got %q", target+"\n", rec.Body.String())
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed reading target file: %v", err)
	}
	if string(data) != "tool-bytes" {
		t.Fatalf("expected persisted payload %q, got %q", "tool-bytes", string(data))
	}
}

func TestParseLanguageRejectsUnsupportedTag(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/parse-language?tag=xx-INVALID", nil)
	rec := httptest.NewRecorder()

	ParseLanguage(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "unsupported language tag\n" {
		t.Fatalf("unexpected response body: %q", string(body))
	}
}
