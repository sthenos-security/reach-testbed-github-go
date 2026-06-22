package safety

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

var hostPattern = regexp.MustCompile(`^[a-zA-Z0-9.-]{1,253}$`)

// privateRanges lists CIDR blocks that must not be reachable via outbound fetches.
var privateRanges []*net.IPNet

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	} {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			panic("safety: invalid built-in CIDR " + cidr + ": " + err.Error())
		}
		privateRanges = append(privateRanges, network)
	}
}

func AllowedHostname(host string) bool {
	return hostPattern.MatchString(host)
}

func AllowedLanguageTag(tag string) bool {
	switch tag {
	case "en", "en-US", "fr", "fr-FR", "es", "es-ES":
		return true
	default:
		return false
	}
}

// SafeFetchURL parses rawURL and returns it only when the scheme is "https"
// and the host does not resolve to a private or internal address.
// It performs DNS resolution to guard against DNS-rebinding attacks.
func SafeFetchURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL")
	}
	if u.Scheme != "https" {
		return nil, fmt.Errorf("only HTTPS URLs are permitted")
	}
	host := u.Hostname()
	if err := assertSafeHost(host); err != nil {
		return nil, err
	}
	return u, nil
}

// assertSafeHost rejects hostnames that are, or resolve to, private/internal addresses.
func assertSafeHost(host string) error {
	lower := strings.ToLower(host)
	if lower == "localhost" || strings.HasSuffix(lower, ".local") || strings.HasSuffix(lower, ".internal") {
		return fmt.Errorf("requests to internal or private addresses are not permitted")
	}

	// If the host is a bare IP, check it directly.
	if ip := net.ParseIP(host); ip != nil {
		if isPrivateIP(ip) {
			return fmt.Errorf("requests to internal or private addresses are not permitted")
		}
		return nil
	}

	// Resolve the hostname and check every returned address to prevent DNS-rebinding.
	addrs, err := net.LookupHost(host)
	if err != nil {
		return fmt.Errorf("unable to resolve host")
	}
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip != nil && isPrivateIP(ip) {
			return fmt.Errorf("requests to internal or private addresses are not permitted")
		}
	}
	return nil
}

// isPrivateIP returns true when ip falls within a private, loopback, or link-local range.
func isPrivateIP(ip net.IP) bool {
	for _, network := range privateRanges {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}
