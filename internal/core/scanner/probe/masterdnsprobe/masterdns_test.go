package masterdnsprobe

import (
	"context"
	"errors"
	"io"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/scanner/portmgr"
	"github.com/MohsenBg/bgscan/internal/core/socks"
	"github.com/MohsenBg/bgscan/internal/core/speedtest"
)

type fakePortManager struct {
	getPort     uint16
	getErr      error
	waitOpenErr error

	getCalled    bool
	released     []uint16
	waitOpenAddr string
}

func (f *fakePortManager) Get(context.Context) (uint16, error) {
	f.getCalled = true
	return f.getPort, f.getErr
}

func (f *fakePortManager) Release(port uint16) {
	f.released = append(f.released, port)
}

func (f *fakePortManager) Close() {}

func (f *fakePortManager) WaitOpen(_ context.Context, addr string, _ time.Duration) error {
	f.waitOpenAddr = addr
	return f.waitOpenErr
}

type fakeMasterDNSService struct {
	runTunnelErr error
	runIP        string
	runPort      uint16
	runConfig    dns.MasterDNSConfig
	runCalled    bool

	closed bool
}

func (f *fakeMasterDNSService) SaveConfig(dns.MasterDNSConfig, string) error { return nil }
func (f *fakeMasterDNSService) EditConfig(dns.MasterDNSConfig, string) error { return nil }
func (f *fakeMasterDNSService) LoadConfig(string) (dns.MasterDNSConfig, error) {
	return dns.MasterDNSConfig{}, nil
}

func (f *fakeMasterDNSService) GetAllConfigFiles() ([]dns.MasterDNSConfigFile, error) {
	return nil, nil
}

func (f *fakeMasterDNSService) ValidateAllConfigs() ([]dns.ConfigValidationResult, error) {
	return nil, nil
}
func (f *fakeMasterDNSService) RenameConfig(string, string) error { return nil }

func (f *fakeMasterDNSService) RunTunnel(
	ctx context.Context,
	config dns.MasterDNSConfig,
	resolverIP string,
	listenPort uint16,
) (io.Closer, error) {
	f.runCalled = true
	f.runConfig = config
	f.runIP = resolverIP
	f.runPort = listenPort

	if f.runTunnelErr != nil {
		return nil, f.runTunnelErr
	}

	return &fakeTunnelHandle{closed: &f.closed}, nil
}

type fakeTunnelHandle struct {
	closed *bool
}

func (h *fakeTunnelHandle) Close() error {
	*h.closed = true
	return nil
}

type fakeSocksService struct {
	connectErr    error
	connectCalled bool
	connectTarget string
}

func (f *fakeSocksService) Connect(ctx context.Context, conn net.Conn, target string, cfg socks.Config) (net.Conn, error) {
	f.connectCalled = true
	f.connectTarget = target

	if f.connectErr != nil {
		_ = conn.Close()
		return nil, f.connectErr
	}

	return conn, nil
}

type fakeSpeedtestService struct {
	latencyErr error
	latencyRTT time.Duration
	gotDialer  func(context.Context, string, string) (net.Conn, error)
}

func (f *fakeSpeedtestService) MeasureLatency(ctx context.Context, cfg speedtest.LatencyConfig) (speedtest.LatencyResult, error) {
	f.gotDialer = cfg.DialContext

	if f.latencyErr != nil {
		return speedtest.LatencyResult{}, f.latencyErr
	}

	return speedtest.LatencyResult{RTT: f.latencyRTT}, nil
}

func (f *fakeSpeedtestService) MeasureDownloadSpeed(context.Context, speedtest.DownloadConfig) (speedtest.SpeedResult, error) {
	return speedtest.SpeedResult{}, nil
}

func (f *fakeSpeedtestService) MeasureUploadSpeed(context.Context, speedtest.UploadConfig) (speedtest.SpeedResult, error) {
	return speedtest.SpeedResult{}, nil
}

func validConfig() dns.MasterDNSConfig {
	cfg := dns.DefaultMasterDNSConfig()
	cfg.Domain = "tunnel.example.com"
	cfg.EncryptionKey = "cd6d78e954f48f62cb74cdcf8a2459d3"
	return cfg
}

