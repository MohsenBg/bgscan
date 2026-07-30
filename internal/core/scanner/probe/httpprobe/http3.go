package httpprobe

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"time"

	"github.com/quic-go/quic-go/http3"

	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/probe"
	"bgscan/internal/logger"
)

// roundTripCloser abstracts the HTTP/3 transport, allowing tests to inject
// a mock without requiring a real QUIC connection.
type roundTripCloser interface {
	RoundTrip(*http.Request) (*http.Response, error)
	Close() error
}

// HTTP3Probe validates HTTP/3 (QUIC) connectivity to a target IP.
type HTTP3Probe struct {
	req       HTTPRequest
	filter    statusFilter
	transport roundTripCloser
}

// NewHTTP3Probe creates an HTTP3Probe. If acceptedCodes is empty or covers all
// known codes, all response status codes are accepted.
func NewHTTP3Probe(req HTTPRequest, acceptedCodes []int) (probe.Probe, error) {
	tlsCfg := newTLSConfig(req)

	return &HTTP3Probe{
		req:    req,
		filter: newStatusFilter(acceptedCodes, totalHTTPStatusCodes),
		transport: &http3.Transport{
			TLSClientConfig: tlsCfg,
		},
	}, nil
}

// Init implements probe.Probe. It is a no-op.
func (p *HTTP3Probe) Init(_ context.Context) error {
	return nil
}

// Close implements probe.Probe, releasing the underlying QUIC transport resources.
func (p *HTTP3Probe) Close() error {
	if err := p.transport.Close(); err != nil {
		return fmt.Errorf("close http3 transport: %w", err)
	}
	return nil
}

// Schema returns the result schema for HTTP/3 probes.
func (p *HTTP3Probe) Schema() result.ResultSchema {
	return Schema
}

// Run implements probe.Probe. It executes an HTTP/3 HEAD request against the target IP.
// It forces the QUIC connection to the specified IP while preserving the original
// hostname in the Host header for virtual hosting.
func (p *HTTP3Probe) Run(ctx context.Context, ip netip.Addr) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, p.req.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	// Force the QUIC connection to target the specific IP.
	req.URL.Host = net.JoinHostPort(ip.String(), req.URL.Port())

	// Preserve the original hostname for virtual hosting.
	req.Host = p.req.Host

	client := &http.Client{
		Transport: p.transport,
		Timeout:   p.req.Timeout,
	}

	start := time.Now()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
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
		HTTPVersion: "HTTP/3.0",
		UseTLS:      true,
		Latency:     time.Since(start),
	}, nil
}
