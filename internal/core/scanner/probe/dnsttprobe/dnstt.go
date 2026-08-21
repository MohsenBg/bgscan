package dnsttprobe

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

// DNSTTProbe verifies connectivity through a DNSTT tunnel.
type DNSTTProbe struct {
	cfg              dns.DNSTTConfig
	dnsttService     dns.DNSTTService
	sshService       ssh.SSHService
	socksService     socks.Service
	speedtestService speedtest.Service
	timeout          time.Duration
}

type Option func(*DNSTTProbe)

func WithDNSTTService(service dns.DNSTTService) Option {
	return func(p *DNSTTProbe) {
		if service != nil {
			p.dnsttService = service
		}
	}
}

func WithSSHService(service ssh.SSHService) Option {
	return func(p *DNSTTProbe) {
		if service != nil {
			p.sshService = service
		}
	}
}

func WithSocksService(service socks.Service) Option {
	return func(p *DNSTTProbe) {
		if service != nil {
			p.socksService = service
		}
	}
}

func WithSpeedtestService(service speedtest.Service) Option {
	return func(p *DNSTTProbe) {
		if service != nil {
			p.speedtestService = service
		}
	}
}

// NewDNSTTProbe creates a DNSTT probe.
func NewDNSTTProbe(
	config dns.DNSTTConfig,
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

	p := &DNSTTProbe{
		cfg:     config,
		timeout: timeout,
	}

	for _, opt := range opts {
		opt(p)
	}

	if p.dnsttService == nil {
		p.dnsttService = dns.NewDNSTTService()
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
func (d *DNSTTProbe) Schema() result.ResultSchema {
	return Schema
}

// Init initializes the probe.
func (d *DNSTTProbe) Init(context.Context) error {
	return nil
}

// Run tests DNSTT connectivity for ip.
func (d *DNSTTProbe) Run(ctx context.Context, ip netip.Addr) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	cfg := d.cfg
	tunnel, err := d.dnsttService.NewTunnel(ctx, cfg, ip)
	if err != nil {
		return nil, fmt.Errorf("create DNSTT tunnel: %w", err)
	}
	defer func() {
		err := tunnel.Close()
		if err != nil {
			logger.CoreError("close DNSTT tunnel: %v", err)
		}
	}()

	var dialContext func(context.Context, string, string) (net.Conn, error)

	switch d.cfg.ProxyType {
	case dns.ResolverProxySSH:
		dialContext, err = d.dialSSH(ctx, tunnel)
	case dns.ResolverProxySOCKS:
		dialContext, err = d.dialSOCKS(tunnel)
	default:
		err = fmt.Errorf("unsupported proxy type: %v", d.cfg.ProxyType)
	}
	if err != nil {
		return nil, err
	}

	return d.measureLatency(ctx, ip, dialContext)
}

// dialSSH authenticates an SSH session over tunnel and returns a
// DialContext that proxies connections through it.
func (d *DNSTTProbe) dialSSH(ctx context.Context, tunnel net.Conn) (func(context.Context, string, string) (net.Conn, error), error) {
	if d.cfg.AuthMethod == dns.AuthNone {
		return nil, errors.New("SSH authentication is required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	port := d.cfg.ProxyPort
	if port == 0 {
		port = 22
	}
	addr := net.JoinHostPort(d.cfg.Domain, strconv.Itoa(int(port)))

	auth := ssh.SSHConfig{
		Password: d.cfg.Password,
		User:     d.cfg.Username,
	}
	if d.cfg.AuthMethod == dns.AuthKey {
		auth = ssh.SSHConfig{PrivateKey: d.cfg.PrivateKey}
	}

	type sshResult struct {
		client *ssh.Client
		err    error
	}
	resultCh := make(chan sshResult, 1)
	go func() {
		client, err := d.sshService.Connect(ctx, tunnel, addr, auth)
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
		return d.sshService.SSHDialContext(res.client), nil
	}
}

// dialSOCKS returns a DialContext that performs the SOCKS5 handshake over
// tunnel using the real destination address requested by the caller.
func (d *DNSTTProbe) dialSOCKS(tunnel net.Conn) (func(context.Context, string, string) (net.Conn, error), error) {
	if d.cfg.AuthMethod == dns.AuthKey {
		return nil, errors.New("SOCKS proxy does not support key authentication")
	}

	socksConfig := socks.Config{
		Password: d.cfg.Password,
		User:     d.cfg.Username,
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
			conn, err := d.socksService.Connect(ctx, tunnel, address, socksConfig)
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

func (d *DNSTTProbe) measureLatency(
	ctx context.Context,
	ip netip.Addr,
	dialContext func(context.Context, string, string) (net.Conn, error),
) (result.Result, error) {
	latency, err := d.speedtestService.MeasureLatency(ctx, speedtest.LatencyConfig{
		ProxyPort:   0,
		Timeout:     d.timeout,
		MaxLatency:  d.timeout,
		DialContext: dialContext,
		URL:         speedtest.GoogleGenerate204HTTP,
	})
	if err != nil {
		return nil, fmt.Errorf("measure latency: %w", err)
	}

	return DNSTTResult{
		IP:        ip,
		Latency:   latency.RTT,
		Transport: d.cfg.ResolverType,
		Port:      d.cfg.ResolverPort,
	}, nil
}

// Close releases resources owned by the probe.
func (v *DNSTTProbe) Close() error {
	return nil
}