func testIP() netip.Addr {
	return netip.MustParseAddr("1.2.3.4")
}

func newTestProbe(t *testing.T, pm portmgr.Manager, svc *fakeMasterDNSService) *MasterDNSProbe {
	t.Helper()

	got, err := NewMasterDNSProbe(
		validConfig(),
		5*time.Second,
		pm,
		WithMasterDNSService(svc),
		WithSocksService(&fakeSocksService{}),
		WithSpeedtestService(&fakeSpeedtestService{latencyRTT: 42 * time.Millisecond}),
	)
	if err != nil {
		t.Fatalf("NewMasterDNSProbe: %v", err)
	}

	p, ok := got.(*MasterDNSProbe)
	if !ok {
		t.Fatalf("NewMasterDNSProbe returned %T, want *MasterDNSProbe", got)
	}

	return p
}

// ---- tests -----------------------------------------------------------

func TestNewMasterDNSProbe_NilPortManager(t *testing.T) {
	if _, err := NewMasterDNSProbe(validConfig(), time.Second, nil); err == nil {
		t.Fatal("expected error for nil port manager")
	}
}

func TestNewMasterDNSProbe_InvalidConfig(t *testing.T) {
	_, err := NewMasterDNSProbe(dns.MasterDNSConfig{}, time.Second, &fakePortManager{})
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}

func TestNewMasterDNSProbe_DefaultsServices(t *testing.T) {
	got, err := NewMasterDNSProbe(validConfig(), time.Second, &fakePortManager{})
	if err != nil {
		t.Fatalf("NewMasterDNSProbe: %v", err)
	}

	p, ok := got.(*MasterDNSProbe)
	if !ok {
		t.Fatalf("returned %T, want *MasterDNSProbe", got)
	}

	if p.masterDNSService == nil {
		t.Error("masterDNSService not defaulted")
	}
	if p.socksService == nil {
		t.Error("socksService not defaulted")
	}
	if p.speedtestService == nil {
		t.Error("speedtestService not defaulted")
	}
}

func TestRun_ContextCanceled(t *testing.T) {
	p := newTestProbe(t, &fakePortManager{}, &fakeMasterDNSService{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.Run(ctx, testIP())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestRun_PortManagerFails(t *testing.T) {
	pm := &fakePortManager{getErr: errors.New("no ports")}
	p := newTestProbe(t, pm, &fakeMasterDNSService{})

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected port manager error")
	}

	if p.masterDNSService.(*fakeMasterDNSService).runCalled {
		t.Error("tunnel should not start when port allocation fails")
	}
}

func TestRun_TunnelStartFails(t *testing.T) {
	pm := &fakePortManager{getPort: 20001}
	svc := &fakeMasterDNSService{runTunnelErr: errors.New("session init failed")}
	p := newTestProbe(t, pm, svc)

	_, err := p.Run(context.Background(), testIP())
	if err == nil || !errors.Is(err, svc.runTunnelErr) {
		t.Fatalf("error = %v, want wrapped session-init failure", err)
	}
}

func TestRun_WaitOpenFails(t *testing.T) {
	pm := &fakePortManager{
		getPort:     20002,
		waitOpenErr: errors.New("listener never opened"),
	}
	svc := &fakeMasterDNSService{}
	p := newTestProbe(t, pm, svc)

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected WaitOpen failure")
	}

	if !svc.closed {
		t.Error("tunnel should be closed after WaitOpen failure")
	}
}

func TestRun_MeasureLatencyFails(t *testing.T) {
	pm := &fakePortManager{getPort: 20003}
	svc := &fakeMasterDNSService{}
	wantErr := errors.New("probe timeout")

	got, err := NewMasterDNSProbe(
		validConfig(),
		5*time.Second,
		pm,
		WithMasterDNSService(svc),
		WithSocksService(&fakeSocksService{}),
		WithSpeedtestService(&fakeSpeedtestService{latencyErr: wantErr}),
	)
	if err != nil {
		t.Fatalf("NewMasterDNSProbe: %v", err)
	}

	_, err = got.Run(context.Background(), testIP())
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapped %v", err, wantErr)
	}

	if !svc.closed {
		t.Error("tunnel should be closed after latency failure")
	}
}

