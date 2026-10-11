package handlers

import (
	"net/http"
	"os/exec"

	"github.com/reachable/reach-testbed-github-go/internal/safety"
)

func DiagnosticPing(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Query().Get("host")
	if !safety.AllowedHostname(host) {
		http.Error(w, "invalid host", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ping diagnostics are disabled\n"))
}

func SafeDiagnosticPing(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Query().Get("host")
	if !safety.AllowedHostname(host) {
		http.Error(w, "invalid host", http.StatusBadRequest)
		return
	}

	out, err := exec.Command("ping", "-c", "1", host).CombinedOutput()
	if err != nil {
		http.Error(w, string(out), http.StatusBadGateway)
		return
	}

	_, _ = w.Write(out)
}
