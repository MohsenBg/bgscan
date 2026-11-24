package httpprobe

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/quic-go/quic-go/http3"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/core/target"
	"github.com/MohsenBg/bgscan/internal/logger"
)

const defaultHTTPSPort = "443"

// roundTripCloser abstracts the HTTP/3 transport, allowing tests to inject
// a mock without requiring a real QUIC connection.
type roundTripCloser interface {
	RoundTrip(*http.Request) (*http.Response, error)
	Close() error
}

// HTTP3Probe validates HTTP/3 (QUIC) connectivity to a target.
//
// A QUIC transport has one TLS config, so transports are cached per SNI.
// Host mode has a single SNI and direct mode uses "", so in practice this
// holds one transport.
type HTTP3Probe struct {
	log      *logger.Logger
	settings Settings
	mode     Mode
	filter   statusFilter

	// newTransport builds the transport for one SNI; tests replace it.
	newTransport func(sni string) roundTripCloser

	mu         sync.Mutex
	transports map[string]roundTripCloser
}

// NewHTTP3Probe creates an HTTP3Probe. If acceptedCodes is empty or covers all
// known codes, all response status codes are accepted.
func NewHTTP3Probe(s Settings, mode Mode, acceptedCodes []int, log *logger.Logger) (probe.Probe, error) {
	return &HTTP3Probe{
		log:      log,
		settings: s,
		mode:     mode,
		filter:   newStatusFilter(acceptedCodes, totalHTTPStatusCodes),
		newTransport: func(sni string) roundTripCloser {
			return &http3.Transport{TLSClientConfig: newTLSConfig(s, sni)}
		},
		transports: make(map[string]roundTripCloser),
	}, nil
}

// transportFor returns the cached transport for sni, creating it on first use.
func (p *HTTP3Probe) transportFor(sni string) roundTripCloser {
	p.mu.Lock()
	defer p.mu.Unlock()

	t, ok := p.transports[sni]
	if !ok {
		t = p.newTransport(sni)
		p.transports[sni] = t
	}
	return t
}

// Input returns the target kind the configured mode accepts.
func (p *HTTP3Probe) Input() target.Kind { return p.mode.Input() }

// Init does nothing. Transports are created on first use in Run.
func (p *HTTP3Probe) Init(context.Context) error { return nil }

// Close releases the underlying QUIC transports and returns the
// joined close errors, if any.
func (p *HTTP3Probe) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errs []error
	for sni, t := range p.transports {
		if err := t.Close(); err != nil {
			errs = append(errs, err)
		}
		delete(p.transports, sni)
	}

	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("close http3 transport: %w", err)
	}
	return nil
}

// Schema returns the result schema shared by the HTTP probes.
func (p *HTTP3Probe) Schema() result.ResultSchema { return Schema }

// Run executes an HTTP/3 HEAD request for the entry. If the request
// has a DialIP, the QUIC connection is forced to it while the
// original host is kept in the Host header for virtual hosting. It
// returns probe.ErrBadResponse when the response status code is not
// accepted.
func (p *HTTP3Probe) Run(ctx context.Context, entry target.Entry) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r, err := p.mode.Build(entry)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, r.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	// Decide the Host header before the URL host is rewritten.
	hostHeader := r.Host
	if hostHeader == "" {
		hostHeader = req.URL.Host
	}

	if r.DialIP.IsValid() {
		port := req.URL.Port()
		if port == "" {
			port = defaultHTTPSPort
		}
		req.URL.Host = net.JoinHostPort(r.DialIP.String(), port)
	}
	req.Host = hostHeader

	client := &http.Client{
		Transport: p.transportFor(r.SNI),
		Timeout:   p.settings.Timeout,
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

	// Record the port that was really probed, not the list's value.
	entry.Port = r.Port

	return HTTPResult{
		Entry:       entry,
		StatusCode:  resp.StatusCode,
		HTTPVersion: "HTTP/3.0",
		UseTLS:      true,
		Latency:     latency,
	}, nil
}
