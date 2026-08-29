package vaydnsprobe

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"time"

	"bgscan/internal/core/dns"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/probe"
	"bgscan/internal/core/socks"
	"bgscan/internal/core/speedtest"
	"bgscan/internal/core/ssh"
	"bgscan/internal/logger"
)

// VayDNSProbe verifies connectivity through a VayDNS tunnel.
type VayDNSProbe struct {
	cfg              dns.VayDNSConfig
	vaydnsService    dns.VayDNSService
	sshService       ssh.SSHService
	socksService     socks.Service
	speedtestService speedtest.Service
	timeout          time.Duration
}

type Option func(*VayDNSProbe)

func WithVayDNSService(service dns.VayDNSService) Option {
	return func(p *VayDNSProbe) {
		if service != nil {
			p.vaydnsService = service
		}
	}
}

func WithSSHService(service ssh.SSHService) Option {
	return func(p *VayDNSProbe) {
		if service != nil {
			p.sshService = service
		}
	}
}

func WithSocksService(service socks.Service) Option {
	return func(p *VayDNSProbe) {
		if service != nil {
			p.socksService = service
		}
	}
}

func WithSpeedtestService(service speedtest.Service) Option {
	return func(p *VayDNSProbe) {
		if service != nil {
			p.speedtestService = service
		}
	}
}

// NewVayDNSProbe creates a VayDNS probe.
func NewVayDNSProbe(
	config dns.VayDNSConfig,
	timeout time.Duration,
	opts ...Option,
) (probe.Probe, error) {
	if errs := config.Validate(); len(errs) != 0 {
		var joined error
		for field, err := range errs {
			joined = errors.Join(joined, fmt.Errorf("%s: %w", field, err))
		}
		return nil, joined
	}

	p := &VayDNSProbe{
		cfg:     config,
		timeout: timeout,
	}

	for _, opt := range opts {
		opt(p)
	}

	if p.vaydnsService == nil {
		p.vaydnsService = dns.NewVayDNSService()
	}
	if p.sshService == nil {
		p.sshService = ssh.NewSSHService()
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
func (v *VayDNSProbe) Schema() result.ResultSchema {
	return Schema
}

// Init initializes the probe.
func (v *VayDNSProbe) Init(context.Context) error {
	return nil
}

// Run tests VayDNS connectivity for ip.
func (v *VayDNSProbe) Run(ctx context.Context, ip netip.Addr) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// The scanned target acts as the DNS resolver, matching the
	// Slipstream and DNSTT probes.
	tunnel, err := v.vaydnsService.NewTunnel(ctx, v.cfg, ip)
	if err != nil {
		return nil, fmt.Errorf("create VayDNS tunnel: %w", err)
	}
	defer func() {
		if err := tunnel.Close(); err != nil {
			logger.CoreError("close VayDNS tunnel: %v", err)
		}
	}()

	var dialContext func(context.Context, string, string) (net.Conn, error)

	switch v.cfg.ProxyType {
	case dns.ResolverProxySSH:
		dialContext, err = v.dialSSH(ctx, tunnel)
	case dns.ResolverProxySOCKS:
		dialContext, err = v.dialSOCKS(tunnel)
	default:
		err = fmt.Errorf("unsupported proxy type: %v", v.cfg.ProxyType)
	}
	if err != nil {
		return nil, err
	}

	return v.measureLatency(ctx, ip, dialContext)
}

// dialSSH authenticates an SSH session over tunnel and returns a
// DialContext that proxies connections through it.
func (v *VayDNSProbe) dialSSH(ctx context.Context, tunnel net.Conn) (func(context.Context, string, string) (net.Conn, error), error) {
	if v.cfg.AuthMethod == dns.AuthNone {
		return nil, errors.New("SSH authentication is required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	port := v.cfg.ProxyPort
	if port == 0 {
		port = 22
	}
	addr := net.JoinHostPort(v.cfg.Domain, strconv.Itoa(int(port)))

	auth := ssh.SSHConfig{
		Password:       v.cfg.Password,
		User:           v.cfg.Username,
		KnownHostsFile: v.cfg.KnownHostsFile,
	}
	if v.cfg.AuthMethod == dns.AuthKey {
		auth = ssh.SSHConfig{PrivateKey: v.cfg.PrivateKey, KnownHostsFile: v.cfg.KnownHostsFile}
	}

	type sshResult struct {
		client *ssh.Client
		err    error
	}
	resultCh := make(chan sshResult, 1)
	go func() {
		client, err := v.sshService.Connect(ctx, tunnel, addr, auth)
		resultCh <- sshResult{client: client, err: err}
	}()

	select {
	case <-ctx.Done():
		_ = tunnel.Close()
		return nil, fmt.Errorf("connect SSH proxy: %w", ctx.Err())
	case res := <-resultCh:
		if res.err != nil {
			return nil, fmt.Errorf("connect SSH proxy: %w", res.err)
		}
		return v.sshService.SSHDialContext(res.client), nil
	}
}

// dialSOCKS returns a DialContext that performs the SOCKS5 handshake over
// tunnel using the real destination address requested by the caller.
func (v *VayDNSProbe) dialSOCKS(tunnel net.Conn) (func(context.Context, string, string) (net.Conn, error), error) {
	if v.cfg.AuthMethod == dns.AuthKey {
		return nil, errors.New("SOCKS proxy does not support key authentication")
	}

	socksConfig := socks.Config{
		Password: v.cfg.Password,
		User:     v.cfg.Username,
	}

	if v.cfg.AuthMethod == dns.AuthNone {
		socksConfig.User = ""
	}

	return func(ctx context.Context, network, address string) (net.Conn, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		type socksResult struct {
			conn net.Conn
			err  error
		}
		resultCh := make(chan socksResult, 1)
		go func() {
			conn, err := v.socksService.Connect(ctx, tunnel, address, socksConfig)
			resultCh <- socksResult{conn: conn, err: err}
		}()

		select {
		case <-ctx.Done():
			_ = tunnel.Close()
			return nil, fmt.Errorf("connect SOCKS proxy: %w", ctx.Err())
		case res := <-resultCh:
			if res.err != nil {
				return nil, fmt.Errorf("connect SOCKS proxy: %w", res.err)
			}
			return res.conn, nil
		}
	}, nil
}

func (v *VayDNSProbe) measureLatency(
	ctx context.Context,
	ip netip.Addr,
	dialContext func(context.Context, string, string) (net.Conn, error),
) (result.Result, error) {
	latency, err := v.speedtestService.MeasureLatency(ctx, speedtest.LatencyConfig{
		ProxyPort:   0,
		Timeout:     v.timeout,
		MaxLatency:  v.timeout,
		DialContext: dialContext,
		URL:         speedtest.GoogleGenerate204HTTP,
	})
	if err != nil {
		return nil, fmt.Errorf("measure latency: %w", err)
	}

	return VayDNSResult{
		IP:                ip,
		Latency:           latency.RTT,
		Transport:         v.cfg.ResolverType,
		Port:              v.cfg.ResolverPort,
		AuthMethod:        v.cfg.AuthMethod,
		ResolverProxyType: v.cfg.ProxyType,
	}, nil
}

// Close releases resources owned by the probe.
func (v *VayDNSProbe) Close() error {
	return nil
}
