package httpprobe

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/core/target"
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
type httpClientFactory func(r Request) httpClientResult

// HTTPProbe validates HTTP/HTTPS connectivity to a target, preserving
// Host and SNI semantics. How an entry becomes a request is decided by mode.
type HTTPProbe struct {
	log           *logger.Logger
	settings      Settings
	mode          Mode
	filter        statusFilter
	dialer        *net.Dialer
	clientFactory httpClientFactory
}

// NewHTTPProbe creates an HTTPProbe. If acceptedCodes is empty or covers all
// known codes, all response status codes are accepted.
func NewHTTPProbe(s Settings, mode Mode, acceptedCodes []int, log *logger.Logger) probe.Probe {
	p := &HTTPProbe{
		log:      log,
		settings: s,
		mode:     mode,
		dialer:   &net.Dialer{Timeout: s.Timeout},
		filter:   newStatusFilter(acceptedCodes, totalHTTPStatusCodes),
	}

	p.clientFactory = p.newStdClient
	if s.Fingerprint != "" {
		p.clientFactory = p.newUTLSClient
	}

	return p
}

// Input implements probe.Probe.
func (p *HTTPProbe) Input() target.Kind { return p.mode.Input() }

// Init implements probe.Probe. It is a no-op.
func (p *HTTPProbe) Init(context.Context) error { return nil }

// Close implements probe.Probe. It is a no-op.
func (p *HTTPProbe) Close() error { return nil }

// Schema implements probe.Probe.
func (p *HTTPProbe) Schema() result.ResultSchema { return Schema }

// Run executes an HTTP HEAD request for the entry. It returns an HTTPResult
// on success, or an error if the request fails or the response status code
// is not accepted.
func (p *HTTPProbe) Run(ctx context.Context, entry target.Entry) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r, err := p.mode.Build(entry)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, r.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if r.Host != "" {
		req.Host = r.Host
	}

	c := p.clientFactory(r)
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

	// Record the port that was really probed, not the list's value.
	entry.Port = r.Port

	return HTTPResult{
		Entry:       entry,
		StatusCode:  resp.StatusCode,
		HTTPVersion: resp.Proto,
		UseTLS:      p.settings.UseTLS,
		Latency:     latency,
	}, nil
}

// newUTLSClient builds a client with a spoofed TLS fingerprint, falling back
// to the standard client if the uTLS client cannot be created.
func (p *HTTPProbe) newUTLSClient(r Request) httpClientResult {
	client, err := utlsHTTPClientFactory(r, p.settings)
	if err != nil {
		p.log.Error("utls client for %s failed, falling back to std client: %v", r.URL, err)
		return p.newStdClient(r)
	}

	return httpClientResult{client: client, close: client.CloseIdleConnections}
}

// newStdClient returns a net/http client for one request. If r.DialIP is
// valid the connection is pinned to it; otherwise the address in the URL is
// resolved normally. A fresh transport per call avoids HTTP/2 readLoop
// goroutine leaks; callers must invoke close when done.
func (p *HTTPProbe) newStdClient(r Request) httpClientResult {
	t := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if r.DialIP.IsValid() {
				_, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, fmt.Errorf("parse addr: %w", err)
				}
				addr = net.JoinHostPort(r.DialIP.String(), port)
			}
			return p.dialer.DialContext(ctx, network, addr)
		},
		DisableKeepAlives:     true,
		TLSHandshakeTimeout:   p.settings.Timeout,
		ResponseHeaderTimeout: p.settings.Timeout,
		TLSClientConfig:       newTLSConfig(p.settings, r.SNI),
		ForceAttemptHTTP2:     p.settings.Version != HTTPVersionH1,
		TLSNextProto:          tlsNextProto(p.settings.Version),
	}

	return httpClientResult{
		client: &http.Client{Transport: t, Timeout: p.settings.Timeout},
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
