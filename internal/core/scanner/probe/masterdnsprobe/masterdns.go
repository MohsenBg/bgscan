package masterdnsprobe

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner/portmgr"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/core/socks"
	"github.com/MohsenBg/bgscan/internal/core/speedtest"
	"github.com/MohsenBg/bgscan/internal/logger"
)

// MasterDNSProbe verifies connectivity through a MasterDNS tunnel.
//
// The embedded MasterDNS client exposes a local no-auth SOCKS5 listener;
// the probe boots it against the scanned IP as resolver and measures
// latency through that listener.
type MasterDNSProbe struct {
	pm               portmgr.Manager
	config           dns.MasterDNSConfig
	masterDNSService dns.MasterDNSService
	socksService     socks.Service
	speedtestService speedtest.Service
	timeout          time.Duration
}

type Option func(*MasterDNSProbe)

func WithMasterDNSService(service dns.MasterDNSService) Option {
	return func(p *MasterDNSProbe) {
		if service != nil {
			p.masterDNSService = service
		}
	}
}

func WithSocksService(service socks.Service) Option {
	return func(p *MasterDNSProbe) {
		if service != nil {
			p.socksService = service
		}
	}
}

func WithSpeedtestService(service speedtest.Service) Option {
	return func(p *MasterDNSProbe) {
		if service != nil {
			p.speedtestService = service
		}
	}
}

// NewMasterDNSProbe creates a MasterDNS probe.
func NewMasterDNSProbe(
	config dns.MasterDNSConfig,
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

	p := &MasterDNSProbe{
		pm:      pm,
		config:  config,
		timeout: timeout,
	}

	for _, opt := range opts {
		opt(p)
	}

	if p.masterDNSService == nil {
		p.masterDNSService = dns.NewMasterDNSService()
	}
	if p.socksService == nil {
		p.socksService = socks.NewService()
	}
	if p.speedtestService == nil {
		p.speedtestService = speedtest.NewService()
	}

	return p, nil
}

// Schema returns the result schema emitted by the probe.
func (p *MasterDNSProbe) Schema() result.ResultSchema {
	return Schema
}

// Init initializes the probe.
func (p *MasterDNSProbe) Init(context.Context) error {
	return nil
}

// Run boots the embedded MasterDNS client with ip as its DNS resolver
// and measures latency through its local SOCKS5 listener.
func (p *MasterDNSProbe) Run(ctx context.Context, ip netip.Addr) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	localPort, err := p.pm.Get(ctx)
	if err != nil {
		return nil, err
	}
	defer p.pm.Release(localPort)

	// RunTunnel blocks until the tunnel session is ready.
	handle, err := p.masterDNSService.RunTunnel(ctx, p.config, ip.String(), localPort)
	if err != nil {
		return nil, fmt.Errorf("start MasterDNS tunnel: %w", err)
	}
	defer func() {
		if err := handle.Close(); err != nil {
			logger.CoreError("close MasterDNS tunnel: %v", err)
		}
	}()

	proxyAddr := net.JoinHostPort("127.0.0.1", fmt.Sprint(localPort))
	if err := p.pm.WaitOpen(ctx, proxyAddr, time.Second); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, fmt.Errorf("wait for MasterDNS proxy: %w", err)
	}

	latency, err := p.speedtestService.MeasureLatency(ctx, speedtest.LatencyConfig{
		Timeout:     p.timeout,
		MaxLatency:  p.timeout,
		DialContext: p.dialSOCKS(proxyAddr),
		URL:         speedtest.GoogleGenerate204HTTP,
	})
	if err != nil {
		return nil, fmt.Errorf("measure latency: %w", err)
	}

	return MasterDNSResult{
		IP:      ip,
		Latency: latency.RTT,
		Port:    localPort,
		Enc:     p.config.DataEncMethod,
	}, nil
}

// dialSOCKS dials through the embedded client's local no-auth SOCKS5
// listener on each call.
func (p *MasterDNSProbe) dialSOCKS(proxyAddr string) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, _, address string) (net.Conn, error) {
		conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", proxyAddr)
		if err != nil {
			return nil, fmt.Errorf("dial MasterDNS proxy: %w", err)
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
func (p *MasterDNSProbe) Close() error {
	return nil
}
