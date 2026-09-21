package slipstreamprobe

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	ffislipstream "github.com/MohsenBg/bgscan/internal/core/ffi/slipstream"
	"github.com/MohsenBg/bgscan/internal/core/socks"
	"github.com/MohsenBg/bgscan/internal/core/speedtest"
)

type fakePortManager struct {
	port        uint16
	getErr      error
	released    []uint16
	waitOpenErr error
}

func (m *fakePortManager) Get(context.Context) (uint16, error) {
	return m.port, m.getErr
}

func (m *fakePortManager) Release(port uint16) {
	m.released = append(m.released, port)
}

func (m *fakePortManager) Close() {}

func (m *fakePortManager) WaitOpen(context.Context, string, time.Duration) error {
	return m.waitOpenErr
}

// fakeClient implements ffislipstream.Client for testing.
type fakeClient struct {
	stopCalled bool
	stopErr    error
	running    bool
}

func (f *fakeClient) Running() bool { return f.running }
func (f *fakeClient) Stop() error {
	f.stopCalled = true
	return f.stopErr
}

var _ ffislipstream.Client = (*fakeClient)(nil)

// fakeSlipstreamService implements dns.SlipstreamService for testing
type fakeSlipstreamService struct {
	client        ffislipstream.Client
	runErr        error
	runCalled     bool
	runConfig     dns.SlipstreamConfig
	runResolverIP netip.Addr
	runListenPort uint16
	runKeepAlive  uint16
}

// EditConfig implements [dns.SlipstreamService].
func (s *fakeSlipstreamService) EditConfig(config dns.SlipstreamConfig, originalName string) error {
	return nil
}

func (s *fakeSlipstreamService) ValidateAllConfigs() ([]dns.ConfigValidationResult, error) {
	return nil, nil
}

func (s *fakeSlipstreamService) SaveConfig(config dns.SlipstreamConfig, name string) error {
	return nil
}

func (s *fakeSlipstreamService) LoadConfig(name string) (dns.SlipstreamConfig, error) {
	return dns.SlipstreamConfig{}, nil
}

func (s *fakeSlipstreamService) GetAllConfigFiles() ([]dns.SlipstreamConfigFile, error) {
	return []dns.SlipstreamConfigFile{}, nil
}

func (s *fakeSlipstreamService) RenameConfig(oldName, newName string) error {
	return nil
}

func (s *fakeSlipstreamService) RunTunnel(ctx context.Context, config dns.SlipstreamConfig, resolverIP netip.Addr, listenPort uint16, keepAlive uint16) (ffislipstream.Client, error) {
	s.runCalled = true
	s.runConfig = config
	s.runResolverIP = resolverIP
	s.runListenPort = listenPort
	s.runKeepAlive = keepAlive
	return s.client, s.runErr
}

// fakeSocksService implements socks.Service for testing
type fakeSocksService struct {
	connectErr    error
	connectCalled bool
	connectConn   net.Conn
	connectTarget string
	connectConfig socks.Config
}

func (s *fakeSocksService) Connect(ctx context.Context, conn net.Conn, target string, config socks.Config) (net.Conn, error) {
	s.connectCalled = true
	s.connectConn = conn
	s.connectTarget = target
	s.connectConfig = config
	return conn, s.connectErr
}

// fakeSpeedtestService implements speedtest.Service for testing
type fakeSpeedtestService struct {
	latencyErr    error
	latencyCalled bool
	latencyRTT    time.Duration
}

func (s *fakeSpeedtestService) MeasureLatency(ctx context.Context, cfg speedtest.LatencyConfig) (speedtest.LatencyResult, error) {
	s.latencyCalled = true
	if s.latencyErr != nil {
		return speedtest.LatencyResult{}, s.latencyErr
	}
	return speedtest.LatencyResult{RTT: s.latencyRTT, MaxLatency: cfg.MaxLatency}, nil
}

func (s *fakeSpeedtestService) MeasureDownloadSpeed(ctx context.Context, cfg speedtest.DownloadConfig) (speedtest.SpeedResult, error) {
	return speedtest.SpeedResult{}, nil
}

func (s *fakeSpeedtestService) MeasureUploadSpeed(ctx context.Context, cfg speedtest.UploadConfig) (speedtest.SpeedResult, error) {
	return speedtest.SpeedResult{}, nil
}

