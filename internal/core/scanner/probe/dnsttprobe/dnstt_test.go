package dnsttprobe

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/socks"
	"github.com/MohsenBg/bgscan/internal/core/speedtest"
)

type fakeDNSTTService struct {
	tunnel      net.Conn
	newErr      error
	called      bool
	gotConfig   dns.DNSTTConfig
	gotResolver netip.Addr
}

// EditConfig implements [dns.DNSTTService].
func (f *fakeDNSTTService) EditConfig(config dns.DNSTTConfig, originalName string) error {
	return nil
}

// ValidateAllConfigs implements [dns.DNSTTService].
func (f *fakeDNSTTService) ValidateAllConfigs() ([]dns.ConfigValidationResult, error) {
	return nil, nil
}

// GetAllConfigFiles implements [dns.DNSTTService].
func (f *fakeDNSTTService) GetAllConfigFiles() ([]dns.DNSTTConfigFile, error) {
	return nil, nil
}

func (f *fakeDNSTTService) RenameConfig(oldName, newName string) error {
	return nil
}

// LoadConfig implements [dns.DNSTTService].
func (f *fakeDNSTTService) LoadConfig(name string) (dns.DNSTTConfig, error) {
	return dns.DNSTTConfig{}, nil
}

// SaveConfig implements [dns.DNSTTService].
func (f *fakeDNSTTService) SaveConfig(config dns.DNSTTConfig, name string) error {
	return nil
}

func (f *fakeDNSTTService) NewTunnel(ctx context.Context, cfg dns.DNSTTConfig, resolverAddr netip.Addr) (net.Conn, error) {
	f.called = true
	f.gotConfig = cfg
	f.gotResolver = resolverAddr

	if f.newErr != nil {
		return nil, f.newErr
	}

	return f.tunnel, nil
}

type fakeSOCKSService struct {
	conn       net.Conn
	connectErr error
	called     bool
	gotTunnel  net.Conn
	gotAddr    string
	gotConfig  socks.Config
}

func (f *fakeSOCKSService) Connect(
	ctx context.Context,
	tunnel net.Conn,
	addr string,
	cfg socks.Config,
) (net.Conn, error) {
	f.called = true
	f.gotTunnel = tunnel
	f.gotAddr = addr
	f.gotConfig = cfg

	if f.connectErr != nil {
		return nil, f.connectErr
	}

	return f.conn, nil
}

type fakeSpeedtestService struct {
	result speedtest.LatencyResult
	err    error

	called bool
	cfg    speedtest.LatencyConfig
}

func (f *fakeSpeedtestService) MeasureLatency(
	ctx context.Context,
	cfg speedtest.LatencyConfig,
) (speedtest.LatencyResult, error) {
	f.called = true
	f.cfg = cfg

	// Exercise the dialer so probe-level connection errors surface,
	// mirroring what the real service does.
	if cfg.DialContext != nil {
		conn, dialErr := cfg.DialContext(ctx, "tcp", "127.0.0.1:1080")
		if dialErr == nil && conn != nil {
			_ = conn.Close()
		}
	}

	if f.err != nil {
		return speedtest.LatencyResult{}, f.err
	}

	return f.result, nil
}

func (f *fakeSpeedtestService) MeasureDownloadSpeed(
	context.Context,
	speedtest.DownloadConfig,
) (speedtest.SpeedResult, error) {
	return speedtest.SpeedResult{}, nil
}

func (f *fakeSpeedtestService) MeasureUploadSpeed(
	context.Context,
	speedtest.UploadConfig,
) (speedtest.SpeedResult, error) {
	return speedtest.SpeedResult{}, nil
}

func validConfig() dns.DNSTTConfig {
	return dns.DNSTTConfig{
		Domain:       "tunnel.example.com",
		PubKey:       "cd6d78e954f48f62cb74cdcf8a2459d3d39786a7e11fc4f74c04bca86371f748",
		ResolverType: dns.ResolverTypeUDP,
		ResolverPort: 53,
		ProxyType:    dns.ResolverProxySOCKS,
		ProxyPort:    1080,
		Username:     "user",
		Password:     "password",
		Fingerprint:  "firefox",
	}
}

