package slipstreamprobe

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"time"

	"bgscan/internal/core/dns"
	"bgscan/internal/core/process"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/portmgr"
	"bgscan/internal/core/scanner/probe"
	"bgscan/internal/core/socks"
	"bgscan/internal/core/speedtest"
	"bgscan/internal/core/ssh"
	"bgscan/internal/logger"
)

// SlipstreamProbe verifies connectivity through a Slipstream DNS tunnel.
type SlipstreamProbe struct {
	pm             portmgr.Manager
	processTracker process.ProcessTracker
	config         dns.SlipstreamConfig
	slipstreamSvc  dns.SlipstreamService
	sshService     ssh.SSHService
	socksService   socks.Service
	speedtestSvc   speedtest.Service
	timeout        time.Duration
}

type Option func(*SlipstreamProbe)

func WithSlipstreamService(service dns.SlipstreamService) Option {
	return func(p *SlipstreamProbe) {
		if service != nil {
			p.slipstreamSvc = service
		}
	}
}

func WithSSHService(service ssh.SSHService) Option {
	return func(p *SlipstreamProbe) {
		if service != nil {
			p.sshService = service
		}
	}
}

func WithSocksService(service socks.Service) Option {
	return func(p *SlipstreamProbe) {
		if service != nil {
			p.socksService = service
		}
	}
}

func WithSpeedtestService(service speedtest.Service) Option {
	return func(p *SlipstreamProbe) {
		if service != nil {
			p.speedtestSvc = service
		}
	}
}

func WithProcessTracker(tracker process.ProcessTracker) Option {
	return func(p *SlipstreamProbe) {
		if tracker != nil {
			p.processTracker = tracker
		}
	}
}

// NewSlipstreamProbe creates a Slipstream tunnel probe.
func NewSlipstreamProbe(
	config dns.SlipstreamConfig,
	timeout time.Duration,
	pm portmgr.Manager,
	opts ...Option,
) (probe.Probe, error) {
	if pm == nil {
		return nil, fmt.Errorf("port manager is nil")
	}

	if errs := config.Validate(); len(errs) != 0 {
		var joined error
		for field, err := range errs {
			joined = errors.Join(joined, fmt.Errorf("%s: %w", field, err))
		}
		return nil, joined
	}

	p := &SlipstreamProbe{
		pm:             pm,
		config:         config,
		processTracker: process.NewProcessTracker(),
		timeout:        timeout,
	}

	for _, opt := range opts {
		opt(p)
	}

	if p.slipstreamSvc == nil {
		svc, err := dns.NewSlipstreamService()
		if err != nil {
			return nil, fmt.Errorf("create Slipstream service: %w", err)
		}
		p.slipstreamSvc = svc
	}
	if p.sshService == nil {
		p.sshService = ssh.NewSSHService()
	}
	if p.socksService == nil {
		p.socksService = socks.NewService()
	}
	if p.speedtestSvc == nil {
		p.speedtestSvc = speedtest.NewService()
	}

	return p, nil
}

// Schema returns the result schema emitted by the probe.
func (s *SlipstreamProbe) Schema() result.ResultSchema {
	return Schema
}

// Init initializes the probe.
func (s *SlipstreamProbe) Init(ctx context.Context) error {
	s.processTracker.Start(ctx)
	return nil
}