func validConfig() dns.SlipstreamConfig {
	return dns.SlipstreamConfig{
		Domain:            "tunnel.example.com",
		ResolverPort:      53,
		CertPath:          "/certs/ca.pem",
		DNSResolution:     dns.DNSResolutionRecursive,
		CongestionControl: dns.CongestionControlBBR,
		GSO:               false,
		KeepAliveInterval: 10,
		ProxyType:         dns.ResolverProxySOCKS,
		ProxyPort:         1080,
		AuthMethod:        dns.AuthPassword,
		Username:          "user",
		Password:          "pass",
	}
}

func testIP() netip.Addr {
	return netip.MustParseAddr("1.2.3.4")
}

func newTestProbe(t *testing.T, config dns.SlipstreamConfig, pm *fakePortManager,
	slipstreamSvc *fakeSlipstreamService, socksSvc *fakeSocksService, speedtestSvc *fakeSpeedtestService,
) *SlipstreamProbe {
	t.Helper()

	got, err := NewSlipstreamProbe(
		config,
		time.Second*5,
		pm,
		WithSlipstreamService(slipstreamSvc),
		WithSocksService(socksSvc),
		WithSpeedtestService(speedtestSvc),
	)
	if err != nil {
		t.Fatal(err)
	}

	p, ok := got.(*SlipstreamProbe)
	if !ok {
		t.Fatalf("NewSlipstreamProbe returned %T, want *SlipstreamProbe", got)
	}

	return p
}

func TestNewSlipstreamProbeValidation(t *testing.T) {
	slipstreamSvc := &fakeSlipstreamService{}
	socksSvc := &fakeSocksService{}
	speedtestSvc := &fakeSpeedtestService{}

	if _, err := NewSlipstreamProbe(validConfig(), time.Second*5, nil, WithSlipstreamService(slipstreamSvc), WithSocksService(socksSvc), WithSpeedtestService(speedtestSvc)); err == nil {
		t.Fatal("expected an error for a nil port manager")
	}

	config := validConfig()
	config.Domain = ""

	if _, err := NewSlipstreamProbe(config, time.Second*5, &fakePortManager{}, WithSlipstreamService(slipstreamSvc), WithSocksService(socksSvc), WithSpeedtestService(speedtestSvc)); err == nil {
		t.Fatal("expected an error for an empty domain")
	}
}

func TestNewSlipstreamProbeAppliesDefaultTimeoutAndOptions(t *testing.T) {
	config := validConfig()

	pm := &fakePortManager{}
	slipstreamSvc := &fakeSlipstreamService{}
	socksSvc := &fakeSocksService{}
	speedtestSvc := &fakeSpeedtestService{}

	got, err := NewSlipstreamProbe(
		config,
		time.Second*5,
		pm,
		WithSlipstreamService(slipstreamSvc),
		WithSocksService(socksSvc),
		WithSpeedtestService(speedtestSvc),
	)
	if err != nil {
		t.Fatal(err)
	}

	p := got.(*SlipstreamProbe)

	if p.config.Domain != config.Domain {
		t.Fatalf("domain not set correctly")
	}

	if p.slipstreamSvc != slipstreamSvc {
		t.Fatal("slipstream service option was not applied")
	}

	if p.socksService != socksSvc {
		t.Fatal("socks service option was not applied")
	}

	if p.speedtestSvc != speedtestSvc {
		t.Fatal("speedtest service option was not applied")
	}
}

func TestInitReturnsNil(t *testing.T) {
	p := newTestProbe(t, validConfig(), &fakePortManager{},
		&fakeSlipstreamService{}, &fakeSocksService{}, &fakeSpeedtestService{})

	if err := p.Init(context.Background()); err != nil {
		t.Fatalf("Init() error = %v, want nil", err)
	}
}

func TestRunReturnsCanceledContextBeforeAllocatingPort(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	pm := &fakePortManager{port: 4000}
	p := newTestProbe(t, validConfig(), pm,
		&fakeSlipstreamService{}, &fakeSocksService{}, &fakeSpeedtestService{})

	_, err := p.Run(ctx, testIP())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}

	if len(pm.released) != 0 {
		t.Fatalf("released ports = %v, want none", pm.released)
	}
}

