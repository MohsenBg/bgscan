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

	"github.com/MohsenBg/bgscan/internal/core/netutil"
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/logger"
)

type httpClientResult struct {
	client *http.Client
	close  func()
}

// httpClientFactory abstracts HTTP client creation, primarily to allow mocking in tests.
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

	if req.Fingerprint != "" {
		p.clientFactory = func(ip netip.Addr) httpClientResult {
			client, err := utlsHTTPClientFactory(
				ip,
				req.Timeout,
				req.Fingerprint,
				req.MinTLSVersion,
				req.MaxTLSVersion,
				req.SkipTLSVerify,
				req.Version,
			)
			if err != nil {
				t, c := p.buildClient(ip)
				return httpClientResult{client: c, close: t.CloseIdleConnections}
			}
			return httpClientResult{
				client: client,
				close:  client.CloseIdleConnections,
			}
		}
	} else {
		p.clientFactory = func(ip netip.Addr) httpClientResult {
			t, client := p.buildClient(ip)
			return httpClientResult{
				client: client,
				close:  t.CloseIdleConnections,
			}
		}
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

	r := p.clientFactory(ip)
	resp, err := r.client.Do(req)

	r.close()

	if err != nil {
		if netutil.IsUnreachable(err) {
			return nil, fmt.Errorf("%w: request failed: %w", probe.ErrEnvironment, err)
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			p.log.Error("close response body: %v", err)
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

func (p *HTTPProbe) Schema() result.ResultSchema {
	return Schema
}

// buildClient returns a transport + client bound to ip. A fresh transport
// per call avoids HTTP/2 readLoop goroutine leaks; callers must
// CloseIdleConnections when done.
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

// tlsNextProto disables HTTP/2 upgrades in H1-only mode (empty map); nil
// keeps default negotiation.
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
