package safety

import (
	"regexp"
	"strings"
)

var hostPattern = regexp.MustCompile(`^[a-zA-Z0-9.-]{1,253}$`)

func AllowedHostname(host string) bool {
	if len(host) == 0 || len(host) > 253 {
		return false
	}
	if strings.HasPrefix(host, "-") || strings.HasSuffix(host, "-") || strings.HasPrefix(host, ".") || strings.HasSuffix(host, ".") {
		return false
	}
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
