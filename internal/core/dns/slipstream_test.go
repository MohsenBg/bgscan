package dns

import (
	"context"
	"errors"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ffislipstream "github.com/MohsenBg/bgscan/internal/core/ffi/slipstream"
)

type fakeStarter struct {
	calls int
	cfg   ffislipstream.Config
	err   error
}

func (f *fakeStarter) Start(cfg ffislipstream.Config) (ffislipstream.Client, error) {
	f.calls++
	f.cfg = cfg

	if f.err != nil {
		return nil, f.err
	}

	return nil, nil
}

func newTestSlipstreamService(
	t *testing.T,
	starter *fakeStarter,
) SlipstreamService {
	t.Helper()

	service, err := NewSlipstreamService(
		WithSlipstreamDir(t.TempDir()),
		WithSlipstreamStarter(starter),
	)
	if err != nil {
		t.Fatalf("NewSlipstreamService: %v", err)
	}

	return service
}

func validSlipstreamConfig() SlipstreamConfig {
	return SlipstreamConfig{
		Domain:            "tunnel.example.com",
		ResolverPort:      53,
		DNSResolution:     DNSResolutionRecursive,
		CongestionControl: CongestionControlBBR,
		GSO:               false,
		KeepAliveInterval: 10,
	}
}

func TestDefaultSlipstreamConfig(t *testing.T) {
	config := DefaultSlipstreamConfig()

	if config.Domain != "" {
		t.Errorf("Domain = %q, want empty", config.Domain)
	}

	if config.ResolverPort != 53 {
		t.Errorf("ResolverPort = %d, want 53", config.ResolverPort)
	}

	if config.CertPath != "" {
		t.Errorf("CertPath = %q, want empty", config.CertPath)
	}

	if config.DNSResolution != DNSResolutionRecursive {
		t.Errorf("DNSResolution = %q, want %q", config.DNSResolution, DNSResolutionRecursive)
	}

	if config.CongestionControl != CongestionControlCubic {
		t.Errorf("CongestionControl = %q, want %q", config.CongestionControl, CongestionControlCubic)
	}

	if config.GSO {
		t.Error("GSO = true, want false")
	}

	if config.KeepAliveInterval == 0 {
		t.Error("KeepAliveInterval = 0, want non-zero default")
	}
}

func TestSlipstreamConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    SlipstreamConfig
		wantError string
	}{
		{
			name:   "valid",
			config: validSlipstreamConfig(),
		},
		{
			name: "valid empty new fields for backward compat",
			config: SlipstreamConfig{
				Domain:       "tunnel.example.com",
				ResolverPort: 53,
			},
		},
		{
			name: "missing domain",
			config: SlipstreamConfig{
				ResolverPort: 53,
			},
			wantError: "domain",
		},
		{
			name: "whitespace domain",
			config: SlipstreamConfig{
				Domain:       "   ",
				ResolverPort: 53,
			},
			wantError: "domain",
		},
		{
			name: "invalid domain",
			config: SlipstreamConfig{
				Domain:       string([]byte{0}),
				ResolverPort: 53,
			},
			wantError: "domain",
		},
		{
			name: "zero DNS port",
			config: SlipstreamConfig{
				Domain:       "tunnel.example.com",
				ResolverPort: 0,
			},
			wantError: "dns_port",
		},
		{
			name: "invalid dns resolution",
			config: func() SlipstreamConfig {
				c := validSlipstreamConfig()
				c.DNSResolution = "round-robin"
				return c
			}(),
			wantError: "dns_resolution",
		},
		{
			name: "invalid congestion control",
			config: func() SlipstreamConfig {
				c := validSlipstreamConfig()
				c.CongestionControl = "reno"
				return c
			}(),
			wantError: "congestion_control",
		},
		{
			name:      "multiple errors",
			config:    SlipstreamConfig{},
			wantError: "multiple",
		},
		{
			name: "international domain",
			config: SlipstreamConfig{
				Domain:       "tunnel.münchen.de",
				ResolverPort: 53,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.config.Validate()

			switch tt.wantError {
			case "":
				if len(errs) != 0 {
					t.Fatalf("Validate() = %v, want no errors", errs)
				}

			case "multiple":
				if _, ok := errs["domain"]; !ok {
					t.Error("expected domain error")
				}

				if _, ok := errs["dns_port"]; !ok {
					t.Error("expected dns_port error")
				}

			default:
				if _, ok := errs[tt.wantError]; !ok {
					t.Errorf(
						"expected %q error, got %v",
						tt.wantError,
						errs,
					)
				}
			}
		})
	}
}

