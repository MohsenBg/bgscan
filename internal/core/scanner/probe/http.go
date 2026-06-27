package probe

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"bgscan/internal/core/config"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/netutil"
	"bgscan/internal/logger"
)

//
// HTTP PROBE
//

// HTTPProbe implements the Probe interface for performing HTTP or HTTPS
// service validation during scanning.
//
// The probe connects directly to a target IP address while preserving
// the original hostname semantics for:
//
//   - HTTP Host headers
//   - TLS Server Name Indication (SNI)
//
// This design allows accurate scanning of virtual‑host based services
// and TLS endpoints while bypassing DNS resolution entirely.
type HTTPProbe struct {
	req HTTPRequest
}

// HTTPRequest represents a fully resolved HTTP request configuration
// used by HTTPProbe during execution.
//
// Instances of HTTPRequest are typically produced by
// NewHTTPRequestFromConfig after validating and normalizing user input.
type HTTPRequest struct {
	// URL is the complete request URL including scheme, hostname, and port.
	//
	// Example:
	//   https://example.com:443
	URL string

	// Host is the normalized hostname extracted from configuration.
	// It is primarily used for hostname validation and TLS server identity.
	Host string

	// SNI defines the Server Name Indication value used during the TLS
	// handshake when HTTPS is enabled.
	//
	// If empty, the TLS implementation may fall back to the hostname.
	SNI string

	// UseTLS indicates whether the request should use HTTPS (true)
	// or plain HTTP (false).
	UseTLS bool

	// SkipTLSVerify disables TLS certificate verification.
	//
	// This is commonly used in scanning environments where certificates
	// may not match the scanned hostname.
	SkipTLSVerify bool

	// Timeout defines the maximum duration allowed for the entire request
	// lifecycle including dialing, TLS handshake, and response headers.
	Timeout time.Duration

	// MinTLSVersion defines the minimum TLS version allowed for the
	// connection handshake.
	MinTLSVersion uint16

	// MaxTLSVersion defines the maximum TLS version allowed for the
	// connection handshake.
	MaxTLSVersion uint16
}

// NewHTTPProbe constructs a new HTTPProbe using the provided HTTPRequest
// configuration.
//
// The returned value implements the Probe interface and can be used
// by the scanner engine to perform HTTP service checks against IP
// addresses.
func NewHTTPProbe(req HTTPRequest) Probe {
	return &HTTPProbe{req: req}
}

// Init prepares the HTTPProbe for execution.
//
// HTTPProbe does not require background initialization, therefore
// Init currently performs no work and returns nil.
//
// The method exists to satisfy the Probe lifecycle contract
// (Init → Run → Close).
func (p *HTTPProbe) Init(ctx context.Context) error {
	return nil
}

// Run executes an HTTP probe against the provided IP address.
//
// The probe performs the following steps:
//
//  1. Builds a custom http.Client bound to the target IP.
//  2. Creates an HTTP HEAD request to minimize bandwidth usage.
//  3. Establishes a TCP connection directly to the IP address.
//  4. Preserves the original hostname for HTTP Host headers and TLS SNI.
//  5. Measures request latency.
//
// The function respects context cancellation and will abort early
// if the provided context is canceled.
func (p *HTTPProbe) Run(ctx context.Context, ip string) (*result.IPScanResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client := p.buildHTTPClient(ip)

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, p.req.URL, nil)
	if err != nil {
		return nil, err
	}

	start := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.CoreError("error closing response body: %v", err)
		}
	}()

	return &result.IPScanResult{
		IP:      ip,
		Latency: time.Since(start),
	}, nil
}

//
// HTTP CLIENT CONSTRUCTION
//

// buildHTTPClient constructs an http.Client configured specifically
// for scanning workflows.
//
// The returned client:
//
//   - Forces TCP connections to the provided IP address
//   - Preserves hostname information for HTTP and TLS layers
//   - Disables connection reuse for predictable probe behavior
//   - Applies strict timeouts suitable for high‑volume scans
func (p *HTTPProbe) buildHTTPClient(ip string) *http.Client {
	dialer := &net.Dialer{
		Timeout: p.req.Timeout,
	}

	transport := &http.Transport{
		// DialContext overrides the destination address so the TCP
		// connection is made directly to the provided IP instead of
		// resolving the hostname via DNS.
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			_, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			target := net.JoinHostPort(ip, port)
			return dialer.DialContext(ctx, network, target)
		},

		// DisableKeepAlives ensures each probe uses a fresh TCP connection.
		// This prevents connection reuse side‑effects during large scans.
		DisableKeepAlives: true,

		ResponseHeaderTimeout: p.req.Timeout,
		TLSHandshakeTimeout:   p.req.Timeout,
	}

	if p.req.UseTLS {
		transport.TLSClientConfig = &tls.Config{
			ServerName:         p.req.SNI,
			InsecureSkipVerify: p.req.SkipTLSVerify,
			MinVersion:         p.req.MinTLSVersion,
			MaxVersion:         p.req.MaxTLSVersion,
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   p.req.Timeout,
	}
}

//
// CONFIG NORMALIZATION
//

// NewHTTPRequestFromConfig converts a user‑provided HTTPConfig into a
// normalized HTTPRequest suitable for execution by HTTPProbe.
//
// The function performs several validation and normalization steps:
//
//   - Resolves protocol scheme (HTTP or HTTPS)
//   - Normalizes and validates hostnames (including IDN handling)
//   - Determines the correct service port
//   - Resolves TLS Server Name Indication (SNI)
//   - Parses TLS version constraints
func NewHTTPRequestFromConfig(cfg config.HTTPConfig) (*HTTPRequest, error) {
	scheme := netutil.ProtocolToScheme(cfg.Protocol)
	isHTTPS := netutil.IsHTTPS(scheme)

	host, err := netutil.ExtractTLSServerName(cfg.Host)
	if err != nil {
		return nil, err
	}

	urlHost, err := netutil.NormalizeHostWithSuffix(cfg.Host)
	if err != nil {
		return nil, err
	}

	port := resolvePort(cfg.Port, isHTTPS)
	url := fmt.Sprintf("%s%s:%d", scheme, urlHost, port)

	sni, err := resolveSNI(cfg.ServerName, isHTTPS)
	if err != nil {
		return nil, err
	}

	minTLS, maxTLS, err := resolveTLSVersions(cfg)
	if err != nil {
		return nil, err
	}

	return &HTTPRequest{
		URL:           url,
		Host:          host,
		SNI:           sni,
		UseTLS:        isHTTPS,
		SkipTLSVerify: !cfg.TLSValidation,
		Timeout:       cfg.Timeout.Duration(),
		MinTLSVersion: minTLS,
		MaxTLSVersion: maxTLS,
	}, nil
}

//
// HELPER FUNCTIONS
//

// resolvePort returns the configured port if provided,
// otherwise the protocol default (443 for HTTPS, 80 for HTTP).
func resolvePort(port int, isHTTPS bool) uint16 {
	if port > 0 {
		return uint16(port)
	}

	if isHTTPS {
		return 443
	}

	return 80
}

// resolveSNI determines the TLS Server Name Indication value
// based on configuration.
//
// If HTTPS is disabled or no server name is provided,
// an empty string is returned.
func resolveSNI(serverName string, isHTTPS bool) (string, error) {
	if !isHTTPS || serverName == "" {
		return "", nil
	}

	return netutil.ExtractTLSServerName(serverName)
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

// Close implements the Probe interface cleanup method.
//
// HTTPProbe does not maintain long‑lived resources,
// therefore Close currently performs no action.
func (p *HTTPProbe) Close() error {
	return nil
}
