package slipstreamprobe

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"testing"
	"time"

	"bgscan/internal/core/dns"
	"bgscan/internal/core/process"
	"bgscan/internal/core/scanner/probe"
	"bgscan/internal/core/socks"
	"bgscan/internal/core/speedtest"
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

type fakeProcess struct {
	stopCalled bool
}

func (f *fakeProcess) StopGracefully(time.Duration) error { f.stopCalled = true; return nil }
func (f *fakeProcess) Kill() error                        { f.stopCalled = true; return nil }
func (*fakeProcess) Wait() error                          { return nil }

type fakeProcessTracker struct {
	started         bool
	registerErr     error
	unregisterErr   error
	registeredID    string
	unregisteredIDs []string
}

func (t *fakeProcessTracker) Start(context.Context) {
	t.started = true
}

func (t *fakeProcessTracker) Register(context.Context, probe.Killable) (string, error) {
	return t.registeredID, t.registerErr
}

func (t *fakeProcessTracker) Unregister(_ context.Context, id string) error {
	t.unregisteredIDs = append(t.unregisteredIDs, id)
	return t.unregisterErr
}

// fakeSlipstreamService implements dns.SlipstreamService for testing
type fakeSlipstreamService struct {
	process       process.Process
	runErr        error
	runCalled     bool
	runConfig     dns.SlipstreamConfig
	runResolverIP string
	runListenPort uint16
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

func (s *fakeSlipstreamService) RunTunnel(ctx context.Context, config dns.SlipstreamConfig, resolverIP string, listenPort uint16) (process.Process, error) {
	s.runCalled = true
	s.runConfig = config
	s.runResolverIP = resolverIP
	s.runListenPort = listenPort
	return s.process, s.runErr
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
		Domain:       "tunnel.example.com",
		ResolverPort: 53,
		CertPath:     "/certs/ca.pem",
		ProxyType:    dns.ResolverProxySOCKS,
		ProxyPort:    0,
		AuthMethod:   dns.AuthPassword,
		Username:     "user",
		Password:     "pass",
	}
}

func testIP() netip.Addr {
	return netip.MustParseAddr("1.2.3.4")
}

func newTestProbe(t *testing.T, config dns.SlipstreamConfig, pm *fakePortManager, tracker *fakeProcessTracker,
	slipstreamSvc *fakeSlipstreamService, socksSvc *fakeSocksService, speedtestSvc *fakeSpeedtestService,
) *SlipstreamProbe {
	t.Helper()

	got, err := NewSlipstreamProbe(
		config,
		time.Second*5,
		pm,
		WithProcessTracker(tracker),
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
	tracker := &fakeProcessTracker{}
	slipstreamSvc := &fakeSlipstreamService{}
	socksSvc := &fakeSocksService{}
	speedtestSvc := &fakeSpeedtestService{}

	got, err := NewSlipstreamProbe(
		config,
		time.Second*5,
		pm,
		WithProcessTracker(tracker),
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

	if p.processTracker != tracker {
		t.Fatal("process tracker option was not applied")
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

func TestInitStartsProcessTracker(t *testing.T) {
	tracker := &fakeProcessTracker{}
	p := newTestProbe(t, validConfig(), &fakePortManager{}, tracker,
		&fakeSlipstreamService{}, &fakeSocksService{}, &fakeSpeedtestService{})

	if err := p.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	if !tracker.started {
		t.Fatal("process tracker was not started")
	}
}

func TestRunReturnsCanceledContextBeforeAllocatingPort(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	pm := &fakePortManager{port: 4000}
	p := newTestProbe(t, validConfig(), pm, &fakeProcessTracker{},
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
	p := newTestProbe(t, validConfig(), pm, &fakeProcessTracker{},
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
	prc := &fakeProcess{}
	pm := &fakePortManager{port: 5000}
	slipstreamSvc := &fakeSlipstreamService{process: prc, runErr: errors.New("start failed")}
	p := newTestProbe(t, validConfig(), pm, &fakeProcessTracker{}, slipstreamSvc, &fakeSocksService{}, &fakeSpeedtestService{})

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected a tunnel error")
	}

	if got := pm.released; len(got) != 1 || got[0] != 5000 {
		t.Fatalf("released ports = %v, want [5000]", got)
	}
}

func TestRunStopsTunnelWhenRegistrationFails(t *testing.T) {
	prc := &fakeProcess{}
	pm := &fakePortManager{port: 6000}
	tracker := &fakeProcessTracker{registerErr: errors.New("register failed")}
	slipstreamSvc := &fakeSlipstreamService{process: prc}
	p := newTestProbe(t, validConfig(), pm, tracker, slipstreamSvc, &fakeSocksService{}, &fakeSpeedtestService{})

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected a registration error")
	}

	if len(tracker.unregisteredIDs) != 0 {
		t.Fatalf("unregistered IDs = %v, want none", tracker.unregisteredIDs)
	}
}

func TestRunReturnsContextErrorWhenMeasureLatencyCancelsContext(t *testing.T) {
	ctx := t.Context()

	tracker := &fakeProcessTracker{registeredID: "process-1"}
	prc := &fakeProcess{}
	slipstreamSvc := &fakeSlipstreamService{process: prc}
	socksSvc := &fakeSocksService{}
	speedtestSvc := &fakeSpeedtestService{
		latencyErr: context.Canceled,
	}

	p := newTestProbe(t, validConfig(), &fakePortManager{port: 8000}, tracker, slipstreamSvc, socksSvc, speedtestSvc)

	_, err := p.Run(ctx, testIP())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestRunSuccess(t *testing.T) {
	prc := &fakeProcess{}
	pm := &fakePortManager{port: 9000}
	tracker := &fakeProcessTracker{registeredID: "process-1"}
	slipstreamSvc := &fakeSlipstreamService{process: prc}
	socksSvc := &fakeSocksService{}
	speedtestSvc := &fakeSpeedtestService{latencyRTT: 50 * time.Millisecond}

	p := newTestProbe(t, validConfig(), pm, tracker, slipstreamSvc, socksSvc, speedtestSvc)

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

	if !slipstreamSvc.runCalled || slipstreamSvc.runResolverIP != "1.2.3.4" || slipstreamSvc.runListenPort != 9000 {
		t.Fatalf("RunTunnel call = (%q, %d), want (1.2.3.4, 9000)", slipstreamSvc.runResolverIP, slipstreamSvc.runListenPort)
	}

	if !speedtestSvc.latencyCalled {
		t.Fatal("MeasureLatency was not called")
	}

	if !prc.stopCalled {
		t.Fatal("StopGracefully was not called")
	}

	if got := tracker.unregisteredIDs; len(got) != 1 || got[0] != "process-1" {
		t.Fatalf("unregistered IDs = %v, want [process-1]", got)
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