func testIP() netip.Addr {
	return netip.MustParseAddr("1.2.3.4")
}

func newTestProbe(
	t *testing.T,
	config dns.DNSTTConfig,
	dnstt dns.DNSTTService,
	socksService socks.Service,
	speedtestService speedtest.Service,
	opts ...Option,
) *DNSTTProbe {
	t.Helper()

	opts = append(
		opts,
		WithDNSTTService(dnstt),
		WithSocksService(socksService),
		WithSpeedtestService(speedtestService),
	)

	got, err := NewDNSTTProbe(
		config,
		2*time.Second,
		opts...,
	)
	if err != nil {
		t.Fatal(err)
	}

	p, ok := got.(*DNSTTProbe)
	if !ok {
		t.Fatalf("NewDNSTTProbe returned %T, want *DNSTTProbe", got)
	}

	return p
}

func TestNewDNSTTProbeValidation(t *testing.T) {
	config := validConfig()
	config.Domain = ""

	_, err := NewDNSTTProbe(config, time.Second)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNewDNSTTProbeAppliesOptions(t *testing.T) {
	dnstt := &fakeDNSTTService{}
	socksService := &fakeSOCKSService{}
	speedtestService := &fakeSpeedtestService{}

	got, err := NewDNSTTProbe(
		validConfig(),
		time.Second,
		WithDNSTTService(dnstt),
		WithSocksService(socksService),
		WithSpeedtestService(speedtestService),
	)
	if err != nil {
		t.Fatal(err)
	}

	p := got.(*DNSTTProbe)

	if p.dnsttService != dnstt {
		t.Fatal("DNSTT service option was not applied")
	}

	if p.socksService != socksService {
		t.Fatal("SOCKS service option was not applied")
	}

	if p.speedtestService != speedtestService {
		t.Fatal("speedtest service option was not applied")
	}
}

func TestNewDNSTTProbeIgnoresNilOptions(t *testing.T) {
	dnstt := &fakeDNSTTService{}
	socksService := &fakeSOCKSService{}
	speedtestService := &fakeSpeedtestService{}

	got, err := NewDNSTTProbe(
		validConfig(),
		time.Second,
		WithDNSTTService(nil),
		WithSocksService(nil),
		WithSpeedtestService(nil),
	)
	if err != nil {
		t.Fatal(err)
	}

	p := got.(*DNSTTProbe)

	if p.dnsttService == nil {
		t.Fatal("default DNSTT service was not created")
	}

	if p.sshService == nil {
		t.Fatal("default SSH service was not created")
	}

	if p.socksService == nil {
		t.Fatal("default SOCKS service was not created")
	}

	if p.speedtestService == nil {
		t.Fatal("default speedtest service was not created")
	}

	_ = dnstt
	_ = socksService
	_ = speedtestService
}

func TestRunReturnsCanceledContextBeforeCreatingTunnel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	dnstt := &fakeDNSTTService{}
	p := newTestProbe(
		t,
		validConfig(),
		dnstt,
		&fakeSOCKSService{},
		&fakeSpeedtestService{},
	)

	_, err := p.Run(ctx, testIP())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}

	if dnstt.called {
		t.Fatal("DNSTT tunnel was created after context cancellation")
	}
}

func TestRunReturnsTunnelError(t *testing.T) {
	expectedErr := errors.New("create tunnel failed")

	dnstt := &fakeDNSTTService{
		newErr: expectedErr,
	}

	p := newTestProbe(
		t,
		validConfig(),
		dnstt,
		&fakeSOCKSService{},
		&fakeSpeedtestService{},
	)

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected tunnel error")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}

	if !dnstt.called {
		t.Fatal("DNSTT service was not called")
	}
}

