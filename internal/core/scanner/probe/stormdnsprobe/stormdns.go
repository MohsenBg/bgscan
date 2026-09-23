package stormdnsprobe

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner/portmgr"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/core/socks"
	"github.com/MohsenBg/bgscan/internal/core/speedtest"
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

// Run boots the embedded StormDNS client with ip as its DNS resolver
// and measures latency through its local SOCKS5 listener, retrying
// failed attempts up to the configured tries.
func (p *StormDNSProbe) Run(ctx context.Context, ip netip.Addr) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	localPort, err := p.pm.Get(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%w: lease port for StormDNS: %w", probe.ErrEnvironment, err)
	}
	defer p.pm.Release(localPort)

	tries := p.tries
	if tries < 1 {
		tries = 1
	}

	for attempt := 0; attempt < tries; attempt++ {
		var res result.Result
		if res, err = p.runOnce(ctx, ip, localPort); err == nil {
			return res, nil
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
	}

	return nil, err
}

func (p *StormDNSProbe) runOnce(ctx context.Context, ip netip.Addr, localPort uint16) (result.Result, error) {
	// RunTunnel blocks until the tunnel session is ready.
	handle, err := p.stormDNSService.RunTunnel(ctx, p.config, ip.String(), localPort)
	if err != nil {
		return nil, fmt.Errorf("start StormDNS tunnel: %w", err)
	}
	defer func() {
		if err := handle.Close(); err != nil {
			p.log.Error("close StormDNS tunnel: %v", err)
		}
	}()

	proxyAddr := net.JoinHostPort("127.0.0.1", fmt.Sprint(localPort))
	if err := p.pm.WaitOpen(ctx, proxyAddr, time.Second); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, fmt.Errorf("%w: wait for StormDNS proxy: %w", probe.ErrEnvironment, err)
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
		IP:        ip,
		Latency:   latency.RTT,
		Port:      p.config.ResolverPort,
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
