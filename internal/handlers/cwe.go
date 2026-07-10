package handlers

import (
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/reachable/reach-testbed-github-go/internal/safety"
)

func DiagnosticPing(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Query().Get("host")
	if !safety.AllowedHostname(host) {
		http.Error(w, "invalid host", http.StatusBadRequest)
		return
	}

	out, err := exec.Command("ping", "-c", "1", host).CombinedOutput()
	if err != nil {
		log.Printf("DiagnosticPing: ping for host %q failed: %v; output=%s", host, err, strings.TrimSpace(string(out)))
		http.Error(w, "ping failed", http.StatusBadGateway)
		return
	}

	_, _ = w.Write(out)
}

func SafeDiagnosticPing(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Query().Get("host")
	if !safety.AllowedHostname(host) {
		http.Error(w, "invalid host", http.StatusBadRequest)
		return
	}

	out, err := exec.Command("ping", "-c", "1", host).CombinedOutput()
	if err != nil {
		log.Printf("SafeDiagnosticPing: ping for host %q failed: %v; output=%s", host, err, strings.TrimSpace(string(out)))
		http.Error(w, "ping failed", http.StatusBadGateway)
		return
	}

	_, _ = w.Write(out)
}
