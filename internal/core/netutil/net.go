package netutil

import (
	"crypto/tls"
	"fmt"
	"net"
	"regexp"
	"strings"

	"golang.org/x/net/idna"
)

// hostPattern validates domain names, localhost, IPv4 addresses,
// and optional CIDR notation for IPv4.
//
// This regex is used as a secondary validation step after IDN
// normalization to ensure the hostname conforms to expected
// network formats.
var hostPattern = regexp.MustCompile(
	`^(localhost|(([a-zA-Z0-9-]{1,63}\.)+[a-zA-Z]{2,63})|((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}(\/(3[0-2]|[12]?[0-9]))?))$`,
)

// --------------------
// Public API
// --------------------

// NormalizeHostWithSuffix normalizes the hostname portion of a URL-like
// input while preserving any path or query suffix.
//
// Examples:
//
//	example.com
//	example.com/path
//	https://example.com:443/index.html
//
// The function performs:
//
//   - scheme stripping (http/https)
//   - host extraction
//   - IDN normalization (unicode → punycode)
//   - hostname validation
//
// The suffix (path/query) is preserved exactly as provided.
func NormalizeHostWithSuffix(input string) (string, error) {
	host, suffix, err := extractHostAndSuffix(input)
	if err != nil {
		return "", err
	}

	normalized, err := normalizeHost(host)
	if err != nil {
		return "", err
	}

	return normalized + suffix, nil
}

// ExtractTLSServerName extracts and normalizes the hostname used for
// TLS Server Name Indication (SNI).
//
// This removes any scheme, port, path, or query components
// and returns only the validated hostname.
//
// Example:
//
//	https://example.com:443/path → example.com
func ExtractTLSServerName(input string) (string, error) {
	host, _, err := extractHostAndSuffix(input)
	if err != nil {
		return "", err
	}

	return normalizeHost(host)
}

// ProtocolToScheme converts a protocol string into its URL scheme.
//
// Examples:
//
//	"https"    → "https://"
//	"http"     → "http://"
//	"https://" → "https://"
//	"http://"  → "http://"
func ProtocolToScheme(protocol string) string {
	if IsHTTPS(protocol) {
		return "https://"
	}
	return "http://"
}

// IsHTTPS returns true if the provided protocol string represents HTTPS.
//
// The comparison is case-insensitive and accepts both:
//
//	"https"
//	"https://"
func IsHTTPS(protocol string) bool {
	return strings.EqualFold(protocol, "https") ||
		strings.EqualFold(protocol, "https://")
}

// ParsePortOrDefault validates a port number and returns it as uint16.
//
// If the port is outside the valid TCP range (0–65535),
// the provided defaultPort is returned instead.
func ParsePortOrDefault(port int, defaultPort uint16) uint16 {
	if port < 0 || port > 65535 {
		return defaultPort
	}
	return uint16(port)
}

// --------------------
// Internal Helpers
// --------------------

// extractHostAndSuffix separates a hostname from any trailing
// path or query string.
//
// Example:
//
//	input:  https://example.com:443/path?a=1
//	host:   example.com
//	suffix: /path?a=1
//
// The function also strips URL schemes and ports.
func extractHostAndSuffix(input string) (host string, suffix string, err error) {
	input = strings.TrimSpace(input)
	input = stripScheme(input)

	// Separate suffix (/path or ?query)
	if i := strings.IndexAny(input, "/?"); i != -1 {
		suffix = input[i:]
		input = input[:i]
	}

	// Attempt host:port parsing
	if h, _, splitErr := net.SplitHostPort(input); splitErr == nil {
		host = h
	} else {
		host = input
	}

	if host == "" {
		return "", "", fmt.Errorf("empty host")
	}

	return host, suffix, nil
}

// normalizeHost converts and validates a hostname.
//
// Steps:
//
//  1. Detects if the host is an IP address.
//  2. Converts internationalized domain names (IDN) to ASCII (punycode).
//  3. Validates the hostname against hostPattern.
//
// This ensures safe usage in networking and TLS contexts.
func normalizeHost(host string) (string, error) {
	// If host is already an IP address, return as-is.
	if ip := net.ParseIP(host); ip != nil {
		return host, nil
	}

	// Convert IDN → ASCII (punycode)
	ascii, err := idna.ToASCII(host)
	if err != nil {
		return "", fmt.Errorf("invalid host: %s", host)
	}

	// Validate hostname structure
	if !hostPattern.MatchString(ascii) {
		return "", fmt.Errorf("invalid host: %s", host)
	}

	return ascii, nil
}

// stripScheme removes common HTTP/HTTPS URL schemes from an input string.
func stripScheme(input string) string {
	input = strings.TrimPrefix(input, "https://")
	input = strings.TrimPrefix(input, "http://")
	return input
}

// ParseTLSVersion converts a TLS version string into the corresponding
// crypto/tls constant.
//
// Supported formats:
//
//	"tls1.0", "1.0"
//	"tls1.1", "1.1"
//	"tls1.2", "1.2"
//	"tls1.3", "1.3"
func ParseTLSVersion(v string) (uint16, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "tls1.0", "1.0":
		return tls.VersionTLS10, nil
	case "tls1.1", "1.1":
		return tls.VersionTLS11, nil
	case "tls1.2", "1.2":
		return tls.VersionTLS12, nil
	case "tls1.3", "1.3":
		return tls.VersionTLS13, nil
	default:
		return 0, fmt.Errorf("unsupported tls version: %s", v)
	}
}

// IsPortAvailable checks whether a TCP port is currently available
// on the local machine.
//
// This attempts to bind to the port on 127.0.0.1.
// If the bind succeeds, the port is considered available.
func IsPortAvailable(port int) bool {
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}

	_ = ln.Close()
	return true
}
