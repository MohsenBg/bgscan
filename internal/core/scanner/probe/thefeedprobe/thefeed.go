package thefeedprobe

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
)

// TheFeedProbe verifies that a scanned IP can serve as a DNS resolver for
// thefeed content. Unlike the tunnel-based probes it allocates no local
// listener: it sends a single encrypted metadata TXT query through the
// target and checks whether the response decrypts with the configured
// passphrase.
type TheFeedProbe struct {
	config     dns.TheFeedConfig
	thefeedSvc dns.TheFeedService
	timeout    time.Duration
	tries      int
}

type Option func(*TheFeedProbe)

func WithTheFeedService(service dns.TheFeedService) Option {
	return func(p *TheFeedProbe) {
		if service != nil {
			p.thefeedSvc = service
		}
	}
}

// WithTries sets how many times a failed probe is retried.
// Values below 1 are ignored; the default is a single attempt.
func WithTries(n int) Option {
	return func(p *TheFeedProbe) {
		if n >= 1 {
			p.tries = n
		}
	}
}

// NewTheFeedProbe creates a TheFeed probe.
func NewTheFeedProbe(
	config dns.TheFeedConfig,
	timeout time.Duration,
	opts ...Option,
) (probe.Probe, error) {
	if errs := config.Validate(); len(errs) != 0 {
		return nil, joinConfigErrors(errs)
	}

	p := &TheFeedProbe{
		config:  config,
		timeout: timeout,
		tries:   1,
	}

	for _, opt := range opts {
		opt(p)
	}

	if p.thefeedSvc == nil {
		p.thefeedSvc = dns.NewTheFeedService()
	}

	return p, nil
}

func joinConfigErrors(errs map[string]error) error {
	var joined error
	for field, err := range errs {
		joined = fmt.Errorf("%s: %w", field, err)
	}
	return joined
}

// Schema returns the result schema emitted by the probe.
func (p *TheFeedProbe) Schema() result.ResultSchema {
	return Schema
}

// Init initializes the probe.
func (p *TheFeedProbe) Init(context.Context) error {
	return nil
}

// Run probes ip as a thefeed DNS resolver and validates the encrypted
// metadata response, retrying failed attempts up to the configured tries.
func (p *TheFeedProbe) Run(ctx context.Context, ip netip.Addr) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	tries := p.tries
	if tries < 1 {
		tries = 1
	}

	var err error
	for attempt := 0; attempt < tries; attempt++ {
		var res result.Result
		if res, err = p.runOnce(ctx, ip); err == nil {
			return res, nil
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
	}

	return nil, err
}

func (p *TheFeedProbe) runOnce(ctx context.Context, ip netip.Addr) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	start := time.Now()
	err := p.thefeedSvc.RunTunnel(ctx, p.config, ip, p.timeout)
	latency := time.Since(start)

	if err != nil {
		return nil, err
	}

	return TheFeedResult{
		IP:        ip,
		Latency:   latency,
		Domain:    p.config.Domain,
		QueryMode: p.config.QueryMode,
	}, nil
}

// Close releases resources owned by the probe.
func (p *TheFeedProbe) Close() error {
	return nil
}
