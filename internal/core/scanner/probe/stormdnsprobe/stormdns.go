package stormdnsprobe

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner/portmgr"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/core/socks"
	"github.com/MohsenBg/bgscan/internal/core/speedtest"
	"github.com/MohsenBg/bgscan/internal/core/target"
	"github.com/MohsenBg/bgscan/internal/logger"
)

// StormDNSProbe verifies connectivity through a StormDNS tunnel.
//
// The embedded StormDNS client exposes a local no-auth SOCKS5 listener;
// the probe boots it against the scanned IP as resolver and measures
// latency through that listener.
type StormDNSProbe struct {
	log             *logger.Logger
	pm              portmgr.Manager
	config          dns.StormDNSConfig
	stormDNSService dns.StormDNSService
	socksService    socks.Service
	speedtestSvc    speedtest.Service
	timeout         time.Duration
	tries           int
}

func (p *StormDNSProbe) Input() target.Kind {
	return target.KindIP
}

type Option func(*StormDNSProbe)

// WithLogger sets the logger for probe lifecycle events (nil-safe).
func WithLogger(log *logger.Logger) Option {
	return func(p *StormDNSProbe) {
		p.log = log
	}
}

func WithStormDNSService(service dns.StormDNSService) Option {
	return func(p *StormDNSProbe) {
		if service != nil {
			p.stormDNSService = service
		}
	}
}

func WithSocksService(service socks.Service) Option {
	return func(p *StormDNSProbe) {
		if service != nil {
			p.socksService = service
		}
	}
}

func WithSpeedtestService(service speedtest.Service) Option {
	return func(p *StormDNSProbe) {
		if service != nil {
			p.speedtestSvc = service
		}
	}
}

// WithTries sets how many times a failed probe is retried.
// Values below 1 are ignored; the default is a single attempt.
func WithTries(n int) Option {
	return func(p *StormDNSProbe) {
		if n >= 1 {
			p.tries = n
		}
	}
}

func NewStormDNSProbe(
	config dns.StormDNSConfig,
	timeout time.Duration,
	pm portmgr.Manager,
	opts ...Option,
) (probe.Probe, error) {
	if pm == nil {
		return nil, fmt.Errorf("port manager is nil")
	}

	if errs := config.Validate(); len(errs) != 0 {
		return nil, joinConfigErrors(errs)
	}

	p := &StormDNSProbe{
		pm:      pm,
		config:  config,
		timeout: timeout,
		tries:   1,
	}

	for _, opt := range opts {
		opt(p)
	}

	if p.stormDNSService == nil {
		p.stormDNSService = dns.NewStormDNSService()
	}
	if p.socksService == nil {
		p.socksService = socks.NewService()
	}
	if p.speedtestSvc == nil {
		p.speedtestSvc = speedtest.NewService()
	}

	return p, nil
}

func (p *StormDNSProbe) Schema() result.ResultSchema {
	return Schema
}

func (p *StormDNSProbe) Init(context.Context) error {
	return nil
}

// Run boots the embedded StormDNS client against the entry's address and
// measures latency through its local SOCKS5 listener, retrying failed
// attempts up to the configured tries.
//
// Domain targets are refused: a probe needs an address, and nothing resolves
// names yet.
func (p *StormDNSProbe) Run(ctx context.Context, entry target.Entry) (result.Result, error) {
	if entry.Kind != target.KindIP {
		return nil, fmt.Errorf("%w: %s", probe.ErrDomainTarget, entry.Domain)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	localPort, err := p.pm.Get(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return nil, err
		}
		return nil, fmt.Errorf("lease port for StormDNS: %w", err)
	}
	defer p.pm.Release(localPort)

	tries := p.tries
	if tries < 1 {
		tries = 1
	}

	for attempt := 0; attempt < tries; attempt++ {
		var res result.Result
		if res, err = p.runOnce(ctx, entry, localPort); err == nil {
			return res, nil
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
	}

	return nil, probe.NormalizeErr(err)
}

func (p *StormDNSProbe) runOnce(ctx context.Context, entry target.Entry, localPort uint16) (result.Result, error) {
	// RunTunnel blocks until the tunnel session is ready.
	handle, err := p.stormDNSService.RunTunnel(ctx, p.config, entry.IP.Addr().String(), localPort)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := handle.Close(); err != nil {
			p.log.Debug("close StormDNS tunnel: %v", err)
		}
	}()

	proxyAddr := net.JoinHostPort("127.0.0.1", fmt.Sprint(localPort))
	if err := p.pm.WaitOpen(ctx, proxyAddr, time.Second); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, fmt.Errorf("wait for StormDNS proxy: %w", err)
	}

	latency, err := p.speedtestSvc.MeasureLatency(ctx, speedtest.LatencyConfig{
		Timeout:     p.timeout,
		MaxLatency:  p.timeout,
		DialContext: p.dialSOCKS(proxyAddr),
		URL:         speedtest.GoogleGenerate204HTTP,
	})
	if err != nil {
		return nil, fmt.Errorf("measure latency: %w", err)
	}

	return StormDNSResult{
		Entry:     entry,
		Latency:   latency.RTT,
		QueryType: normalizeQueryType(p.config.DNSQueryType),
		Enc:       p.config.DataEncMethod,
	}, nil
}

// dialSOCKS dials through the embedded client's local no-auth SOCKS5
// listener on each call.
func (p *StormDNSProbe) dialSOCKS(proxyAddr string) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, _, address string) (net.Conn, error) {
		conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", proxyAddr)
		if err != nil {
			return nil, fmt.Errorf("dial StormDNS proxy: %w", err)
		}

		socksConn, err := p.socksService.Connect(ctx, conn, address, socks.Config{})
		if err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("connect SOCKS proxy: %w", err)
		}

		return socksConn, nil
	}
}

// Close releases no shared resources. Each Run cleans up its own tunnel.
func (p *StormDNSProbe) Close() error {
	return nil
}

// normalizeQueryType uppercases a tunnel query type; empty defaults to
// TXT.
func normalizeQueryType(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" {
		return dns.StormDNSQueryTXT
	}
	return s
}
