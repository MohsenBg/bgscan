package httpprobe

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/probe"
	"bgscan/internal/logger"
)

// HTTPProbe performs HTTP/HTTPS validation against a target IP while preserving
// Host and SNI semantics.
type HTTPProbe struct {
	req    HTTPRequest
	filter statusFilter
	dialer *net.Dialer
	tls    *tls.Config
}

// NewHTTPProbe creates a new HTTPProbe with optional accepted status codes.
// If acceptedCodes is empty or covers all known codes, all responses are accepted.
func NewHTTPProbe(req HTTPRequest, acceptedCodes []int) probe.Probe {
	return &HTTPProbe{
		req:    req,
		filter: newStatusFilter(acceptedCodes, totalHTTPStatusCodes),
		dialer: &net.Dialer{Timeout: req.Timeout},
		tls:    newTLSConfig(req),
	}
}

// Init implements [probe.Probe]. It is a no-op for HTTP probes.
func (p *HTTPProbe) Init(context.Context) error { return nil }

// Close implements [probe.Probe]. It is a no-op for HTTP probes.
func (p *HTTPProbe) Close() error { return nil }

// Run executes a single HTTP HEAD request against the target IP.
func (p *HTTPProbe) Run(ctx context.Context, ip string) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, p.req.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	start := time.Now()

	t, client := p.buildClient(ip)
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

// buildClient creates a fresh *http.Transport bound to the given target IP.
//
// A new transport is created per Run call because HTTP/2 connections spawn a
// persistent readLoop goroutine per connection that only exits when the
// transport is closed. The caller must call t.CloseIdleConnections() after
// the request completes to release that goroutine immediately.
func (p *HTTPProbe) buildClient(ip string) (*http.Transport, *http.Client) {
	t := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			_, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, fmt.Errorf("parse addr: %w", err)
			}
			return p.dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
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

// tlsNextProto returns the TLSNextProto map for the transport.
// For H1-only mode, returning an empty (non-nil) map disables ALPN-based
// HTTP/2 upgrade even if the server advertises h2 in its TLS handshake.
// For all other modes nil lets the transport manage h2 itself.
func tlsNextProto(v HTTPVersion) map[string]func(authority string, c *tls.Conn) http.RoundTripper {
	if v == HTTPVersionH1 {
		return map[string]func(authority string, c *tls.Conn) http.RoundTripper{}
	}
	return nil
}

func isHTTPS(proto string) bool {
	p := strings.ToLower(proto)
	p = strings.TrimSpace(p)
	return strings.HasPrefix(p, "https")
}
