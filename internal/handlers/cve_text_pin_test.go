package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGolangOrgXTextPinnedAtLeast038(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	modPath := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "go.mod"))
	data, err := os.ReadFile(modPath)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !strings.Contains(string(data), "golang.org/x/text v0.3.8") {
		t.Fatalf("go.mod missing pinned golang.org/x/text v0.3.8:\n%s", data)
	}

	q := url.Values{}
	q.Set("tag", "en-US")
	req := httptest.NewRequest(http.MethodPost, "/parse-language?"+q.Encode(), nil)
	rr := httptest.NewRecorder()
	ParseLanguage(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("ParseLanguage status = %d, want %d; body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if strings.TrimSpace(rr.Body.String()) == "" {
		t.Fatal("ParseLanguage returned empty body")
	}
}
