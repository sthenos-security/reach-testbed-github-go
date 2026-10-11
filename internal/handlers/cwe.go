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
	if !safety.AllowedHostname(host) || strings.HasPrefix(host, "-") {
		http.Error(w, "invalid host", http.StatusBadRequest)
		return
	}

	out, err := exec.Command("ping", "-c", "1", host).CombinedOutput()
	if err != nil {
		log.Printf("diagnostic ping failed request_id=%q error=%q", r.Header.Get("X-Request-ID"), err.Error())
		http.Error(w, "internal error", http.StatusBadGateway)
		return
	}

	_, _ = w.Write(out)
}

func SafeDiagnosticPing(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Query().Get("host")
	if !safety.AllowedHostname(host) || strings.HasPrefix(host, "-") {
		http.Error(w, "invalid host", http.StatusBadRequest)
		return
	}

	out, err := exec.Command("ping", "-c", "1", host).CombinedOutput()
	if err != nil {
		log.Printf("safe diagnostic ping failed request_id=%q error=%q", r.Header.Get("X-Request-ID"), err.Error())
		http.Error(w, "internal error", http.StatusBadGateway)
		return
	}

	_, _ = w.Write(out)
}
