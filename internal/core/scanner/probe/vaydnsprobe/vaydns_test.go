package vaydnsprobe

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"testing"
	"time"

	"bgscan/internal/core/dns"
	"bgscan/internal/core/socks"
	"bgscan/internal/core/speedtest"
)

type mockVayDNSService struct {
	newTunnelFn func(dns.VayDNSConfig, netip.Addr) (net.Conn, error)
}

// EditConfig implements [dns.VayDNSService].
func (m *mockVayDNSService) EditConfig(config dns.VayDNSConfig, originalName string) error {
	return nil
}

// ValidateAllConfigs implements [dns.VayDNSService].
func (m *mockVayDNSService) ValidateAllConfigs() ([]dns.ConfigValidationResult, error) {
	return nil, nil
}

func (m *mockVayDNSService) SaveConfig(dns.VayDNSConfig, string) error { return nil }
func (m *mockVayDNSService) LoadConfig(string) (dns.VayDNSConfig, error) {
	return dns.VayDNSConfig{}, nil
}
func (m *mockVayDNSService) GetAllConfigFiles() ([]dns.VayDNSConfigFile, error) { return nil, nil }
func (m *mockVayDNSService) NewTunnel(ctx context.Context, cfg dns.VayDNSConfig, resolverAddr netip.Addr) (net.Conn, error) {
	return m.newTunnelFn(cfg, resolverAddr)
}

func (f *mockVayDNSService) RenameConfig(oldName, newName string) error {
	return nil
}

type mockSocksService struct {
	connectFn func(ctx context.Context, conn net.Conn, target string, cfg socks.Config) (net.Conn, error)
}

func (m *mockSocksService) Connect(ctx context.Context, conn net.Conn, target string, cfg socks.Config) (net.Conn, error) {
	return m.connectFn(ctx, conn, target, cfg)
}

type mockSpeedtestService struct {
	measureFn func(ctx context.Context, cfg speedtest.LatencyConfig) (speedtest.LatencyResult, error)
}

// MeasureDownloadSpeed implements [speedtest.Service].
func (m *mockSpeedtestService) MeasureDownloadSpeed(context.Context, speedtest.DownloadConfig) (speedtest.SpeedResult, error) {
	return speedtest.SpeedResult{}, nil
}

// MeasureUploadSpeed implements [speedtest.Service].
func (m *mockSpeedtestService) MeasureUploadSpeed(context.Context, speedtest.UploadConfig) (speedtest.SpeedResult, error) {
	return speedtest.SpeedResult{}, nil
}

func (m *mockSpeedtestService) MeasureLatency(ctx context.Context, cfg speedtest.LatencyConfig) (speedtest.LatencyResult, error) {
	return m.measureFn(ctx, cfg)
}

// ---- helpers -----------------------------------------------------------

func validConfig(t *testing.T) dns.VayDNSConfig {
	t.Helper()

	cfg := dns.DefaultVayDNSConfig()
	cfg.Domain = "tunnel.example.com"
	cfg.PubKey = "cd6d78e954f48f62cb74cdcf8a2459d3d39786a7e11fc4f74c04bca86371f748"
	cfg.ProxyPort = 1080
	cfg.AuthMethod = dns.AuthPassword
	cfg.Username = "user"
	cfg.Password = "pass"

	if errs := cfg.Validate(); len(errs) != 0 {
		t.Fatalf("test config is invalid: %v", errs)
	}

	return cfg
}

// netPipeConn returns one end of an in-memory net.Conn pair; the caller
// is responsible for closing both ends if needed.
func netPipeConn() net.Conn {
	client, _ := net.Pipe()
	return client
}

// ---- tests -----------------------------------------------------------

func TestNewVayDNSProbe_InvalidConfig(t *testing.T) {
	_, err := NewVayDNSProbe(dns.VayDNSConfig{}, time.Second)
	if err == nil {
		t.Fatal("expected error for invalid config, got nil")
	}
}

