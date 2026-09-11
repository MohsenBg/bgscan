package xrayprobe

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/core/speedtest"
	"github.com/MohsenBg/bgscan/internal/core/xray"
)

type fakePortManager struct {
	port        uint16
	getErr      error
	waitOpenErr error
	released    []uint16
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

type fakeInstance struct {
	closed   bool
	closeErr error
}

func (i *fakeInstance) Close() error {
	i.closed = true
	return i.closeErr
}

type fakeXrayService struct {
	templateErr error
	generateErr error
	validateErr error
	startErr    error

	instance *fakeInstance
}

func (s *fakeXrayService) GetOutboundTemplateByName(string) (*xray.XrayOutboundsFile, error) {
	if s.templateErr != nil {
		return nil, s.templateErr
	}

	return &xray.XrayOutboundsFile{}, nil
}

func (s *fakeXrayService) GenerateConfig(string, netip.Addr, uint16) (*xray.XrayConfig, error) {
	if s.generateErr != nil {
		return nil, s.generateErr
	}

	return &xray.XrayConfig{}, nil
}

func (s *fakeXrayService) ValidateConfig(context.Context, *xray.XrayConfig) error {
	return s.validateErr
}

func (s *fakeXrayService) Start(context.Context, *xray.XrayConfig) (Instance, error) {
	if s.startErr != nil {
		return nil, s.startErr
	}

	return s.instance, nil
}

type fakeSpeedTester struct {
	latencyResult  speedtest.LatencyResult
	latencyErr     error
	downloadResult speedtest.SpeedResult
	downloadErr    error
	uploadResult   speedtest.SpeedResult
	uploadErr      error
}

func (s *fakeSpeedTester) MeasureLatency(context.Context, speedtest.LatencyConfig) (speedtest.LatencyResult, error) {
	return s.latencyResult, s.latencyErr
}

func (s *fakeSpeedTester) MeasureDownloadSpeed(context.Context, speedtest.DownloadConfig) (speedtest.SpeedResult, error) {
	return s.downloadResult, s.downloadErr
}

func (s *fakeSpeedTester) MeasureUploadSpeed(context.Context, speedtest.UploadConfig) (speedtest.SpeedResult, error) {
	return s.uploadResult, s.uploadErr
}

func validConfig() *config.XrayConfig {
	return &config.XrayConfig{
		Timeout:              config.NewDurationMS(5 * time.Second),
		SpeedTestTimeout:     config.NewDurationMS(5 * time.Second),
		DownloadSpeed:        1000,
		UploadSpeed:          500,
		ConnectivityTestType: config.ConnectivityOnly,
	}
}

func testIP() netip.Addr {
	return netip.MustParseAddr("1.2.3.4")
}

func newTestProbe(mode config.ConnectivityTest) (*XrayProbe, *fakePortManager, *fakeXrayService, *fakeSpeedTester) {
	pm := &fakePortManager{port: 1080}
	service := &fakeXrayService{instance: &fakeInstance{}}
	speed := &fakeSpeedTester{
		latencyResult:  speedtest.LatencyResult{RTT: 25 * time.Millisecond},
		downloadResult: speedtest.SpeedResult{Speed: 50 * speedtest.Mbps},
		uploadResult:   speedtest.SpeedResult{Speed: 20 * speedtest.Mbps},
	}

	p := &XrayProbe{
		pm:              pm,
		xray:            service,
		speed:           speed,
		outbound:        "test-outbound",
		latencyTimeout:  5 * time.Second,
		transferTimeout: 5 * time.Second,
		testMode:        mode,
		downloadBytes:   1024,
		uploadBytes:     512,
		minDownload:     1000,
		minUpload:       500,
	}

	return p, pm, service, speed
}

func TestNewXrayProbeValidation(t *testing.T) {
	service := &fakeXrayService{}

	if _, err := NewXrayProbe(nil, "outbound", &fakePortManager{}, WithXrayService(service)); err == nil {
		t.Fatal("expected an error for a nil config")
	}

	if _, err := NewXrayProbe(validConfig(), "outbound", nil, WithXrayService(service)); err == nil {
		t.Fatal("expected an error for a nil port manager")
	}

	if _, err := NewXrayProbe(validConfig(), "", &fakePortManager{}, WithXrayService(service)); err == nil {
		t.Fatal("expected an error for an empty outbound name")
	}
}

func TestNewXrayProbeRejectsUnknownTemplate(t *testing.T) {
	service := &fakeXrayService{templateErr: errors.New("not found")}

	_, err := NewXrayProbe(
		validConfig(),
		"missing",
		&fakePortManager{},
		WithXrayService(service),
	)
	if err == nil {
		t.Fatal("expected an error")
	}

	if !errors.Is(err, service.templateErr) {
		t.Fatalf("error = %v, want wrapped template error", err)
	}
}

func TestNewXrayProbeAppliesOptionsAndCalculatesTransferSizes(t *testing.T) {
	cfg := validConfig()
	cfg.Timeout = config.NewDurationMS(10 * time.Second)
	cfg.SpeedTestTimeout = config.NewDurationMS(10 * time.Second)
	cfg.DownloadSpeed = 8000
	cfg.UploadSpeed = 4000
	cfg.ConnectivityTestType = config.Both

	pm := &fakePortManager{}
	service := &fakeXrayService{}
	speed := &fakeSpeedTester{}

	got, err := NewXrayProbe(
		cfg,
		"outbound",
		pm,
		WithXrayService(service),
		WithSpeedTester(speed),
	)
	if err != nil {
		t.Fatal(err)
	}

	p := got.(*XrayProbe)

	if p.xray != XrayService(service) || p.speed != speed {
		t.Fatal("injected dependency was not applied")
	}

	if p.downloadBytes != 10_000_000 {
		t.Fatalf("download bytes = %d, want 10000000", p.downloadBytes)
	}

	if p.uploadBytes != 5_000_000 {
		t.Fatalf("upload bytes = %d, want 5000000", p.uploadBytes)
	}
}

func TestInitValidatesTemplate(t *testing.T) {
	p, pm, _, _ := newTestProbe(config.ConnectivityOnly)

	if err := p.Init(context.Background()); err != nil {
		t.Fatal(err)
	}

	if got := pm.released; len(got) != 1 || got[0] != 1080 {
		t.Fatalf("released ports = %v, want [1080]", got)
	}
}

func TestRunReturnsCanceledContextBeforeAllocatingPort(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p, pm, _, _ := newTestProbe(config.ConnectivityOnly)

	_, err := p.Run(ctx, testIP())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}

	if len(pm.released) != 0 {
		t.Fatalf("released ports = %v, want none", pm.released)
	}
}

