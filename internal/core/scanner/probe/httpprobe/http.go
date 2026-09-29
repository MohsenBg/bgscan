package httpprobe

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/logger"
)

// httpClientResult bundles a client with the function that releases its
// resources.
type httpClientResult struct {
	client *http.Client
	close  func()
}

// httpClientFactory abstracts HTTP client creation, primarily to allow
// mocking in tests.
type httpClientFactory func(ip netip.Addr) httpClientResult

// HTTPProbe validates HTTP/HTTPS connectivity to a target IP, preserving
// Host and SNI semantics.
type HTTPProbe struct {
	log           *logger.Logger
	req           HTTPRequest
	filter        statusFilter
	dialer        *net.Dialer
	tls           *tls.Config
	clientFactory httpClientFactory
}

// NewHTTPProbe creates an HTTPProbe. If acceptedCodes is empty or covers all
// known codes, all response status codes are accepted.
func NewHTTPProbe(req HTTPRequest, acceptedCodes []int, log *logger.Logger) probe.Probe {
	p := &HTTPProbe{
		log:    log,
		req:    req,
		dialer: &net.Dialer{Timeout: req.Timeout},
		tls:    newTLSConfig(req),
		filter: newStatusFilter(acceptedCodes, totalHTTPStatusCodes),
	}

	p.clientFactory = p.newStdClient
	if req.Fingerprint != "" {
		p.clientFactory = p.newUTLSClient
	}

	return p
}

// Init implements probe.Probe. It is a no-op.
func (p *HTTPProbe) Init(context.Context) error { return nil }

// Close implements probe.Probe. It is a no-op.
func (p *HTTPProbe) Close() error { return nil }

// Schema implements probe.Probe.
func (p *HTTPProbe) Schema() result.ResultSchema { return Schema }

// Run executes an HTTP HEAD request against the target IP. It returns an
// HTTPResult on success, or an error if the request fails or the response
// status code is not accepted.
func (p *HTTPProbe) Run(ctx context.Context, ip netip.Addr) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, p.req.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c := p.clientFactory(ip)
	defer c.close()

	start := time.Now()
	resp, err := c.client.Do(req)
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
		HTTPVersion: resp.Proto,
		UseTLS:      p.req.UseTLS,
		Latency:     latency,
	}, nil
}

// newUTLSClient builds a client with a spoofed TLS fingerprint, falling back
// to the standard client if the uTLS client cannot be created.
func (p *HTTPProbe) newUTLSClient(ip netip.Addr) httpClientResult {
	client, err := utlsHTTPClientFactory(
		ip,
		p.req.Timeout,
		p.req.Fingerprint,
		p.req.MinTLSVersion,
		p.req.MaxTLSVersion,
		p.req.SkipTLSVerify,
		p.req.Version,
	)
	if err != nil {
		p.log.Error("utls client for %s failed, falling back to std client: %v", ip, err)
		return p.newStdClient(ip)
	}

	return httpClientResult{client: client, close: client.CloseIdleConnections}
}

// newStdClient returns a net/http client bound to ip. A fresh transport per
// call avoids HTTP/2 readLoop goroutine leaks; callers must invoke close
// when done.
func (p *HTTPProbe) newStdClient(ip netip.Addr) httpClientResult {
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

	return httpClientResult{
		client: &http.Client{Transport: t, Timeout: p.req.Timeout},
		close:  t.CloseIdleConnections,
	}
}

// tlsNextProto disables HTTP/2 upgrades in H1-only mode (empty map); nil
// keeps default negotiation.
func tlsNextProto(v HTTPVersion) map[string]func(string, *tls.Conn) http.RoundTripper {
	if v == HTTPVersionH1 {
		return map[string]func(string, *tls.Conn) http.RoundTripper{}
	}
	return nil
}
