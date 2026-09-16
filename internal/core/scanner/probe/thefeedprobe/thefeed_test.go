package thefeedprobe

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
)

// mockTheFeedService is a controllable TheFeedService for probe tests.
type mockTheFeedService struct {
	runErr   error
	runIP    netip.Addr
	runCfg   dns.TheFeedConfig
	runT     time.Duration
	runCalls int
}

func (m *mockTheFeedService) SaveConfig(dns.TheFeedConfig, string) error { return nil }
func (m *mockTheFeedService) EditConfig(dns.TheFeedConfig, string) error { return nil }
func (m *mockTheFeedService) LoadConfig(string) (dns.TheFeedConfig, error) {
	return dns.TheFeedConfig{}, nil
}
func (m *mockTheFeedService) GetAllConfigFiles() ([]dns.TheFeedConfigFile, error) {
	return nil, nil
}
func (m *mockTheFeedService) ValidateAllConfigs() ([]dns.ConfigValidationResult, error) {
	return nil, nil
}
func (m *mockTheFeedService) RenameConfig(string, string) error { return nil }

func (m *mockTheFeedService) RunTunnel(
	ctx context.Context,
	config dns.TheFeedConfig,
	resolverAddr netip.Addr,
	timeout time.Duration,
) error {
	m.runCalls++
	m.runCfg = config
	m.runIP = resolverAddr
	m.runT = timeout
	return m.runErr
}

// validProbeConfig returns a TheFeedConfig that passes Validate().
func validProbeConfig() dns.TheFeedConfig {
	cfg := dns.DefaultTheFeedConfig()
	cfg.Domain = "t.example.com"
	cfg.Passphrase = "secret"
	cfg.ResolverPort = 5300
	return cfg
}

func TestNewTheFeedProbe_InvalidConfig(t *testing.T) {
	_, err := NewTheFeedProbe(dns.TheFeedConfig{}, time.Second)
	if err == nil {
		t.Fatal("expected error for invalid config, got nil")
	}
}

func TestNewTheFeedProbe_DefaultsService(t *testing.T) {
	p, err := NewTheFeedProbe(validProbeConfig(), time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	probe, ok := p.(*TheFeedProbe)
	if !ok {
		t.Fatalf("expected *TheFeedProbe, got %T", p)
	}
	if probe.thefeedSvc == nil {
		t.Error("expected default thefeedService to be set")
	}
	if probe.timeout != time.Second {
		t.Errorf("timeout = %v, want %v", probe.timeout, time.Second)
	}
}

func TestNewTheFeedProbe_WithService(t *testing.T) {
	mock := &mockTheFeedService{}
	p, err := NewTheFeedProbe(
		validProbeConfig(),
		time.Second,
		WithTheFeedService(mock),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	probe := p.(*TheFeedProbe)
	if probe.thefeedSvc != mock {
		t.Error("expected injected service to be used")
	}
}

func TestSchema(t *testing.T) {
	p, err := NewTheFeedProbe(validProbeConfig(), time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.Schema().Name != "TheFeed" {
		t.Errorf("schema name = %q, want TheFeed", p.Schema().Name)
	}
}

func TestInitNoop(t *testing.T) {
	p, err := NewTheFeedProbe(validProbeConfig(), time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := p.Init(context.Background()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestCloseNoop(t *testing.T) {
	p, err := NewTheFeedProbe(validProbeConfig(), time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := p.Close(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRun_ContextAlreadyCanceled(t *testing.T) {
	mock := &mockTheFeedService{}
	p, err := NewTheFeedProbe(
		validProbeConfig(),
		time.Second,
		WithTheFeedService(mock),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = p.Run(ctx, netip.MustParseAddr("2.2.2.2"))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestRun_ServiceError(t *testing.T) {
	wantErr := errors.New("dns exchange: timeout")
	mock := &mockTheFeedService{runErr: wantErr}
	p, err := NewTheFeedProbe(
		validProbeConfig(),
		time.Second,
		WithTheFeedService(mock),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = p.Run(context.Background(), netip.MustParseAddr("2.2.2.2"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if mock.runCalls != 1 {
		t.Errorf("RunTunnel calls = %d, want 1", mock.runCalls)
	}
	if mock.runIP != netip.MustParseAddr("2.2.2.2") {
		t.Errorf("resolver IP passed to service = %s, want 2.2.2.2", mock.runIP)
	}
	if mock.runT != time.Second {
		t.Errorf("timeout passed to service = %v, want %v", mock.runT, time.Second)
	}
}

func TestRun_Success(t *testing.T) {
	cfg := validProbeConfig()
	mock := &mockTheFeedService{}
	p, err := NewTheFeedProbe(
		cfg,
		time.Second,
		WithTheFeedService(mock),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	res, err := p.Run(context.Background(), netip.MustParseAddr("2.2.2.2"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, ok := res.(TheFeedResult)
	if !ok {
		t.Fatalf("expected TheFeedResult, got %T", res)
	}
	if got.IP != netip.MustParseAddr("2.2.2.2") {
		t.Errorf("IP = %v, want 2.2.2.2", got.IP)
	}
	if got.Domain != cfg.Domain {
		t.Errorf("Domain = %q, want %q", got.Domain, cfg.Domain)
	}
	if got.QueryMode != cfg.QueryMode {
		t.Errorf("QueryMode = %q, want %q", got.QueryMode, cfg.QueryMode)
	}
	if got.Latency <= 0 {
		t.Errorf("Latency = %v, want positive", got.Latency)
	}

	// The scanned target must be passed through to the service.
	if mock.runIP != netip.MustParseAddr("2.2.2.2") {
		t.Errorf("resolver IP passed to service = %s, want 2.2.2.2", mock.runIP)
	}
}

func TestRun_PassesConfig(t *testing.T) {
	cfg := validProbeConfig()
	cfg.QueryMode = "double"

	mock := &mockTheFeedService{}
	p, err := NewTheFeedProbe(
		cfg,
		time.Second,
		WithTheFeedService(mock),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := p.Run(context.Background(), netip.MustParseAddr("2.2.2.2")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.runCfg.Domain != cfg.Domain {
		t.Errorf("service domain = %q, want %q", mock.runCfg.Domain, cfg.Domain)
	}
	if mock.runCfg.QueryMode != "double" {
		t.Errorf("service query mode = %q, want double", mock.runCfg.QueryMode)
	}
}

func TestRun_RetriesUpToTries(t *testing.T) {
	mock := &mockTheFeedService{runErr: errors.New("boom")}
	p, err := NewTheFeedProbe(
		validProbeConfig(),
		time.Second,
		WithTheFeedService(mock),
		WithTries(3),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ip := netip.MustParseAddr("10.0.0.1")
	if _, err := p.Run(context.Background(), ip); err == nil {
		t.Fatal("expected error after exhausting tries, got nil")
	}
	if mock.runCalls != 3 {
		t.Fatalf("runCalls = %d, want 3", mock.runCalls)
	}
}

func TestRun_SuccessStopsEarly(t *testing.T) {
	mock := &mockTheFeedService{}
	p, err := NewTheFeedProbe(
		validProbeConfig(),
		time.Second,
		WithTheFeedService(mock),
		WithTries(3),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ip := netip.MustParseAddr("10.0.0.1")
	if _, err := p.Run(context.Background(), ip); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.runCalls != 1 {
		t.Fatalf("runCalls = %d, want 1", mock.runCalls)
	}
}