func TestRunReleasesPortWhenConfigGenerationFails(t *testing.T) {
	p, pm, service, _ := newTestProbe(config.ConnectivityOnly)
	service.generateErr = errors.New("generation failed")

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected an error")
	}

	if got := pm.released; len(got) != 1 || got[0] != 1080 {
		t.Fatalf("released ports = %v, want [1080]", got)
	}
}

func TestRunReleasesPortWhenStartFails(t *testing.T) {
	p, pm, service, _ := newTestProbe(config.ConnectivityOnly)
	service.startErr = errors.New("start failed")

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected an error")
	}

	if got := pm.released; len(got) != 1 || got[0] != 1080 {
		t.Fatalf("released ports = %v, want [1080]", got)
	}
}

func TestInitFailsOnValidationError(t *testing.T) {
	p, pm, service, _ := newTestProbe(config.ConnectivityOnly)
	service.validateErr = errors.New("invalid config")

	err := p.Init(context.Background())
	if err == nil {
		t.Fatal("expected an error")
	}

	if got := pm.released; len(got) != 1 || got[0] != 1080 {
		t.Fatalf("released ports = %v, want [1080]", got)
	}
}

func TestRunClosesInstanceAfterWaitOpenFailure(t *testing.T) {
	p, pm, service, _ := newTestProbe(config.ConnectivityOnly)
	pm.waitOpenErr = errors.New("proxy did not open")

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected an error")
	}

	if !service.instance.closed {
		t.Fatal("instance was not closed")
	}

	if got := pm.released; len(got) != 1 || got[0] != 1080 {
		t.Fatalf("released ports = %v, want [1080]", got)
	}
}

func TestRunConnectivityOnly(t *testing.T) {
	p, pm, service, _ := newTestProbe(config.ConnectivityOnly)

	result, err := p.Run(context.Background(), testIP())
	if err != nil {
		t.Fatal(err)
	}

	got, ok := result.(XrayResult)
	if !ok {
		t.Fatalf("result type = %T, want XrayResult", result)
	}

	if got.IP != testIP() || got.Latency != 25*time.Millisecond {
		t.Fatalf("unexpected result: %#v", got)
	}

	if !service.instance.closed {
		t.Fatal("instance was not closed")
	}

	if len(pm.released) != 1 {
		t.Fatal("port was not released")
	}
}

func TestRunBoth(t *testing.T) {
	p, _, _, _ := newTestProbe(config.Both)

	result, err := p.Run(context.Background(), testIP())
	if err != nil {
		t.Fatal(err)
	}

	got := result.(XrayResult)

	if got.Download != 50*speedtest.Mbps {
		t.Fatalf("download = %s, want 50 Mbps", got.Download)
	}

	if got.Upload != 20*speedtest.Mbps {
		t.Fatalf("upload = %s, want 20 Mbps", got.Upload)
	}
}

func TestRunReturnsSpeedTestError(t *testing.T) {
	p, _, _, speed := newTestProbe(config.DownloadSpeedOnly)
	speed.downloadErr = errors.New("speed below minimum")

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected an error")
	}

	if !errors.Is(err, speed.downloadErr) {
		t.Fatalf("error = %v, want wrapped download error", err)
	}
}

func TestClose(t *testing.T) {
	if err := (&XrayProbe{}).Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
