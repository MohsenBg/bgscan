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
}

type Option func(*TheFeedProbe)

func WithTheFeedService(service dns.TheFeedService) Option {
	return func(p *TheFeedProbe) {
		if service != nil {
			p.thefeedSvc = service
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
// metadata response.
func (p *TheFeedProbe) Run(ctx context.Context, ip netip.Addr) (result.Result, error) {
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
