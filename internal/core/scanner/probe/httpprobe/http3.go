package httpprobe

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"time"

	"github.com/quic-go/quic-go/http3"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/logger"
)

const defaultHTTPSPort = "443"

// roundTripCloser abstracts the HTTP/3 transport, allowing tests to inject
// a mock without requiring a real QUIC connection.
type roundTripCloser interface {
	RoundTrip(*http.Request) (*http.Response, error)
	Close() error
}

// HTTP3Probe validates HTTP/3 (QUIC) connectivity to a target IP.
type HTTP3Probe struct {
	log       *logger.Logger
	req       HTTPRequest
	filter    statusFilter
	transport roundTripCloser
}

// NewHTTP3Probe creates an HTTP3Probe. If acceptedCodes is empty or covers all
// known codes, all response status codes are accepted.
func NewHTTP3Probe(req HTTPRequest, acceptedCodes []int, log *logger.Logger) (probe.Probe, error) {
	return &HTTP3Probe{
		log:    log,
		req:    req,
		filter: newStatusFilter(acceptedCodes, totalHTTPStatusCodes),
		transport: &http3.Transport{
			TLSClientConfig: newTLSConfig(req),
		},
	}, nil
}

// Init implements probe.Probe. It is a no-op.
func (p *HTTP3Probe) Init(context.Context) error { return nil }

// Close implements probe.Probe, releasing the underlying QUIC transport.
func (p *HTTP3Probe) Close() error {
	if err := p.transport.Close(); err != nil {
		return fmt.Errorf("close http3 transport: %w", err)
	}
	return nil
}

// Schema implements probe.Probe.
func (p *HTTP3Probe) Schema() result.ResultSchema { return Schema }

// Run executes an HTTP/3 HEAD request against the target IP. The QUIC
// connection is forced to the given IP while the original hostname is kept
// in the Host header for virtual hosting.
func (p *HTTP3Probe) Run(ctx context.Context, ip netip.Addr) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, p.req.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	port := req.URL.Port()
	if port == "" {
		port = defaultHTTPSPort
	}

	// Dial the target IP, but keep the original hostname for virtual hosting.
	req.URL.Host = net.JoinHostPort(ip.String(), port)
	req.Host = p.req.Host

	client := &http.Client{
		Transport: p.transport,
		Timeout:   p.req.Timeout,
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, probe.NormalizeErr(err)
	}
	latency := time.Since(start)

	defer func() {
		if err := resp.Body.Close(); err != nil {
			p.log.Error("close response body: %v", err)
		}
	}()

	if !p.filter.isAccepted(resp.StatusCode) {
		return nil, fmt.Errorf("%w: status %d not accepted", probe.ErrBadResponse, resp.StatusCode)
	}

	return HTTPResult{
		IP:          ip,
		StatusCode:  resp.StatusCode,
		HTTPVersion: "HTTP/3.0",
		UseTLS:      true,
		Latency:     latency,
	}, nil
}
