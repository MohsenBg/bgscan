package slipstreamprobe

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"time"

	"bgscan/internal/core/dns"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/portmgr"
	"bgscan/internal/core/scanner/probe"
	"bgscan/internal/logger"
)

// SlipstreamConfig configures a Slipstream tunnel probe.
type SlipstreamConfig struct {
	// Domain is the DNS domain used by the tunnel.
	Domain string

	// CertPath is an optional TLS certificate path for slipstream-client.
	CertPath string

	// DNSPort is the DNS resolver port.
	DNSPort uint16

	// Timeout limits the end-to-end proxy connectivity check.
	Timeout time.Duration
}

// SlipstreamProbe verifies connectivity through a Slipstream DNS tunnel.
type SlipstreamProbe struct {
	pm             portmgr.Manager
	processTracker probe.ProcessTracker
	config         SlipstreamConfig
	client         dns.SlipstreamClient

	testProxy func(context.Context, string, time.Duration) bool
	waitOpen  func(context.Context, string, time.Duration) error
	now       func() time.Time
}

// Option configures a SlipstreamProbe.
type Option func(*SlipstreamProbe)

// WithProcessTracker uses tracker to manage tunnel processes.
//
// It is useful when multiple probes share a shutdown lifecycle or when tests
// need to inspect registered processes.
func WithProcessTracker(tracker probe.ProcessTracker) Option {
	return func(p *SlipstreamProbe) {
		if tracker != nil {
			p.processTracker = tracker
		}
	}
}

// WithClient uses client to manage Slipstream tunnels.
//
// It is primarily useful for tests. A client stores its active process, so do
// not use this option when concurrent Run calls share the same probe.
func WithClient(client dns.SlipstreamClient) Option {
	return func(p *SlipstreamProbe) {
		if client != nil {
			p.client = client
		}
	}
}

// NewSlipstreamProbe creates a Slipstream tunnel probe.
//
// Init must be called before Run so tracked tunnel processes are terminated
// when the probe's lifecycle context is canceled.
func NewSlipstreamProbe(
	config SlipstreamConfig,
	pm portmgr.Manager,
	opts ...Option,
) (probe.Probe, error) {
	if pm == nil {
		return nil, fmt.Errorf("port manager is nil")
	}

	if config.Domain == "" {
		return nil, fmt.Errorf("slipstream domain is empty")
	}

	if config.Timeout <= 0 {
		config.Timeout = 5 * time.Second
	}

	p := &SlipstreamProbe{
		pm:             pm,
		processTracker: probe.NewProcessTracker(),
		config:         config,
		testProxy:      dns.TestProxy,
		waitOpen:       portmgr.WaitOpen,
		now:            time.Now,
	}

	for _, opt := range opts {
		opt(p)
	}

	if p.client == nil {
		client, err := dns.NewSlipstreamClient(
			config.Domain,
			config.DNSPort,
			config.CertPath,
		)
		if err != nil {
			return nil, fmt.Errorf("create Slipstream client: %w", err)
		}

		p.client = client
	}

	return p, nil
}

// Schema returns the result schema emitted by the probe.
func (s *SlipstreamProbe) Schema() result.ResultSchema {
	return Schema
}

// Init starts process tracking for the probe.
func (s *SlipstreamProbe) Init(ctx context.Context) error {
	s.processTracker.Start(ctx)

	return nil
}

// Run opens a Slipstream tunnel to ip and verifies its local SOCKS5 listener.
func (s *SlipstreamProbe) Run(ctx context.Context, ip netip.Addr) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	localPort, err := s.pm.Get(ctx)
	if err != nil {
		return nil, err
	}
	defer s.pm.Release(localPort)

	proc, err := s.client.RunTunnel(ctx, ip.String(), localPort)
	if err != nil {
		return nil, fmt.Errorf("start Slipstream tunnel: %w", err)
	}

	id, err := s.processTracker.Register(ctx, proc)
	if err != nil {
		return nil, fmt.Errorf("register Slipstream process: %w", err)
	}

	defer func() {
		_ = proc.StopGracefully(time.Second)

		cleanupCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if err := s.processTracker.Unregister(cleanupCtx, id); err != nil {
			logger.CoreError("unregister DNSTT process %s: %s", id, err)
		}
	}()

	addr := net.JoinHostPort("127.0.0.1", fmt.Sprint(localPort))

	if err := s.waitOpen(ctx, addr, time.Second); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}

		return nil, fmt.Errorf("wait for Slipstream proxy: %w", err)
	}

	start := s.now()

	if ok := s.testProxy(ctx, addr, s.config.Timeout); !ok {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("slipstream tunnel connectivity failed for %s", ip)
	}

	return SlipstreamResult{
		IP:      ip,
		Latency: s.now().Sub(start),
		Port:    localPort,
	}, nil
}

// Close releases no shared resources. Each Run cleans up its own tunnel.
func (s *SlipstreamProbe) Close() error {
	return nil
}
