package slipstreamprobe

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/netutil"
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner/portmgr"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/core/socks"
	"github.com/MohsenBg/bgscan/internal/core/speedtest"
	"github.com/MohsenBg/bgscan/internal/core/ssh"
	"github.com/MohsenBg/bgscan/internal/logger"
)

// SlipstreamProbe verifies connectivity through a Slipstream DNS tunnel.
type SlipstreamProbe struct {
	pm            portmgr.Manager
	config        dns.SlipstreamConfig
	slipstreamSvc dns.SlipstreamService
	sshService    ssh.SSHService
	socksService  socks.Service
	speedtestSvc  speedtest.Service
	timeout       time.Duration
	tries         int
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

// WithTries sets how many times a failed probe is retried.
// Values below 1 are ignored; the default is a single attempt.
func WithTries(n int) Option {
	return func(p *SlipstreamProbe) {
		if n >= 1 {
			p.tries = n
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
		pm:      pm,
		config:  config,
		timeout: timeout,
		tries:   1,
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
	return nil
}

// Run opens a Slipstream tunnel to ip (via the embedded libslipstream
// library) and verifies connectivity through its local proxy listener,
// retrying failed attempts up to the configured tries.
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

	tries := s.tries
	if tries < 1 {
		tries = 1
	}

	for attempt := 0; attempt < tries; attempt++ {
		var res result.Result
		if res, err = s.runOnce(ctx, ip, localPort); err == nil {
			return res, nil
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
	}

	return nil, err
}

func (s *SlipstreamProbe) runOnce(ctx context.Context, ip netip.Addr, localPort uint16) (result.Result, error) {
	client, err := s.slipstreamSvc.RunTunnel(ctx, s.config, ip, localPort, uint16(s.timeout.Seconds()))
	if err != nil {
		return nil, fmt.Errorf("start Slipstream tunnel: %w", err)
	}
	defer func() {
		if err := client.Stop(); err != nil {
			logger.CoreError("close Slipstream tunnel: %v", err)
		}
	}()

	// Wait for the local proxy listener to come up.
	proxyAddr := net.JoinHostPort(netutil.Loopback(ip).String(), fmt.Sprint(localPort))
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
		Password:       s.config.Password,
		User:           s.config.Username,
		KnownHostsFile: s.config.KnownHostsFile,
	}
	if s.config.AuthMethod == dns.AuthKey {
		auth = ssh.SSHConfig{PrivateKey: s.config.PrivateKey, KnownHostsFile: s.config.KnownHostsFile}
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
