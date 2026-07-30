package dnsttprobe

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

// DNSTTConfig configures a DNSTT tunnel probe.
type DNSTTConfig struct {
	// Domain is the DNSTT server domain.
	Domain string

	// PubKey is the server's public encryption key.
	PubKey string

	// Transport selects the DNS transport used by the tunnel.
	Transport dns.Transport

	// DNSPort is the DNS resolver port.
	DNSPort uint16

	// Timeout limits the end-to-end proxy connectivity check.
	Timeout time.Duration
}

// DNSTTProbe verifies connectivity through a DNSTT tunnel.
type DNSTTProbe struct {
	pm             portmgr.Manager
	processTracker probe.ProcessTracker
	config         DNSTTConfig
	client         dns.DNSTTClient

	testProxy func(context.Context, string, time.Duration) bool
	waitOpen  func(context.Context, string, time.Duration) error
	now       func() time.Time
}

// Option configures a DNSTTProbe.
type Option func(*DNSTTProbe)

// WithProcessTracker uses tracker to register tunnel processes.
//
// It is useful when several probes must share the same shutdown lifecycle or
// when tests need to inspect process registration.
func WithProcessTracker(tracker probe.ProcessTracker) Option {
	return func(p *DNSTTProbe) {
		if tracker != nil {
			p.processTracker = tracker
		}
	}
}

// WithClient uses client to create and stop DNSTT tunnels.
//
// It is primarily useful in tests.
func WithClient(client dns.DNSTTClient) Option {
	return func(p *DNSTTProbe) {
		if client != nil {
			p.client = client
		}
	}
}

// NewDNSTTProbe creates a DNSTT probe.
//
// The caller must invoke Init before Run so the process tracker can terminate
// registered tunnel processes during shutdown.
func NewDNSTTProbe(
	config DNSTTConfig,
	pm portmgr.Manager,
	opts ...Option,
) (probe.Probe, error) {
	if pm == nil {
		return nil, fmt.Errorf("port manager is nil")
	}

	if config.Domain == "" {
		return nil, fmt.Errorf("DNSTT domain is empty")
	}

	if config.Timeout <= 0 {
		config.Timeout = 5 * time.Second
	}

	p := &DNSTTProbe{
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
		client, err := dns.NewDNSTTClient(
			config.Domain,
			config.PubKey,
			config.Transport,
			config.DNSPort,
		)
		if err != nil {
			return nil, fmt.Errorf("create DNSTT client: %w", err)
		}

		p.client = client
	}

	return p, nil
}

// Schema returns the result schema emitted by the probe.
func (d *DNSTTProbe) Schema() result.ResultSchema {
	return Schema
}

// Init starts process tracking for the probe.
func (d *DNSTTProbe) Init(ctx context.Context) error {
	d.processTracker.Start(ctx)

	return nil
}

// Run opens a DNSTT tunnel to ip and verifies connectivity through its local
// SOCKS5 listener.
func (d *DNSTTProbe) Run(ctx context.Context, ip netip.Addr) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	localPort, err := d.pm.Get(ctx)
	if err != nil {
		return nil, err
	}
	defer d.pm.Release(localPort)

	proc, err := d.client.RunTunnel(ctx, ip.String(), localPort)
	if err != nil {
		return nil, fmt.Errorf("start DNSTT tunnel: %w", err)
	}

	id, err := d.processTracker.Register(ctx, proc)
	if err != nil {
		return nil, fmt.Errorf("register DNSTT process: %w", err)
	}

	defer func() {
		_ = proc.StopGracefully(time.Second)

		cleanupCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if err := d.processTracker.Unregister(cleanupCtx, id); err != nil {
			logger.CoreError("unregister DNSTT process %s: %s", id, err)
		}
	}()

	addr := net.JoinHostPort("127.0.0.1", fmt.Sprint(localPort))

	if err := d.waitOpen(ctx, addr, time.Second); err != nil {
		return nil, fmt.Errorf("wait for DNSTT proxy: %w", err)
	}

	start := d.now()

	if ok := d.testProxy(ctx, addr, d.config.Timeout); !ok {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("DNSTT tunnel connectivity failed for %s", ip)
	}

	return DNSTTResult{
		IP:        ip,
		Latency:   d.now().Sub(start),
		Transport: d.config.Transport,
		Port:      localPort,
	}, nil
}

func (d *DNSTTProbe) Close() error {
	return nil
}
