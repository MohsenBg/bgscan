package httpprobe

import (
	"crypto/tls"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"bgscan/internal/core/config"
	"bgscan/internal/core/scanner/netutil"
)

// totalHTTPStatusCodes is the total number of recognized HTTP status codes.
// Used to determine whether a supplied accepted-codes list effectively means
// "accept everything" (i.e. the list covers all codes).
const totalHTTPStatusCodes = 63

// HTTPVersion represents the HTTP protocol negotiation mode via ALPN.
type HTTPVersion uint8

const (
	// HTTPVersionH1H2 enables HTTP/1.1 and HTTP/2 negotiation (default).
	HTTPVersionH1H2 HTTPVersion = iota

	// HTTPVersionH1 forces HTTP/1.1 only.
	HTTPVersionH1

	// HTTPVersionH2 forces HTTP/2 only (TLS required).
	HTTPVersionH2
)

// HTTPRequest is a normalized, ready-to-execute HTTP probe configuration.
// It is shared by both HTTPProbe (HTTP/1.1, HTTP/2) and HTTP3Probe (QUIC).
type HTTPRequest struct {
	URL           string
	Host          string
	SNI           string
	Version       HTTPVersion
	UseTLS        bool
	SkipTLSVerify bool
	Timeout       time.Duration
	MinTLSVersion uint16
	MaxTLSVersion uint16
}

// statusFilter is an optional allow-list of HTTP status codes considered
// valid by a probe. A zero-value statusFilter accepts every status code.
type statusFilter struct {
	accepted map[int]struct{}
}

// newStatusFilter builds a statusFilter from a slice of accepted status codes.
//
// If codes is empty, or covers at least total distinct values (i.e. effectively
// the full set), the filter is left empty and will accept everything.
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

// newTLSConfig builds a *tls.Config from an HTTPRequest.
// Returns nil when req.UseTLS is false.
func newTLSConfig(req HTTPRequest) *tls.Config {
	if !req.UseTLS {
		return nil
	}

	return &tls.Config{
		ServerName:         req.SNI,
		InsecureSkipVerify: req.SkipTLSVerify,
		MinVersion:         req.MinTLSVersion,
		MaxVersion:         req.MaxTLSVersion,
	}
}

// defaultPort returns the default port for HTTP or HTTPS depending on useTLS.
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

// resolveSNI returns a validated SNI value.
// If serverName is empty, it returns an empty string without error,
// allowing the caller to derive the SNI from the host (kept as an extension point).
func resolveSNI(serverName string, useTLS bool) (string, error) {
	if serverName != "" {
		return serverName, nil
	}
	if !useTLS {
		return "", nil
	}
	return "", nil // callers may derive SNI from host; kept as extension point
}

// resolveHTTPVersion maps a protocol string from config to an HTTPVersion.
// Recognised values (case-insensitive): "h1", "http/1", "http/1.1" → H1;
// "h2", "http/2" → H2; anything else (including empty) → H1H2.
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

// NewHTTPRequestFromConfig builds an HTTPRequest from a generic HTTPConfig,
// suitable for HTTP/1.1 and HTTP/2 probing.
func NewHTTPRequestFromConfig(cfg config.HTTPConfig) (*HTTPRequest, error) {
	scheme := "http://"
	useHTTPS := isHTTPS(cfg.Protocol)

	if useHTTPS {
		scheme = "https://"
	}

	host, err := netutil.ExtractTLSServerName(cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("extract host: %w", err)
	}

	urlHost, err := netutil.NormalizeHostWithSuffix(cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("normalize host: %w", err)
	}

	port, err := defaultPort(cfg.Port, useHTTPS)
	if err != nil {
		return nil, err
	}

	sni := cfg.ServerName
	if sni == "" {
		sni = host
	}

	sni, err = resolveSNI(sni, useHTTPS)
	if err != nil {
		return nil, fmt.Errorf("resolve SNI: %w", err)
	}

	minTLS, maxTLS, err := resolveTLSVersions(cfg)
	if err != nil {
		return nil, err
	}

	return &HTTPRequest{
		URL:           fmt.Sprintf("%s%s:%d", scheme, urlHost, port),
		Host:          host,
		SNI:           sni,
		Version:       resolveHTTPVersion(cfg.Version), // was never set before
		UseTLS:        useHTTPS,
		SkipTLSVerify: !cfg.TLSValidation,
		Timeout:       cfg.Timeout.Duration(),
		MinTLSVersion: minTLS,
		MaxTLSVersion: maxTLS,
	}, nil
}

// resolveTLSVersions parses TLS version constraints from configuration
// and validates that the minimum version is not greater than the maximum.
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