// Run opens a Slipstream tunnel to ip (via the external slipstream binary)
// and verifies connectivity through its local proxy listener.
func (s *SlipstreamProbe) Run(ctx context.Context, ip netip.Addr) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Allocate a local port for the Slipstream proxy listener.
	localPort, err := s.pm.Get(ctx)
	if err != nil {
		return nil, err
	}
	defer s.pm.Release(localPort)

	// Determine resolver IP from the target IP.
	resolverIP := ip.String()

	// Start the external Slipstream tunnel process.
	proc, err := s.slipstreamSvc.RunTunnel(ctx, s.config, resolverIP, localPort)
	if err != nil {
		return nil, fmt.Errorf("start Slipstream tunnel: %w", err)
	}

	// Track the process for cleanup.
	id, err := s.processTracker.Register(ctx, proc)
	if err != nil {
		_ = proc.StopGracefully(time.Second)
		return nil, fmt.Errorf("register Slipstream process: %w", err)
	}

	// Ensure the external process and its tracker entry are always cleaned up.
	defer func() {
		_ = proc.StopGracefully(time.Second)

		cleanupCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if err := s.processTracker.Unregister(cleanupCtx, id); err != nil {
			logger.CoreError("unregister Slipstream process %s: %s", id, err)
		}
	}()

	// Wait for the local proxy listener to come up.
	proxyAddr := net.JoinHostPort("127.0.0.1", fmt.Sprint(localPort))
	if err := s.pm.WaitOpen(ctx, proxyAddr, time.Second); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, fmt.Errorf("wait for Slipstream proxy: %w", err)
	}

	dialContext, err := s.buildDialer(proxyAddr)
	if err != nil {
		return nil, err
	}

	// Use speedtest to measure latency through the tunnel.
	latency, err := s.speedtestSvc.MeasureLatency(ctx, speedtest.LatencyConfig{
		Timeout:     s.timeout,
		MaxLatency:  s.timeout,
		DialContext: dialContext,
		URL:         speedtest.GoogleGenerate204HTTP,
	})
	if err != nil {
		return nil, fmt.Errorf("measure latency: %w", err)
	}

	return SlipstreamResult{
		IP:                ip,
		Latency:           latency.RTT,
		Port:              localPort,
		AuthMethod:        s.config.AuthMethod,
		ResolverProxyType: s.config.ProxyType,
	}, nil
}

// buildDialer returns a DialContext that proxies connections through the
// local Slipstream listener, using either SSH or SOCKS depending on config.
func (s *SlipstreamProbe) buildDialer(proxyAddr string) (func(context.Context, string, string) (net.Conn, error), error) {
	switch s.config.ProxyType {
	case dns.ResolverProxySSH:
		return s.dialSSH(proxyAddr)
	case dns.ResolverProxySOCKS:
		return s.dialSOCKS(proxyAddr)
	default:
		return nil, fmt.Errorf("unsupported proxy type: %v", s.config.ProxyType)
	}
}

// dialSSH authenticates a fresh SSH session over the local Slipstream
// listener on each call and proxies the dial through it.
func (s *SlipstreamProbe) dialSSH(proxyAddr string) (func(context.Context, string, string) (net.Conn, error), error) {
	if s.config.AuthMethod == dns.AuthNone {
		return nil, errors.New("SSH authentication is required")
	}

	auth := ssh.SSHConfig{
		Password: s.config.Password,
		User:     s.config.Username,
	}
	if s.config.AuthMethod == dns.AuthKey {
		auth = ssh.SSHConfig{PrivateKey: s.config.PrivateKey}
	}

	return func(ctx context.Context, network, address string) (net.Conn, error) {
		conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", proxyAddr)
		if err != nil {
			return nil, fmt.Errorf("dial Slipstream proxy: %w", err)
		}

		sshPort := s.config.ProxyPort
		if s.config.ProxyPort == 0 {
			sshPort = 22
		}

		addr := net.JoinHostPort(s.config.Domain, strconv.Itoa(int(sshPort)))
		client, err := s.sshService.Connect(ctx, conn, addr, auth)
		if err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("connect SSH proxy: %w", err)
		}

		return s.sshService.SSHDialContext(client)(ctx, network, address)
	}, nil
}

// dialSOCKS authenticates a fresh SOCKS session over the local Slipstream
// listener on each call.
func (s *SlipstreamProbe) dialSOCKS(proxyAddr string) (func(context.Context, string, string) (net.Conn, error), error) {
	if s.config.AuthMethod == dns.AuthKey {
		return nil, errors.New("SOCKS proxy does not support key authentication")
	}

	socksConfig := socks.Config{
		User:     s.config.Username,
		Password: s.config.Password,
	}

	if s.config.AuthMethod == dns.AuthNone {
		socksConfig.User = ""
	}

	return func(ctx context.Context, network, address string) (net.Conn, error) {
		conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", proxyAddr)
		if err != nil {
			return nil, fmt.Errorf("dial Slipstream proxy: %w", err)
		}

		socksConn, err := s.socksService.Connect(ctx, conn, address, socksConfig)
		if err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("connect SOCKS proxy: %w", err)
		}

		return socksConn, nil
	}, nil
}

// Close releases no shared resources. Each Run cleans up its own tunnel.
func (s *SlipstreamProbe) Close() error {
	return nil
}
