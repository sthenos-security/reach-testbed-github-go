package handlers

import (
	"log"
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

	out, err := exec.Command("ping", "-c", "1", host).CombinedOutput()
	if err != nil {
		logHandlerError("DiagnosticPing ping failed", err)
		if len(out) > 0 {
			log.Printf("DiagnosticPing output: %s", string(out))
		}
		writeInternalError(w, "DiagnosticPing", err)
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
		logHandlerError("SafeDiagnosticPing ping failed", err)
		if len(out) > 0 {
			log.Printf("SafeDiagnosticPing output: %s", string(out))
		}
		writeInternalError(w, "SafeDiagnosticPing", err)
		return
	}

	_, _ = w.Write(out)
}