func TestNewVayDNSProbe_DefaultsServicesWhenNotProvided(t *testing.T) {
	p, err := NewVayDNSProbe(validConfig(t), time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	probe, ok := p.(*VayDNSProbe)
	if !ok {
		t.Fatalf("expected *VayDNSProbe, got %T", p)
	}

	if probe.vaydnsService == nil {
		t.Error("expected default vaydnsService to be set")
	}
	if probe.sshService == nil {
		t.Error("expected default sshService to be set")
	}
	if probe.socksService == nil {
		t.Error("expected default socksService to be set")
	}
	if probe.speedtestService == nil {
		t.Error("expected default speedtestService to be set")
	}
}

func TestRun_ContextAlreadyCanceled(t *testing.T) {
	cfg := validConfig(t)
	cfg.ProxyType = dns.ResolverProxySOCKS

	p, err := NewVayDNSProbe(cfg, time.Second)
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

func TestRun_TunnelCreationFails(t *testing.T) {
	cfg := validConfig(t)
	cfg.ProxyType = dns.ResolverProxySOCKS

	wantErr := errors.New("boom")
	p, err := NewVayDNSProbe(
		cfg, time.Second,
		WithVayDNSService(&mockVayDNSService{
			newTunnelFn: func(dns.VayDNSConfig, netip.Addr) (net.Conn, error) {
				return nil, wantErr
			},
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = p.Run(context.Background(), netip.MustParseAddr("2.2.2.2"))
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped %v, got %v", wantErr, err)
	}
}

func TestRun_UnsupportedProxyType(t *testing.T) {
	cfg := validConfig(t)
	cfg.ProxyType = dns.ResolverProxyType("")

	tunnel := netPipeConn()
	defer func() {
		_ = tunnel.Close()
	}()

	p, err := NewVayDNSProbe(
		cfg, time.Second,
		WithVayDNSService(&mockVayDNSService{
			newTunnelFn: func(dns.VayDNSConfig, netip.Addr) (net.Conn, error) {
				return tunnel, nil
			},
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = p.Run(context.Background(), netip.MustParseAddr("2.2.2.2"))
	if err == nil {
		t.Fatal("expected error for unsupported proxy type, got nil")
	}
}

func TestRun_SOCKS_Success(t *testing.T) {
	cfg := validConfig(t)
	cfg.ProxyType = dns.ResolverProxySOCKS
	cfg.AuthMethod = dns.AuthPassword

	tunnelClosed := false
	tunnel := netPipeConn()

	wantLatency := 42 * time.Millisecond
	targetIP := netip.MustParseAddr("2.2.2.2")

	var gotResolverAddr netip.Addr

	p, err := NewVayDNSProbe(
		cfg, time.Second,
		WithVayDNSService(&mockVayDNSService{
			newTunnelFn: func(c dns.VayDNSConfig, r netip.Addr) (net.Conn, error) {
				gotResolverAddr = r
				return &closeTrackingConn{Conn: tunnel, closed: &tunnelClosed}, nil
			},
		}),
		WithSocksService(&mockSocksService{
			connectFn: func(ctx context.Context, conn net.Conn, target string, cfg socks.Config) (net.Conn, error) {
				return conn, nil
			},
		}),
		WithSpeedtestService(&mockSpeedtestService{
			measureFn: func(ctx context.Context, cfg speedtest.LatencyConfig) (speedtest.LatencyResult, error) {
				if cfg.DialContext == nil {
					t.Error("expected DialContext to be set")
				}
				return speedtest.LatencyResult{RTT: wantLatency}, nil
			},
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	res, err := p.Run(context.Background(), targetIP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, ok := res.(VayDNSResult)
	if !ok {
		t.Fatalf("expected VayDNSResult, got %T", res)
	}
	if got.IP != targetIP {
		t.Errorf("IP = %v, want %v", got.IP, targetIP)
	}
	if got.Latency != wantLatency {
		t.Errorf("Latency = %v, want %v", got.Latency, wantLatency)
	}
	if !tunnelClosed {
		t.Error("expected tunnel to be closed after Run")
	}

	// The scanned target must be passed as the DNS resolver address.
	if gotResolverAddr != targetIP {
		t.Errorf(
			"resolverAddr = %s, want %s",
			gotResolverAddr,
			targetIP,
		)
	}
}

func TestRun_SOCKS_KeyAuthRejected(t *testing.T) {
	cfg := validConfig(t)
	cfg.ProxyType = dns.ResolverProxySOCKS
	cfg.AuthMethod = dns.AuthKey

	tunnel := netPipeConn()
	defer func() {
		_ = tunnel.Close()
	}()

	_, err := NewVayDNSProbe(
		cfg, time.Second,
		WithVayDNSService(&mockVayDNSService{
			newTunnelFn: func(dns.VayDNSConfig, netip.Addr) (net.Conn, error) {
				return tunnel, nil
			},
		}),
	)
	if err == nil {
		t.Fatal("expected error for SOCKS + key auth from constructor")
	}
}

func TestRun_SSH_NoAuthRejected(t *testing.T) {
	cfg := validConfig(t)
	cfg.ProxyType = dns.ResolverProxySSH
	cfg.AuthMethod = dns.AuthNone

	tunnel := netPipeConn()
	defer func() {
		_ = tunnel.Close()
	}()

	_, err := NewVayDNSProbe(
		cfg, time.Second,
		WithVayDNSService(&mockVayDNSService{
			newTunnelFn: func(dns.VayDNSConfig, netip.Addr) (net.Conn, error) {
				return tunnel, nil
			},
		}),
	)
	if err == nil {
		t.Fatal("expected error for SSH + no auth from constructor")
	}
}

func TestRun_MeasureLatencyFails(t *testing.T) {
	cfg := validConfig(t)
	cfg.ProxyType = dns.ResolverProxySOCKS

	tunnel := netPipeConn()
	defer func() {
		_ = tunnel.Close()
	}()

	wantErr := errors.New("timeout")

	p, err := NewVayDNSProbe(
		cfg, time.Second,
		WithVayDNSService(&mockVayDNSService{
			newTunnelFn: func(dns.VayDNSConfig, netip.Addr) (net.Conn, error) {
				return tunnel, nil
			},
		}),
		WithSocksService(&mockSocksService{
			connectFn: func(ctx context.Context, conn net.Conn, target string, cfg socks.Config) (net.Conn, error) {
				return conn, nil
			},
		}),
		WithSpeedtestService(&mockSpeedtestService{
			measureFn: func(ctx context.Context, cfg speedtest.LatencyConfig) (speedtest.LatencyResult, error) {
				return speedtest.LatencyResult{}, wantErr
			},
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = p.Run(context.Background(), netip.MustParseAddr("2.2.2.2"))
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped %v, got %v", wantErr, err)
	}
}

func TestClose_NoOp(t *testing.T) {
	p, err := NewVayDNSProbe(validConfig(t), time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := p.Close(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

// closeTrackingConn wraps a net.Conn and records whether Close was called,
// so tests can assert the probe cleans up its tunnel.
type closeTrackingConn struct {
	net.Conn
	closed *bool
}

func (c *closeTrackingConn) Close() error {
	*c.closed = true
	return c.Conn.Close()
}