func TestRun_Success(t *testing.T) {
	const listenPort uint16 = 20010

	pm := &fakePortManager{getPort: listenPort}
	svc := &fakeMasterDNSService{}
	socksSvc := &fakeSocksService{}
	speedtestSvc := &fakeSpeedtestService{latencyRTT: 42 * time.Millisecond}

	got, err := NewMasterDNSProbe(
		validConfig(),
		5*time.Second,
		pm,
		WithMasterDNSService(svc),
		WithSocksService(socksSvc),
		WithSpeedtestService(speedtestSvc),
	)
	if err != nil {
		t.Fatalf("NewMasterDNSProbe: %v", err)
	}

	res, err := got.Run(context.Background(), testIP())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// The scanned target must be passed as the resolver IP and the
	// allocated port as the listener port.
	if !svc.runCalled {
		t.Fatal("RunTunnel was not called")
	}
	if svc.runIP != testIP().String() {
		t.Errorf("resolver IP = %q, want %q", svc.runIP, testIP().String())
	}
	if svc.runPort != listenPort {
		t.Errorf("listen port = %d, want %d", svc.runPort, listenPort)
	}
	if svc.runConfig.Domain != validConfig().Domain {
		t.Errorf("config domain = %q, want %q", svc.runConfig.Domain, validConfig().Domain)
	}

	// The proxy listener address must be checked and the port released.
	if pm.waitOpenAddr != net.JoinHostPort("127.0.0.1", "20010") {
		t.Errorf("WaitOpen addr = %q, want local proxy addr", pm.waitOpenAddr)
	}
	if len(pm.released) != 1 || pm.released[0] != listenPort {
		t.Errorf("released = %v, want [%d]", pm.released, listenPort)
	}

	// The tunnel must be shut down after the probe finishes.
	if !svc.closed {
		t.Error("tunnel handle was not closed")
	}

	// The latency dialer must be wired through the local SOCKS listener.
	if speedtestSvc.gotDialer == nil {
		t.Fatal("MeasureLatency got no DialContext")
	}

	gotResult, ok := res.(MasterDNSResult)
	if !ok {
		t.Fatalf("result type = %T, want MasterDNSResult", res)
	}
	if gotResult.IP != testIP() {
		t.Errorf("IP = %v, want %v", gotResult.IP, testIP())
	}
	if gotResult.Latency != 42*time.Millisecond {
		t.Errorf("Latency = %v, want 42ms", gotResult.Latency)
	}
	if gotResult.Port != validConfig().ResolverPort {
		t.Errorf("Port = %d, want resolver port %d", gotResult.Port, validConfig().ResolverPort)
	}
	if gotResult.Enc != dns.EncXOR {
		t.Errorf("Enc = %v, want XOR", gotResult.Enc)
	}
}

func TestDialSOCKS(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = listener.Close() }()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		_, _ = conn.Read(make([]byte, 64))
	}()

	socksSvc := &fakeSocksService{}
	p := newTestProbe(t, &fakePortManager{}, &fakeMasterDNSService{})
	p.socksService = socksSvc

	dialer := p.dialSOCKS(listener.Addr().String())

	conn, err := dialer(context.Background(), "tcp", "www.google.com:80")
	if err != nil {
		t.Fatalf("dialer: %v", err)
	}
	_ = conn.Close()

	if !socksSvc.connectCalled {
		t.Error("SOCKS connect was not invoked")
	}
	if socksSvc.connectTarget != "www.google.com:80" {
		t.Errorf("SOCKS target = %q, want www.google.com:80", socksSvc.connectTarget)
	}
}

func TestClose_NoOp(t *testing.T) {
	p := newTestProbe(t, &fakePortManager{}, &fakeMasterDNSService{})

	if err := p.Close(); err != nil {
		t.Fatalf("Close() = %v, want nil", err)
	}
}
