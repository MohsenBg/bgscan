package netutil

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/netip"
	"regexp"
	"strings"

	"golang.org/x/net/idna"
)

// hostPattern validates domain names and localhost.
// IP validation is handled more reliably by net/netip.
var hostPattern = regexp.MustCompile(`^(localhost|([a-zA-Z0-9-]{1,63}\.)+[a-zA-Z]{2,63})$`)

// schemePattern matches a leading URL scheme such as http://, https://, or
// ftp:// in any letter case.
var schemePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://`)

// HasScheme reports whether the value starts with a URL scheme like http://.
func HasScheme(v string) bool {
	return schemePattern.MatchString(v)
}

// NormalizeHostWithSuffix normalizes the hostname portion of a URL-like
// input while preserving any path or query suffix.
//
// Examples:
//
//	"192.168.001.1/path" -> "192.168.1.1/path"
//	"[2001:db8::01]:443/api" -> "[2001:db8::1]/api"
//	"EXAMPLE.com" -> "example.com"
func NormalizeHostWithSuffix(input string) (string, error) {
	host, suffix, err := extractHostAndSuffix(input)
	if err != nil {
		return "", err
	}

	normalized, err := normalizeHost(host)
	if err != nil {
		return "", err
	}

	if addr, err := netip.ParseAddr(normalized); err == nil && addr.Is6() {
		return "[" + normalized + "]" + suffix, nil
	}

	return normalized + suffix, nil
}

// ExtractTLSServerName returns the normalized hostname for TLS SNI, without
// IPv6 brackets.
func ExtractTLSServerName(input string) (string, error) {
	host, _, err := extractHostAndSuffix(input)
	if err != nil {
		return "", err
	}

	return normalizeHost(host)
}

// ProtocolToScheme maps a protocol string to its URL scheme.
func ProtocolToScheme(protocol string) string {
	if IsHTTPS(protocol) {
		return "https://"
	}
	return "http://"
}

// IsHTTPS reports whether protocol means HTTPS.
func IsHTTPS(protocol string) bool {
	return strings.EqualFold(protocol, "https") ||
		strings.EqualFold(protocol, "https://")
}

// ParsePortOrDefault returns port, or defaultPort when out of range.
func ParsePortOrDefault(port int, defaultPort uint16) uint16 {
	if port < 0 || port > 65535 {
		return defaultPort
	}
	return uint16(port)
}

// extractHostAndSuffix separates a hostname from any trailing path or query
// string, handling schemes, ports, and IPv6 brackets.
func extractHostAndSuffix(input string) (host string, suffix string, err error) {
	input = strings.TrimSpace(input)

	if idx := strings.Index(input, "://"); idx != -1 {
		input = input[idx+3:]
	}

	if i := strings.IndexAny(input, "/?"); i != -1 {
		suffix = input[i:]
		input = input[:i]
	}

	if h, _, splitErr := net.SplitHostPort(input); splitErr == nil {
		host = h
	} else {
		// No port, but the host can still be bracketed IPv6 ([2001:db8::1]).
		host = strings.Trim(input, "[]")
	}

	if host == "" {
		return "", "", fmt.Errorf("empty host")
	}

	return host, suffix, nil
}

// normalizeHost converts and validates a hostname: IPs are canonicalized
// (leading zeros stripped), IDNs converted to ASCII, then matched against
// hostPattern.
func normalizeHost(host string) (string, error) {
	if addr, err := netip.ParseAddr(host); err == nil {
		return addr.String(), nil
	}

	ascii, err := idna.ToASCII(host)
	if err != nil {
		return "", fmt.Errorf("invalid host: %s", host)
	}

	if !hostPattern.MatchString(ascii) {
		return "", fmt.Errorf("invalid host: %s", host)
	}

	return ascii, nil
}

// ParseTLSVersion maps a version string to its crypto/tls constant.
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

// IsPortAvailable reports whether a TCP port is free on all local
// interfaces.
func IsPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

// ValidateDomain rejects empty domains, schemes, bad IDNA, empty/dot
// edges, and labels with bad syntax or length.
func ValidateDomain(domain string) error {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return fmt.Errorf("domain is required")
	}

	if HasScheme(domain) {
		return fmt.Errorf("invalid domain %q: domain must not include a protocol scheme (e.g. http:// or https://)", domain)
	}

	ascii, err := idna.Lookup.ToASCII(domain)
	if err != nil {
		return fmt.Errorf("invalid domain %q: %v", domain, err)
	}

	if strings.HasPrefix(ascii, ".") {
		return fmt.Errorf("invalid domain %q: domain cannot start with '.'", domain)
	}

	if strings.HasSuffix(ascii, ".") {
		return fmt.Errorf("invalid domain %q: domain cannot end with '.'", domain)
	}

	if strings.Contains(ascii, "..") {
		return fmt.Errorf("invalid domain %q: domain cannot contain consecutive dots", domain)
	}

	labels := strings.Split(ascii, ".")
	if len(labels) < 2 {
		return fmt.Errorf(
			"invalid domain %q: domain must contain at least two labels (for example, example.com)",
			domain,
		)
	}

	for _, label := range labels {
		if len(label) > 63 {
			return fmt.Errorf(
				"invalid domain %q: label %q is too long (maximum is 63 characters)",
				domain,
				label,
			)
		}

		if label[0] == '-' {
			return fmt.Errorf(
				"invalid domain %q: label %q cannot start with '-'",
				domain,
				label,
			)
		}

		if label[len(label)-1] == '-' {
			return fmt.Errorf(
				"invalid domain %q: label %q cannot end with '-'",
				domain,
				label,
			)
		}

		for _, r := range label {
			if (r >= 'a' && r <= 'z') ||
				(r >= 'A' && r <= 'Z') ||
				(r >= '0' && r <= '9') ||
				r == '-' {
				continue
			}

			return fmt.Errorf(
				"invalid domain %q: label %q contains invalid character %q",
				domain,
				label,
				string(r),
			)
		}
	}

	if len(ascii) > 253 {
		return fmt.Errorf(
			"invalid domain %q: domain is too long (maximum is 253 characters)",
			domain,
		)
	}

	return nil
}