func TestDNSResolutionIsValid(t *testing.T) {
	for _, valid := range []DNSResolution{"", DNSResolutionRecursive, DNSResolutionAuthoritative} {
		if !valid.IsValid() {
			t.Errorf("IsValid(%q) = false, want true", valid)
		}
	}

	if DNSResolution("bogus").IsValid() {
		t.Error("IsValid(bogus) = true, want false")
	}
}

func TestCongestionControlIsValid(t *testing.T) {
	for _, valid := range []CongestionControl{"", CongestionControlBBR, CongestionControlCubic} {
		if !valid.IsValid() {
			t.Errorf("IsValid(%q) = false, want true", valid)
		}
	}

	if CongestionControl("reno").IsValid() {
		t.Error("IsValid(reno) = true, want false")
	}
}

func TestDNSResolutionLibValue(t *testing.T) {
	if got := DNSResolutionAuthoritative.libValue(); got != ffislipstream.ModeAuthoritative {
		t.Errorf("authoritative libValue = %v, want %v", got, ffislipstream.ModeAuthoritative)
	}

	if got := DNSResolutionRecursive.libValue(); got != ffislipstream.ModeRecursive {
		t.Errorf("recursive libValue = %v, want %v", got, ffislipstream.ModeRecursive)
	}

	if got := DNSResolution("").libValue(); got != ffislipstream.ModeRecursive {
		t.Errorf("empty libValue = %v, want recursive %v", got, ffislipstream.ModeRecursive)
	}
}