// TestRunPassesTargetIPAsResolverAddr verifies that the scanned target is
// used as the DNS resolver address when creating the tunnel.
func TestRunPassesTargetIPAsResolverAddr(t *testing.T) {
	tunnel, tunnelPeer := net.Pipe()
	defer func() {
		_ = tunnelPeer.Close()
	}()

	dnstt := &fakeDNSTTService{
		tunnel: tunnel,
	}

	p := newTestProbe(
		t,
		validConfig(),
		dnstt,
		&fakeSOCKSService{},
		&fakeSpeedtestService{result: speedtest.LatencyResult{RTT: time.Millisecond}},
	)

	if _, err := p.Run(context.Background(), testIP()); err != nil {
		t.Fatal(err)
	}

	if !dnstt.called {
		t.Fatal("DNSTT service was not called")
	}

	want := netip.MustParseAddr("1.2.3.4")
	if dnstt.gotResolver != want {
		t.Fatalf(
			"resolverAddr = %s, want %s",
			dnstt.gotResolver,
			want,
		)
	}
}

func TestRunReturnsUnsupportedProxyType(t *testing.T) {
	config := validConfig()
	config.ProxyType = dns.ResolverProxyType("")

	dnsttConn, peer := net.Pipe()
	defer func() {
		_ = peer.Close()
	}()

	dnstt := &fakeDNSTTService{
		tunnel: dnsttConn,
	}

	p := newTestProbe(
		t,
		config,
		dnstt,
		&fakeSOCKSService{},
		&fakeSpeedtestService{},
	)

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected unsupported proxy error")
	}
}

func TestRunSSHRequiresAuthentication(t *testing.T) {
	config := validConfig()
	config.ProxyType = dns.ResolverProxySSH
	config.AuthMethod = dns.AuthNone

	dnsttConn, peer := net.Pipe()
	defer func() {
		_ = peer.Close()
	}()

	_, err := NewDNSTTProbe(
		config,
		2*time.Second,
		WithDNSTTService(&fakeDNSTTService{tunnel: dnsttConn}),
		WithSocksService(&fakeSOCKSService{}),
		WithSpeedtestService(&fakeSpeedtestService{}),
	)
	if err == nil {
		t.Fatal("expected SSH authentication error from constructor")
	}
}

func TestRunSOCKSRejectsKeyAuthentication(t *testing.T) {
	config := validConfig()
	config.ProxyType = dns.ResolverProxySOCKS
	config.AuthMethod = dns.AuthKey

	dnsttConn, peer := net.Pipe()
	defer func() {
		_ = peer.Close()
	}()

	_, err := NewDNSTTProbe(
		config,
		2*time.Second,
		WithDNSTTService(&fakeDNSTTService{tunnel: dnsttConn}),
		WithSocksService(&fakeSOCKSService{}),
		WithSpeedtestService(&fakeSpeedtestService{}),
	)
	if err == nil {
		t.Fatal("expected SOCKS authentication error from constructor")
	}
}

func TestRunSOCKSConnectError(t *testing.T) {
	expectedErr := errors.New("SOCKS connection failed")

	config := validConfig()
	config.ProxyType = dns.ResolverProxySOCKS
	config.AuthMethod = dns.AuthNone

	tunnel, peer := net.Pipe()
	defer func() {
		_ = peer.Close()
	}()

	dnstt := &fakeDNSTTService{
		tunnel: tunnel,
	}

	socksService := &fakeSOCKSService{
		connectErr: expectedErr,
	}

	speedtestService := &fakeSpeedtestService{
		err: expectedErr,
	}

	p := newTestProbe(
		t,
		config,
		dnstt,
		socksService,
		speedtestService,
	)

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected SOCKS connection error")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}

	if !socksService.called {
		t.Fatal("SOCKS service was not called")
	}

	if socksService.gotAddr != "127.0.0.1:1080" {
		t.Fatalf(
			"SOCKS address = %q, want 127.0.0.1:1080",
			socksService.gotAddr,
		)
	}
}

