package httpprobe

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/quic-go/quic-go/http3"

	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/probe"
	"bgscan/internal/logger"
)

// HTTP3Probe performs HTTP/3 (QUIC-based) probing against an IP address.
type HTTP3Probe struct {
	req       HTTPRequest
	filter    statusFilter
	transport *http3.Transport
}

// NewHTTP3Probe returns a new HTTP/3 probe configured with the given request
// and accepted status codes.
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

// Init is a no-op for the HTTP/3 probe.
func (p *HTTP3Probe) Init(_ context.Context) error {
	return nil
}

// Close releases the underlying QUIC transport resources.
func (p *HTTP3Probe) Close() error {
	if err := p.transport.Close(); err != nil {
		return fmt.Errorf("close http3 transport: %w", err)
	}

	return nil
}

// Schema returns the result schema for HTTP probes.
func (p *HTTP3Probe) Schema() result.ResultSchema {
	return Schema
}

// Run executes the HTTP/3 probe against the specified IP address.
func (p *HTTP3Probe) Run(
	ctx context.Context,
	ip string,
) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodHead,
		p.req.URL,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	// Force QUIC to target IP.
	req.URL.Host = net.JoinHostPort(
		ip,
		req.URL.Port(),
	)

	// Preserve hostname for virtual hosting.
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
			logger.CoreError(
				"close response body: %v",
				err,
			)
		}
	}()

	if !p.filter.isAccepted(resp.StatusCode) {
		return nil, fmt.Errorf(
			"status %d not accepted",
			resp.StatusCode,
		)
	}

	return HTTPResult{
		IP:          ip,
		StatusCode:  resp.StatusCode,
		HTTPVersion: "HTTP/3.0",
		UseTLS:      true,
		Latency:     time.Since(start),
	}, nil
}