func TestNewSlipstreamService(t *testing.T) {
	starter := &fakeStarter{}

	service, err := NewSlipstreamService(
		WithSlipstreamDir(t.TempDir()),
		WithSlipstreamStarter(starter),
	)
	if err != nil {
		t.Fatalf("NewSlipstreamService: %v", err)
	}

	if service == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestNewSlipstreamServiceNilStarter(t *testing.T) {
	starter := &fakeStarter{}

	// A nil starter must not clear a previously injected one.
	_, err := NewSlipstreamService(
		WithSlipstreamDir(t.TempDir()),
		WithSlipstreamStarter(starter),
		WithSlipstreamStarter(nil),
	)
	if err != nil {
		t.Fatalf("NewSlipstreamService: %v", err)
	}
}

func TestRunTunnel(t *testing.T) {
	starter := &fakeStarter{}

	service := newTestSlipstreamService(t, starter)

	config := validSlipstreamConfig()

	_, err := service.RunTunnel(
		context.Background(),
		config,
		netip.MustParseAddr("1.2.3.4"),
		5300,
		0,
	)
	if err != nil {
		t.Fatalf("RunTunnel: %v", err)
	}

	if starter.calls != 1 {
		t.Fatalf("start calls = %d, want 1", starter.calls)
	}

	got := starter.cfg

	if len(got.Resolvers) != 1 {
		t.Fatalf("resolvers = %d, want 1", len(got.Resolvers))
	}

	if got.Resolvers[0].Host != "1.2.3.4" {
		t.Errorf("resolver host = %q, want %q", got.Resolvers[0].Host, "1.2.3.4")
	}

	if got.Resolvers[0].Port != 53 {
		t.Errorf("resolver port = %d, want 53", got.Resolvers[0].Port)
	}

	if got.Resolvers[0].Mode != ffislipstream.ModeRecursive {
		t.Errorf("resolver mode = %v, want recursive", got.Resolvers[0].Mode)
	}

	if got.ListenHost != "127.0.0.1" {
		t.Errorf("listen host = %q, want 127.0.0.1", got.ListenHost)
	}

	if got.ListenPort != 5300 {
		t.Errorf("listen port = %d, want 5300", got.ListenPort)
	}

	if got.Domain != "tunnel.example.com" {
		t.Errorf("domain = %q, want tunnel.example.com", got.Domain)
	}

	if got.CongestionControl != string(CongestionControlBBR) {
		t.Errorf("congestion = %q, want bbr", got.CongestionControl)
	}

	if got.GSO {
		t.Error("GSO = true, want false")
	}

	// Zero keepalive falls back to the config value.
	if got.KeepAliveInterval != 10 {
		t.Errorf("keepalive = %d, want 10 (config fallback)", got.KeepAliveInterval)
	}
}

func TestRunTunnelKeepAliveOverride(t *testing.T) {
	starter := &fakeStarter{}
	service := newTestSlipstreamService(t, starter)

	config := validSlipstreamConfig()
	config.KeepAliveInterval = 10

	_, err := service.RunTunnel(
		context.Background(),
		config,
		netip.MustParseAddr("1.2.3.4"),
		5300,
		30,
	)
	if err != nil {
		t.Fatalf("RunTunnel: %v", err)
	}

	if starter.cfg.KeepAliveInterval != 30 {
		t.Errorf("keepalive = %d, want 30 (explicit override)", starter.cfg.KeepAliveInterval)
	}
}

func TestRunTunnelIPv6ListenHost(t *testing.T) {
	starter := &fakeStarter{}
	service := newTestSlipstreamService(t, starter)

	_, err := service.RunTunnel(
		context.Background(),
		validSlipstreamConfig(),
		netip.MustParseAddr("2001:db8::1"),
		5300,
		0,
	)
	if err != nil {
		t.Fatalf("RunTunnel: %v", err)
	}

	if starter.cfg.ListenHost != "::1" {
		t.Errorf("listen host = %q, want ::1", starter.cfg.ListenHost)
	}
}

func TestRunTunnelWithCert(t *testing.T) {
	starter := &fakeStarter{}
	service := newTestSlipstreamService(t, starter)

	config := validSlipstreamConfig()
	config.CertPath = "/etc/slipstream/ca.pem"

	_, err := service.RunTunnel(
		context.Background(),
		config,
		netip.MustParseAddr("1.2.3.4"),
		5300,
		0,
	)
	if err != nil {
		t.Fatalf("RunTunnel: %v", err)
	}

	if starter.cfg.CertPath != "/etc/slipstream/ca.pem" {
		t.Errorf("cert = %q, want /etc/slipstream/ca.pem", starter.cfg.CertPath)
	}
}

func TestRunTunnelAuthoritativeAndGSO(t *testing.T) {
	starter := &fakeStarter{}
	service := newTestSlipstreamService(t, starter)

	config := validSlipstreamConfig()
	config.DNSResolution = DNSResolutionAuthoritative
	config.CongestionControl = CongestionControlCubic
	config.GSO = true

	_, err := service.RunTunnel(
		context.Background(),
		config,
		netip.MustParseAddr("1.2.3.4"),
		5300,
		0,
	)
	if err != nil {
		t.Fatalf("RunTunnel: %v", err)
	}

	if starter.cfg.Resolvers[0].Mode != ffislipstream.ModeAuthoritative {
		t.Errorf("mode = %v, want authoritative", starter.cfg.Resolvers[0].Mode)
	}

	if starter.cfg.CongestionControl != "cubic" {
		t.Errorf("congestion = %q, want cubic", starter.cfg.CongestionControl)
	}

	if !starter.cfg.GSO {
		t.Error("GSO = false, want true")
	}
}

func TestRunTunnelCustomDNSPort(t *testing.T) {
	starter := &fakeStarter{}
	service := newTestSlipstreamService(t, starter)

	config := validSlipstreamConfig()
	config.ResolverPort = 5353

	_, err := service.RunTunnel(
		context.Background(),
		config,
		netip.MustParseAddr("10.0.0.1"),
		5300,
		0,
	)
	if err != nil {
		t.Fatalf("RunTunnel: %v", err)
	}

	if starter.cfg.Resolvers[0].Host != "10.0.0.1" {
		t.Errorf("resolver host = %q, want 10.0.0.1", starter.cfg.Resolvers[0].Host)
	}

	if starter.cfg.Resolvers[0].Port != 5353 {
		t.Errorf("resolver port = %d, want 5353", starter.cfg.Resolvers[0].Port)
	}
}

func TestRunTunnelInvalidConfig(t *testing.T) {
	starter := &fakeStarter{}
	service := newTestSlipstreamService(t, starter)

	config := SlipstreamConfig{}

	_, err := service.RunTunnel(
		context.Background(),
		config,
		netip.MustParseAddr("1.2.3.4"),
		5300,
		0,
	)
	if err == nil {
		t.Fatal("expected validation error")
	}

	if starter.calls != 0 {
		t.Fatal("library should not start with invalid config")
	}
}

func TestRunTunnelInvalidResolverIP(t *testing.T) {
	starter := &fakeStarter{}
	service := newTestSlipstreamService(t, starter)

	_, err := service.RunTunnel(
		context.Background(),
		validSlipstreamConfig(),
		netip.Addr{},
		5300,
		0,
	)
	if err == nil {
		t.Fatal("expected resolver IP error")
	}

	if starter.calls != 0 {
		t.Fatal("library should not start without resolver IP")
	}
}

func TestRunTunnelZeroListenPort(t *testing.T) {
	starter := &fakeStarter{}
	service := newTestSlipstreamService(t, starter)

	_, err := service.RunTunnel(
		context.Background(),
		validSlipstreamConfig(),
		netip.MustParseAddr("1.2.3.4"),
		0,
		0,
	)
	if err == nil {
		t.Fatal("expected listen port error")
	}

	if starter.calls != 0 {
		t.Fatal("library should not start with zero listen port")
	}
}

func TestRunTunnelStarterError(t *testing.T) {
	sentinel := errors.New("start failed")

	starter := &fakeStarter{
		err: sentinel,
	}

	service := newTestSlipstreamService(t, starter)

	_, err := service.RunTunnel(
		context.Background(),
		validSlipstreamConfig(),
		netip.MustParseAddr("1.2.3.4"),
		5300,
		0,
	)
	if err == nil {
		t.Fatal("expected starter error")
	}

	if !errors.Is(err, sentinel) {
		t.Errorf(
			"error = %v, want wrapped %v",
			err,
			sentinel,
		)
	}
}

func TestRunTunnelCanceledContext(t *testing.T) {
	starter := &fakeStarter{}
	service := newTestSlipstreamService(t, starter)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.RunTunnel(
		ctx,
		validSlipstreamConfig(),
		netip.MustParseAddr("1.2.3.4"),
		5300,
		0,
	)
	if err == nil {
		t.Fatal("expected context error")
	}

	if starter.calls != 0 {
		t.Fatal("library should not start with canceled context")
	}
}

func TestRunTunnelNilLib(t *testing.T) {
	service := &slipstreamService{
		configs: newConfigStore[SlipstreamConfig](t.TempDir(), "Slipstream"),
		slip:    nil,
	}

	_, err := service.RunTunnel(
		context.Background(),
		validSlipstreamConfig(),
		netip.MustParseAddr("1.2.3.4"),
		5300,
		0,
	)
	if err == nil {
		t.Fatal("expected library-not-loaded error")
	}
}

func TestLibConfig(t *testing.T) {
	config := validSlipstreamConfig()
	config.ResolverPort = 5353
	config.CertPath = "/tmp/ca.pem"

	got := config.LibConfig(
		netip.MustParseAddr("9.9.9.9"),
		netip.MustParseAddr("127.0.0.1"),
		5400,
		7,
	)

	if got.Domain != "tunnel.example.com" {
		t.Errorf("domain = %q", got.Domain)
	}

	if len(got.Resolvers) != 1 || got.Resolvers[0].Host != "9.9.9.9" || got.Resolvers[0].Port != 5353 {
		t.Errorf("resolvers = %+v", got.Resolvers)
	}

	if got.ListenHost != "127.0.0.1" || got.ListenPort != 5400 {
		t.Errorf("listen = %s:%d", got.ListenHost, got.ListenPort)
	}

	if got.KeepAliveInterval != 7 {
		t.Errorf("keepalive = %d, want 7", got.KeepAliveInterval)
	}

	if got.CertPath != "/tmp/ca.pem" {
		t.Errorf("cert = %q", got.CertPath)
	}
}

func TestLibConfigNil(t *testing.T) {
	var config *SlipstreamConfig

	got := config.LibConfig(
		netip.MustParseAddr("1.1.1.1"),
		netip.MustParseAddr("127.0.0.1"),
		5300,
		5,
	)

	if len(got.Resolvers) != 0 {
		t.Errorf("expected no resolvers, got %+v", got.Resolvers)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	dir := resolvedSlipstreamDir(t)

	service := newTestSlipstreamService(
		t,
		&fakeStarter{},
	)

	config := SlipstreamConfig{
		Domain:            "roundtrip.example.com",
		ResolverPort:      853,
		CertPath:          "/tmp/cert.pem",
		DNSResolution:     DNSResolutionAuthoritative,
		GSO:               true,
		CongestionControl: CongestionControlCubic,
		KeepAliveInterval: 15,
	}

	const name = "test-roundtrip"

	t.Cleanup(func() {
		_ = os.Remove(
			filepath.Join(dir, name+".toml"),
		)
	})

	if err := service.SaveConfig(config, name); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	got, err := service.LoadConfig(name)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	// KeepAliveInterval is tagged toml:"-" (runtime-only), so it is not
	// persisted. Compare everything else, then assert the loaded value is
	// the zero value.
	want := config
	want.KeepAliveInterval = 0

	if got != want {
		t.Errorf(
			"loaded config = %#v, want %#v",
			got,
			want,
		)
	}
}

func TestSaveConfigEmptyName(t *testing.T) {
	service := newTestSlipstreamService(
		t,
		&fakeStarter{},
	)

	err := service.SaveConfig(
		validSlipstreamConfig(),
		"",
	)
	if err == nil {
		t.Fatal("expected error for empty config name")
	}
}

func TestSaveConfigWhitespaceName(t *testing.T) {
	service := newTestSlipstreamService(
		t,
		&fakeStarter{},
	)

	err := service.SaveConfig(
		validSlipstreamConfig(),
		"   ",
	)
	if err == nil {
		t.Fatal("expected error for whitespace config name")
	}
}

func TestSaveConfigInvalidConfig(t *testing.T) {
	service := newTestSlipstreamService(
		t,
		&fakeStarter{},
	)

	err := service.SaveConfig(
		SlipstreamConfig{},
		"invalid-config",
	)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSlipstreamLoadConfigNotFound(t *testing.T) {
	service := newTestSlipstreamService(
		t,
		&fakeStarter{},
	)

	_, err := service.LoadConfig(
		"config-that-does-not-exist",
	)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestSlipstreamGetAllConfigFiles(t *testing.T) {
	service := newTestSlipstreamService(t, &fakeStarter{})

	names := []string{
		"test-list-alpha",
		"test-list-beta",
	}

	for _, name := range names {
		if err := service.SaveConfig(validSlipstreamConfig(), name); err != nil {
			t.Fatalf("SaveConfig(%q): %v", name, err)
		}
	}

	files, err := service.GetAllConfigFiles()
	if err != nil {
		t.Fatalf("GetAllConfigFiles: %v", err)
	}

	found := make(map[string]bool)
	for _, file := range files {
		found[file.Name] = true
	}

	for _, name := range names {
		if !found[name] {
			t.Errorf("config %q was not returned", name)
		}
	}
}

func TestGetAllConfigFilesIgnoresNonTOML(t *testing.T) {
	dir := resolvedSlipstreamDir(t)

	junk := filepath.Join(
		dir,
		"test-slipstream-junk.json",
	)

	if err := os.WriteFile(
		junk,
		[]byte("{}"),
		0o600,
	); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Remove(junk)
	})

	service := newTestSlipstreamService(
		t,
		&fakeStarter{},
	)

	files, err := service.GetAllConfigFiles()
	if err != nil {
		t.Fatalf("GetAllConfigFiles: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name) != ".toml" {
			t.Errorf(
				"non-TOML file returned: %q",
				file.Name,
			)
		}
	}
}

func TestConfigPathNormalizesExtension(t *testing.T) {
	service := &slipstreamService{
		configs: newConfigStore[SlipstreamConfig]("/cfg", "Slipstream"),
	}

	got := service.configs.configPath("test.toml")
	want := filepath.Join("/cfg", "test.toml")

	if got != want {
		t.Errorf(
			"configPath() = %q, want %q",
			got,
			want,
		)
	}
}

func TestConfigPathAddsExtension(t *testing.T) {
	service := &slipstreamService{
		configs: newConfigStore[SlipstreamConfig]("/cfg", "Slipstream"),
	}

	got := service.configs.configPath("test")
	want := filepath.Join("/cfg", "test.toml")

	if got != want {
		t.Errorf(
			"configPath() = %q, want %q",
			got,
			want,
		)
	}
}

func resolvedSlipstreamDir(t *testing.T) string {
	t.Helper()

	dir := tunnelConfigDir("slipstream")

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf(
			"create Slipstream config directory: %v",
			err,
		)
	}

	return dir
}

func TestSlipstreamEditConfigUpdatesExisting(t *testing.T) {
	service := newTestSlipstreamService(t, &fakeStarter{})

	if err := service.SaveConfig(validSlipstreamConfig(), "my-tunnel"); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	updated := validSlipstreamConfig()
	updated.Domain = "updated.example.com"

	if err := service.EditConfig(updated, "my-tunnel"); err != nil {
		t.Fatalf("EditConfig() error = %v", err)
	}

	got, err := service.LoadConfig("my-tunnel")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if got.Domain != "updated.example.com" {
		t.Fatalf("Domain = %q, want %q", got.Domain, "updated.example.com")
	}
}

func TestSlipstreamEditConfigMissingConfigReturnsError(t *testing.T) {
	service := newTestSlipstreamService(t, &fakeStarter{})

	err := service.EditConfig(validSlipstreamConfig(), "does-not-exist")
	if err == nil {
		t.Fatal("EditConfig() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("EditConfig() error = %q, want does not exist", err)
	}
}