func TestRunSOCKSSuccess(t *testing.T) {
	config := validConfig()
	config.ProxyType = dns.ResolverProxySOCKS
	config.AuthMethod = dns.AuthPassword

	tunnel, tunnelPeer := net.Pipe()
	defer func() {
		_ = tunnelPeer.Close()
	}()

	proxyConn, proxyPeer := net.Pipe()
	defer func() {
		_ = proxyPeer.Close()
	}()

	dnstt := &fakeDNSTTService{
		tunnel: tunnel,
	}

	socksService := &fakeSOCKSService{
		conn: proxyConn,
	}

	speedtestService := &fakeSpeedtestService{
		result: speedtest.LatencyResult{
			RTT:        120 * time.Millisecond,
			MaxLatency: 2 * time.Second,
		},
	}

	p := newTestProbe(
		t,
		config,
		dnstt,
		socksService,
		speedtestService,
	)

	gotResult, err := p.Run(context.Background(), testIP())
	if err != nil {
		t.Fatal(err)
	}

	got, ok := gotResult.(DNSTTResult)
	if !ok {
		t.Fatalf("result type = %T, want DNSTTResult", gotResult)
	}

	if got.IP != testIP() {
		t.Fatalf("IP = %s, want %s", got.IP, testIP())
	}

	if got.Latency != 120*time.Millisecond {
		t.Fatalf(
			"latency = %s, want 120ms",
			got.Latency,
		)
	}

	if got.Transport != config.ResolverType {
		t.Fatalf(
			"transport = %v, want %v",
			got.Transport,
			config.ResolverType,
		)
	}

	if got.Port != config.ResolverPort {
		t.Fatalf(
			"port = %d, want %d",
			got.Port,
			config.ResolverPort,
		)
	}

	if !socksService.called {
		t.Fatal("SOCKS service was not called")
	}

	if socksService.gotAddr != "127.0.0.1:1080" {
		t.Fatalf(
			"SOCKS address = %q, want 127.0.0.1:1080",
			socksService.gotAddr,
		)
	}

	if socksService.gotConfig.User != config.Username {
		t.Fatalf(
			"username = %q, want %q",
			socksService.gotConfig.User,
			config.Username,
		)
	}

	if socksService.gotConfig.Password != config.Password {
		t.Fatalf(
			"password = %q, want %q",
			socksService.gotConfig.Password,
			config.Password,
		)
	}

	if !speedtestService.called {
		t.Fatal("speedtest service was not called")
	}

	if speedtestService.cfg.Timeout != 2*time.Second {
		t.Fatalf(
			"latency timeout = %s, want 2s",
			speedtestService.cfg.Timeout,
		)
	}

	if speedtestService.cfg.MaxLatency != 2*time.Second {
		t.Fatalf(
			"max latency = %s, want 2s",
			speedtestService.cfg.MaxLatency,
		)
	}

	if speedtestService.cfg.ProxyPort != 0 {
		t.Fatalf(
			"proxy port = %d, want 0",
			speedtestService.cfg.ProxyPort,
		)
	}

	if speedtestService.cfg.DialContext == nil {
		t.Fatal("DialContext is nil")
	}

	conn, err := speedtestService.cfg.DialContext(
		context.Background(),
		"tcp",
		"example.com:443",
	)
	if err != nil {
		t.Fatalf("DialContext() error = %v", err)
	}

	if conn != proxyConn {
		t.Fatal("DialContext returned unexpected connection")
	}
}

func TestRunReturnsLatencyError(t *testing.T) {
	expectedErr := errors.New("latency failed")

	config := validConfig()
	config.ProxyType = dns.ResolverProxySOCKS
	config.AuthMethod = dns.AuthNone

	tunnel, tunnelPeer := net.Pipe()
	defer func() {
		_ = tunnelPeer.Close()
	}()

	proxyConn, proxyPeer := net.Pipe()
	defer func() {
		_ = proxyPeer.Close()
	}()

	speedtestService := &fakeSpeedtestService{
		err: expectedErr,
	}

	p := newTestProbe(
		t,
		config,
		&fakeDNSTTService{tunnel: tunnel},
		&fakeSOCKSService{conn: proxyConn},
		speedtestService,
	)

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected latency error")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
}

func TestClose(t *testing.T) {
	p := &DNSTTProbe{}

	if err := p.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
