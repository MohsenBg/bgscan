package httpprobe

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/probe"
	"bgscan/internal/logger"
)

// httpClientFactory abstracts HTTP client creation, primarily to allow mocking in tests.
type httpClientFactory func(ip netip.Addr) (*http.Transport, *http.Client)

// HTTPProbe validates HTTP/HTTPS connectivity to a target IP, preserving
// Host and SNI semantics.
type HTTPProbe struct {
	req           HTTPRequest
	filter        statusFilter
	dialer        *net.Dialer
	tls           *tls.Config
	clientFactory httpClientFactory
}

// NewHTTPProbe creates an HTTPProbe. If acceptedCodes is empty or covers all
// known codes, all response status codes are accepted.
func NewHTTPProbe(req HTTPRequest, acceptedCodes []int) probe.Probe {
	p := &HTTPProbe{
		req:    req,
		dialer: &net.Dialer{Timeout: req.Timeout},
		tls:    newTLSConfig(req),
		filter: newStatusFilter(acceptedCodes, totalHTTPStatusCodes),
	}

	p.clientFactory = func(ip netip.Addr) (*http.Transport, *http.Client) {
		return p.buildClient(ip)
	}

	return p
}

// Init implements probe.Probe. It is a no-op.
func (p *HTTPProbe) Init(context.Context) error { return nil }

// Close implements probe.Probe. It is a no-op.
func (p *HTTPProbe) Close() error { return nil }

// Run executes an HTTP HEAD request against the target IP.
// It returns an HTTPResult on success, or an error if the request fails
// or the response status code is not in the accepted list.
func (p *HTTPProbe) Run(ctx context.Context, ip netip.Addr) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, p.req.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	start := time.Now()

	t, client := p.clientFactory(ip)
	resp, err := client.Do(req)

	t.CloseIdleConnections()

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.CoreError("close response body: %v", err)
		}
	}()

	if !p.filter.isAccepted(resp.StatusCode) {
		return nil, fmt.Errorf("status %d not accepted", resp.StatusCode)
	}

	return HTTPResult{
		IP:          ip,
		StatusCode:  resp.StatusCode,
		HTTPVersion: resp.Proto,
		UseTLS:      p.req.UseTLS,
		Latency:     time.Since(start),
	}, nil
}

// Schema returns the result schema for HTTP probes.
func (p *HTTPProbe) Schema() result.ResultSchema {
	return Schema
}

// buildClient creates a fresh *http.Transport and *http.Client bound to the target IP.
// A new transport is instantiated per call to prevent HTTP/2 readLoop goroutine leaks.
// The caller must invoke CloseIdleConnections on the returned transport after the request completes.
func (p *HTTPProbe) buildClient(ip netip.Addr) (*http.Transport, *http.Client) {
	t := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			_, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, fmt.Errorf("parse addr: %w", err)
			}
			return p.dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		},
		DisableKeepAlives:     true,
		TLSHandshakeTimeout:   p.req.Timeout,
		ResponseHeaderTimeout: p.req.Timeout,
		TLSClientConfig:       p.tls,
		ForceAttemptHTTP2:     p.req.Version == HTTPVersionH2,
		TLSNextProto:          tlsNextProto(p.req.Version),
	}

	return t, &http.Client{
		Transport: t,
		Timeout:   p.req.Timeout,
	}
}

// tlsNextProto configures ALPN behavior for the transport.
// An empty map disables HTTP/2 upgrades for H1-only mode, while nil allows default HTTP/2 negotiation.
func tlsNextProto(v HTTPVersion) map[string]func(authority string, c *tls.Conn) http.RoundTripper {
	if v == HTTPVersionH1 {
		return map[string]func(authority string, c *tls.Conn) http.RoundTripper{}
	}
	return nil
}

// isHTTPS reports whether the protocol string indicates an HTTPS connection.
func isHTTPS(proto string) bool {
	p := strings.ToLower(proto)
	p = strings.TrimSpace(p)
	return strings.HasPrefix(p, "https")
}
