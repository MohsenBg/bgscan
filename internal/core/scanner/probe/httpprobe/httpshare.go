package httpprobe

import (
	"crypto/tls"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/core/netutil"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

// totalHTTPStatusCodes is the approximate number of recognized HTTP status codes.
// If an accepted-codes list reaches this size, it is treated as "accept all".
const totalHTTPStatusCodes = 63

// HTTPVersion represents the HTTP protocol negotiation mode via ALPN.
type HTTPVersion uint8

const (
	// HTTPVersionH1H2 enables HTTP/1.1 and HTTP/2 negotiation (default).
	HTTPVersionH1H2 HTTPVersion = iota

	// HTTPVersionH1 forces HTTP/1.1 only.
	HTTPVersionH1

	// HTTPVersionH2 forces HTTP/2 only (requires TLS).
	HTTPVersionH2
)

// Settings contains probe configuration shared across all targets.
// It is built once from config and remains immutable during a scan.
// Target-specific settings belong in Request (see mode.go).
// MinTLSVersion and MaxTLSVersion hold TLS version numbers as defined
// in crypto/tls, and zero leaves that end of the range at the Go
// default. Fingerprint names a uTLS client fingerprint, and an empty
// Fingerprint uses the standard TLS stack.
type Settings struct {
	Timeout       time.Duration
	Version       HTTPVersion
	UseTLS        bool
	SkipTLSVerify bool
	Fingerprint   string
	MinTLSVersion uint16
	MaxTLSVersion uint16

	// DefaultPort is used when dynamic port selection is disabled
	// or the target does not provide a port.
	DefaultPort uint16
	DynamicPort bool
}

// NewSettings builds Settings from an HTTPConfig.
func NewSettings(cfg config.HTTPConfig) (Settings, error) {
	useTLS := isHTTPS(cfg.Protocol)

	port, err := defaultPort(cfg.Port, useTLS)
	if err != nil {
		return Settings{}, err
	}

	minTLS, maxTLS, err := resolveTLSVersions(cfg)
	if err != nil {
		return Settings{}, err
	}

	return Settings{
		Timeout:       cfg.Timeout.Duration(),
		Version:       resolveHTTPVersion(cfg.Version),
		UseTLS:        useTLS,
		SkipTLSVerify: !cfg.TLSValidation,
		Fingerprint:   cfg.Fingerprint,
		MinTLSVersion: minTLS,
		MaxTLSVersion: maxTLS,
		DefaultPort:   port,
		DynamicPort:   cfg.DynamicPort,
	}, nil
}

func (s Settings) scheme() string {
	if s.UseTLS {
		return "https"
	}
	return "http"
}

// resolvePort picks the port to probe for an entry. With dynamic
// ports enabled, a nonzero entry port wins. Otherwise the configured
// default port is used.
func (s Settings) resolvePort(e target.Entry) uint16 {
	if s.DynamicPort && e.Port != 0 {
		return e.Port
	}
	return s.DefaultPort
}

// statusFilter is an allow-list of valid HTTP status codes; the zero value
// accepts everything.
type statusFilter struct {
	accepted map[int]struct{}
}

// newStatusFilter builds a statusFilter; empty or exhaustive lists accept all.
func newStatusFilter(codes []int, total int) statusFilter {
	if len(codes) == 0 || len(codes) >= total {
		return statusFilter{}
	}

	m := make(map[int]struct{}, len(codes))
	for _, c := range codes {
		m[c] = struct{}{}
	}

	return statusFilter{accepted: m}
}

func (f statusFilter) isAccepted(code int) bool {
	if len(f.accepted) == 0 {
		return true
	}
	_, ok := f.accepted[code]
	return ok
}

// newTLSConfig builds a *tls.Config for one request, or nil when TLS is
// disabled. An empty sni lets Go derive the server name from the URL host.
func newTLSConfig(s Settings, sni string) *tls.Config {
	if !s.UseTLS {
		return nil
	}
	return &tls.Config{
		ServerName:         sni,
		InsecureSkipVerify: s.SkipTLSVerify,
		MinVersion:         s.MinTLSVersion,
		MaxVersion:         s.MaxTLSVersion,
	}
}

// defaultPort returns port, or 443/80 for TLS/plain HTTP when port is 0.
func defaultPort(port int, useTLS bool) (uint16, error) {
	if port < 0 || port > math.MaxUint16 {
		return 0, errors.New("invalid port number")
	}

	if port != 0 {
		return uint16(port), nil
	}
	if useTLS {
		return 443, nil
	}
	return 80, nil
}

// resolveHTTPVersion maps a protocol string to an HTTPVersion
// ("h1"/"http/1.x" → H1, "h2"/"http/2" → H2, anything else → H1H2).
func resolveHTTPVersion(protocol string) HTTPVersion {
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case "h1", "http/1", "http/1.1", "http1":
		return HTTPVersionH1
	case "h2", "http/2", "http2":
		return HTTPVersionH2
	default:
		return HTTPVersionH1H2
	}
}

// resolveTLSVersions parses the TLS version range, rejecting min > max.
func resolveTLSVersions(cfg config.HTTPConfig) (uint16, uint16, error) {
	minTLS, err := netutil.ParseTLSVersion(cfg.MinTLSVersion)
	if err != nil {
		return 0, 0, err
	}

	maxTLS, err := netutil.ParseTLSVersion(cfg.MaxTLSVersion)
	if err != nil {
		return 0, 0, err
	}

	if minTLS > maxTLS {
		return 0, 0, fmt.Errorf("min TLS version cannot be greater than max TLS version")
	}

	return minTLS, maxTLS, nil
}

func isHTTPS(s string) bool {
	s = strings.TrimSpace(s)
	return len(s) >= len("https") && strings.EqualFold(s[:len("https")], "https")
}
