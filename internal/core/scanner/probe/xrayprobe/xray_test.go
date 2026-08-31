package xrayprobe

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"bgscan/internal/core/config"
	"bgscan/internal/core/process"
	"bgscan/internal/core/speedtest"
	"bgscan/internal/core/xray"
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

type fakeProcess struct {
	killed  bool
	killErr error
}

func (*fakeProcess) StopGracefully(time.Duration) error { return nil }

func (p *fakeProcess) Kill() error {
	p.killed = true
	return p.killErr
}

func (*fakeProcess) Wait() error { return nil }

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

func (t *fakeProcessTracker) Register(context.Context, process.Killable) (string, error) {
	return t.registeredID, t.registerErr
}

func (t *fakeProcessTracker) Unregister(_ context.Context, id string) error {
	t.unregisteredIDs = append(t.unregisteredIDs, id)
	return t.unregisterErr
}

type fakeXrayService struct {
	templateErr error
	generateErr error
	validateErr error
	startErr    error

	configPath string
	process    process.Process
}

func (s *fakeXrayService) GetOutboundTemplateByName(string) (*xray.XrayOutboundsFile, error) {
	if s.templateErr != nil {
		return nil, s.templateErr
	}

	return &xray.XrayOutboundsFile{}, nil
}

func (s *fakeXrayService) GenerateConfig(string, netip.Addr, uint16) (string, error) {
	if s.generateErr != nil {
		return "", s.generateErr
	}

	return s.configPath, nil
}

func (s *fakeXrayService) ValidateConfig(context.Context, string) error {
	return s.validateErr
}

func (s *fakeXrayService) Start(context.Context, string) (process.Process, error) {
	if s.startErr != nil {
		return nil, s.startErr
	}

	return s.process, nil
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

func newTestProbe(mode config.ConnectivityTest) (*XrayProbe, *fakePortManager, *fakeProcessTracker, *fakeXrayService, *fakeSpeedTester) {
	pm := &fakePortManager{port: 1080}
	tracker := &fakeProcessTracker{registeredID: "process-1"}
	service := &fakeXrayService{
		configPath: "/tmp/xray.json",
		process:    &fakeProcess{},
	}
	speed := &fakeSpeedTester{
		latencyResult:  speedtest.LatencyResult{RTT: 25 * time.Millisecond},
		downloadResult: speedtest.SpeedResult{Speed: 50 * speedtest.Mbps},
		uploadResult:   speedtest.SpeedResult{Speed: 20 * speedtest.Mbps},
	}

	p := &XrayProbe{
		pm:              pm,
		processTracker:  tracker,
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
		remove:          func(string) error { return nil },
	}

	return p, pm, tracker, service, speed
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
	tracker := &fakeProcessTracker{}
	service := &fakeXrayService{}
	speed := &fakeSpeedTester{}

	got, err := NewXrayProbe(
		cfg,
		"outbound",
		pm,
		WithProcessTracker(tracker),
		WithXrayService(service),
		WithSpeedTester(speed),
	)
	if err != nil {
		t.Fatal(err)
	}

	p := got.(*XrayProbe)

	if p.processTracker != tracker || p.xray != service || p.speed != speed {
		t.Fatal("injected dependency was not applied")
	}

	if p.downloadBytes != 10_000_000 {
		t.Fatalf("download bytes = %d, want 10000000", p.downloadBytes)
	}

	if p.uploadBytes != 5_000_000 {
		t.Fatalf("upload bytes = %d, want 5000000", p.uploadBytes)
	}
}

func TestInitStartsProcessTracker(t *testing.T) {
	p, _, tracker, _, _ := newTestProbe(config.ConnectivityOnly)

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

	p, pm, _, _, _ := newTestProbe(config.ConnectivityOnly)

	_, err := p.Run(ctx, testIP())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}

	if len(pm.released) != 0 {
		t.Fatalf("released ports = %v, want none", pm.released)
	}
}

func TestRunReleasesPortWhenConfigGenerationFails(t *testing.T) {
	p, pm, _, service, _ := newTestProbe(config.ConnectivityOnly)
	service.generateErr = errors.New("generation failed")

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected an error")
	}

	if got := pm.released; len(got) != 1 || got[0] != 1080 {
		t.Fatalf("released ports = %v, want [1080]", got)
	}
}

func TestInitRemovesConfigAfterValidationFailure(t *testing.T) {
	p, pm, _, service, _ := newTestProbe(config.ConnectivityOnly)
	service.validateErr = errors.New("invalid config")

	var removed string
	p.remove = func(path string) error {
		removed = path
		return nil
	}

	err := p.Init(context.Background())
	if err == nil {
		t.Fatal("expected an error")
	}

	if removed != "/tmp/xray.json" {
		t.Fatalf("removed path = %q, want /tmp/xray.json", removed)
	}

	if got := pm.released; len(got) != 1 || got[0] != 1080 {
		t.Fatalf("released ports = %v, want [1080]", got)
	}
}

func TestRunKillsProcessWhenRegistrationFails(t *testing.T) {
	p, _, tracker, service, _ := newTestProbe(config.ConnectivityOnly)
	tracker.registerErr = errors.New("registry full")

	process := service.process.(*fakeProcess)

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected an error")
	}

	if !process.killed {
		t.Fatal("process was not killed")
	}
}

func TestRunCleansUpAfterWaitOpenFailure(t *testing.T) {
	p, pm, tracker, service, _ := newTestProbe(config.ConnectivityOnly)
	pm.waitOpenErr = errors.New("proxy did not open")

	process := service.process.(*fakeProcess)

	_, err := p.Run(context.Background(), testIP())
	if err == nil {
		t.Fatal("expected an error")
	}

	if !process.killed {
		t.Fatal("process was not killed")
	}

	if got := tracker.unregisteredIDs; len(got) != 1 || got[0] != "process-1" {
		t.Fatalf("unregistered IDs = %v, want [process-1]", got)
	}

	if got := pm.released; len(got) != 1 || got[0] != 1080 {
		t.Fatalf("released ports = %v, want [1080]", got)
	}
}

func TestRunConnectivityOnly(t *testing.T) {
	p, pm, tracker, service, _ := newTestProbe(config.ConnectivityOnly)

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

	if !service.process.(*fakeProcess).killed {
		t.Fatal("process was not killed")
	}

	if len(tracker.unregisteredIDs) != 1 || len(pm.released) != 1 {
		t.Fatal("process and port cleanup did not occur")
	}
}

func TestRunBoth(t *testing.T) {
	p, _, _, _, _ := newTestProbe(config.Both)

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
	p, _, _, _, speed := newTestProbe(config.DownloadSpeedOnly)
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