func TestRunReleasesPortWhenAllocationFails(t *testing.T) {
	pm := &fakePortManager{getErr: errors.New("no ports available")}
	p := newTestProbe(t, validConfig(), pm,
		&fakeSlipstreamService{}, &fakeSocksService{}, &fakeSpeedtestService{})

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected an allocation error")
	}

	if len(pm.released) != 0 {
		t.Fatalf("released ports = %v, want none", pm.released)
	}
}

func TestRunReleasesPortWhenTunnelFails(t *testing.T) {
	client := &fakeClient{}
	pm := &fakePortManager{port: 5000}
	slipstreamSvc := &fakeSlipstreamService{client: client, runErr: errors.New("start failed")}
	p := newTestProbe(t, validConfig(), pm, slipstreamSvc, &fakeSocksService{}, &fakeSpeedtestService{})

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected a tunnel error")
	}

	if client.stopCalled {
		t.Fatal("Stop should not be called when the tunnel never started")
	}

	if got := pm.released; len(got) != 1 || got[0] != 5000 {
		t.Fatalf("released ports = %v, want [5000]", got)
	}
}

func TestRunStopsTunnelWhenWaitOpenFails(t *testing.T) {
	client := &fakeClient{}
	pm := &fakePortManager{port: 6000, waitOpenErr: errors.New("proxy never opened")}
	slipstreamSvc := &fakeSlipstreamService{client: client}
	p := newTestProbe(t, validConfig(), pm, slipstreamSvc, &fakeSocksService{}, &fakeSpeedtestService{})

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected a wait-open error")
	}

	if !client.stopCalled {
		t.Fatal("Stop was not called after wait-open failure")
	}

	if got := pm.released; len(got) != 1 || got[0] != 6000 {
		t.Fatalf("released ports = %v, want [6000]", got)
	}
}

func TestRunReturnsContextErrorWhenMeasureLatencyCancelsContext(t *testing.T) {
	ctx := t.Context()

	client := &fakeClient{}
	slipstreamSvc := &fakeSlipstreamService{client: client}
	socksSvc := &fakeSocksService{}
	speedtestSvc := &fakeSpeedtestService{
		latencyErr: context.Canceled,
	}

	p := newTestProbe(t, validConfig(), &fakePortManager{port: 8000}, slipstreamSvc, socksSvc, speedtestSvc)

	_, err := p.Run(ctx, testIP())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}

	if !client.stopCalled {
		t.Fatal("Stop was not called after latency error")
	}
}

func TestRunSuccess(t *testing.T) {
	client := &fakeClient{}
	pm := &fakePortManager{port: 9000}
	slipstreamSvc := &fakeSlipstreamService{client: client}
	socksSvc := &fakeSocksService{}
	speedtestSvc := &fakeSpeedtestService{latencyRTT: 50 * time.Millisecond}

	p := newTestProbe(t, validConfig(), pm, slipstreamSvc, socksSvc, speedtestSvc)

	result, err := p.Run(context.Background(), testIP())
	if err != nil {
		t.Fatal(err)
	}

	got, ok := result.(SlipstreamResult)
	if !ok {
		t.Fatalf("result type = %T, want SlipstreamResult", result)
	}

	if got.IP != testIP() || got.Port != 9000 {
		t.Fatalf("unexpected result: %#v", got)
	}

	if got.Latency != 50*time.Millisecond {
		t.Fatalf("latency = %s, want 50ms", got.Latency)
	}

	if !slipstreamSvc.runCalled || slipstreamSvc.runResolverIP != testIP() || slipstreamSvc.runListenPort != 9000 {
		t.Fatalf("RunTunnel call = (%q, %d), want (1.2.3.4, 9000)", slipstreamSvc.runResolverIP, slipstreamSvc.runListenPort)
	}

	// The probe derives keepalive from its timeout (5s in tests).
	if slipstreamSvc.runKeepAlive != 5 {
		t.Fatalf("RunTunnel keepalive = %d, want 5 (probe timeout)", slipstreamSvc.runKeepAlive)
	}

	if !speedtestSvc.latencyCalled {
		t.Fatal("MeasureLatency was not called")
	}

	if !client.stopCalled {
		t.Fatal("Stop was not called")
	}

	if got := pm.released; len(got) != 1 || got[0] != 9000 {
		t.Fatalf("released ports = %v, want [9000]", got)
	}
}

func TestClose(t *testing.T) {
	if err := (&SlipstreamProbe{}).Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
