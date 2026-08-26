package dns

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bgscan/internal/core/process"
)

type mockProcess struct {
	killed bool
	waited bool
}

func (m *mockProcess) StopGracefully(time.Duration) error {
	return nil
}

func (m *mockProcess) Kill() error {
	m.killed = true
	return nil
}

func (m *mockProcess) Wait() error {
	m.waited = true
	return nil
}

func (m *mockProcess) Pid() int {
	return 9999
}

func (m *mockProcess) String() string {
	return "mock-process(9999)"
}

type captureStarter struct {
	calls []startCall
	err   error
	proc  process.Process
}

type startCall struct {
	bin  string
	args []string
}

func (c *captureStarter) Start(
	_ context.Context,
	bin string,
	args ...string,
) (process.Process, error) {
	c.calls = append(c.calls, startCall{
		bin:  bin,
		args: append([]string(nil), args...),
	})

	if c.err != nil {
		return nil, c.err
	}

	if c.proc != nil {
		return c.proc, nil
	}

	return &mockProcess{}, nil
}

func newTestSlipstreamService(
	t *testing.T,
	starter *captureStarter,
) SlipstreamService {
	t.Helper()

	service, err := NewSlipstreamService(
		WithSlipstreamDir(t.TempDir()),
		WithSlipstreamClientBinary("/usr/local/bin/slipstream-client"),
		WithSlipstreamProcessStarter(starter.Start),
	)
	if err != nil {
		t.Fatalf("NewSlipstreamService: %v", err)
	}

	return service
}

func validSlipstreamConfig() SlipstreamConfig {
	return SlipstreamConfig{
		Domain:       "tunnel.example.com",
		ResolverPort: 53,
	}
}

func TestDefaultSlipstreamConfig(t *testing.T) {
	config := DefaultSlipstreamConfig()

	if config.Domain != "" {
		t.Errorf("Domain = %q, want empty", config.Domain)
	}

	if config.ResolverPort != 53 {
		t.Errorf("DNSPort = %d, want 53", config.ResolverPort)
	}

	if config.CertPath != "" {
		t.Errorf("CertPath = %q, want empty", config.CertPath)
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

func TestNewSlipstreamService(t *testing.T) {
	starter := &captureStarter{}

	service, err := NewSlipstreamService(
		WithSlipstreamClientBinary("/fake/slipstream-client"),
		WithSlipstreamProcessStarter(starter.Start),
	)
	if err != nil {
		t.Fatalf("NewSlipstreamService: %v", err)
	}

	if service == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestNewSlipstreamServiceNilStarter(t *testing.T) {
	_, err := NewSlipstreamService(
		WithSlipstreamClientBinary("/fake/slipstream-client"),
		WithSlipstreamProcessStarter(nil),
	)
	if err != nil {
		t.Fatalf("NewSlipstreamService: %v", err)
	}
}

func TestRunTunnel(t *testing.T) {
	starter := &captureStarter{
		proc: &mockProcess{},
	}

	service := newTestSlipstreamService(t, starter)

	config := validSlipstreamConfig()

	gotProc, err := service.RunTunnel(
		context.Background(),
		config,
		"1.2.3.4",
		5300,
	)
	if err != nil {
		t.Fatalf("RunTunnel: %v", err)
	}

	if gotProc != starter.proc {
		t.Fatalf("process = %v, want %v", gotProc, starter.proc)
	}

	if len(starter.calls) != 1 {
		t.Fatalf(
			"start calls = %d, want 1",
			len(starter.calls),
		)
	}

	call := starter.calls[0]

	if call.bin != "/usr/local/bin/slipstream-client" {
		t.Errorf(
			"binary = %q, want %q",
			call.bin,
			"/usr/local/bin/slipstream-client",
		)
	}

	wantArgs := []string{
		"-d", "tunnel.example.com",
		"-r", "1.2.3.4:53",
		"-l", "5300",
	}

	assertArgs(t, call.args, wantArgs)
}

func TestRunTunnelWithCert(t *testing.T) {
	starter := &captureStarter{}
	service := newTestSlipstreamService(t, starter)

	config := validSlipstreamConfig()
	config.CertPath = "/etc/slipstream/ca.pem"

	_, err := service.RunTunnel(
		context.Background(),
		config,
		"1.2.3.4",
		5300,
	)
	if err != nil {
		t.Fatalf("RunTunnel: %v", err)
	}

	wantArgs := []string{
		"-d", "tunnel.example.com",
		"-r", "1.2.3.4:53",
		"-l", "5300",
		"--cert", "/etc/slipstream/ca.pem",
	}

	assertArgs(t, starter.calls[0].args, wantArgs)
}

func TestRunTunnelCustomDNSPort(t *testing.T) {
	starter := &captureStarter{}
	service := newTestSlipstreamService(t, starter)

	config := validSlipstreamConfig()
	config.ResolverPort = 5353

	_, err := service.RunTunnel(
		context.Background(),
		config,
		"10.0.0.1",
		5300,
	)
	if err != nil {
		t.Fatalf("RunTunnel: %v", err)
	}

	wantArgs := []string{
		"-d", "tunnel.example.com",
		"-r", "10.0.0.1:5353",
		"-l", "5300",
	}

	assertArgs(t, starter.calls[0].args, wantArgs)
}

func TestRunTunnelInvalidConfig(t *testing.T) {
	starter := &captureStarter{}
	service := newTestSlipstreamService(t, starter)

	config := SlipstreamConfig{}

	_, err := service.RunTunnel(
		context.Background(),
		config,
		"1.2.3.4",
		5300,
	)
	if err == nil {
		t.Fatal("expected validation error")
	}

	if len(starter.calls) != 0 {
		t.Fatal("process should not start with invalid config")
	}
}

func TestRunTunnelEmptyResolverIP(t *testing.T) {
	starter := &captureStarter{}
	service := newTestSlipstreamService(t, starter)

	_, err := service.RunTunnel(
		context.Background(),
		validSlipstreamConfig(),
		"",
		5300,
	)
	if err == nil {
		t.Fatal("expected resolver IP error")
	}

	if len(starter.calls) != 0 {
		t.Fatal("process should not start without resolver IP")
	}
}

func TestRunTunnelWhitespaceResolverIP(t *testing.T) {
	starter := &captureStarter{}
	service := newTestSlipstreamService(t, starter)

	_, err := service.RunTunnel(
		context.Background(),
		validSlipstreamConfig(),
		"   ",
		5300,
	)
	if err == nil {
		t.Fatal("expected resolver IP error")
	}

	if len(starter.calls) != 0 {
		t.Fatal("process should not start with invalid resolver IP")
	}
}

func TestRunTunnelZeroListenPort(t *testing.T) {
	starter := &captureStarter{}
	service := newTestSlipstreamService(t, starter)

	_, err := service.RunTunnel(
		context.Background(),
		validSlipstreamConfig(),
		"1.2.3.4",
		0,
	)
	if err == nil {
		t.Fatal("expected listen port error")
	}

	if len(starter.calls) != 0 {
		t.Fatal("process should not start with zero listen port")
	}
}

func TestRunTunnelStarterError(t *testing.T) {
	sentinel := errors.New("start failed")

	starter := &captureStarter{
		err: sentinel,
	}

	service := newTestSlipstreamService(t, starter)

	_, err := service.RunTunnel(
		context.Background(),
		validSlipstreamConfig(),
		"1.2.3.4",
		5300,
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

func TestSaveAndLoadConfig(t *testing.T) {
	dir := resolvedSlipstreamDir(t)

	service := newTestSlipstreamService(
		t,
		&captureStarter{},
	)

	config := SlipstreamConfig{
		Domain:       "roundtrip.example.com",
		ResolverPort: 853,
		CertPath:     "/tmp/cert.pem",
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

	if got != config {
		t.Errorf(
			"loaded config = %#v, want %#v",
			got,
			config,
		)
	}
}

func TestSaveConfigEmptyName(t *testing.T) {
	service := newTestSlipstreamService(
		t,
		&captureStarter{},
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
		&captureStarter{},
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
		&captureStarter{},
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
		&captureStarter{},
	)

	_, err := service.LoadConfig(
		"config-that-does-not-exist",
	)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestSlipstreamGetAllConfigFiles(t *testing.T) {
	service := newTestSlipstreamService(t, &captureStarter{})

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
		&captureStarter{},
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
	service := &slipstreamService{dir: "/cfg"}

	got := service.configPath("test.toml")
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
	service := &slipstreamService{dir: "/cfg"}

	got := service.configPath("test")
	want := filepath.Join("/cfg", "test.toml")

	if got != want {
		t.Errorf(
			"configPath() = %q, want %q",
			got,
			want,
		)
	}
}

func TestSlipstreamClientPaths(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	want := []string{
		filepath.Join(wd, "assets", "slipstream-client"),
		filepath.Join(wd, "assets", "slipstream", "slipstream-client"),
		filepath.Join(wd, "slipstream-client"),
		wd,
	}

	got := getSlipstreamPaths()

	if len(got) != len(want) {
		t.Fatalf(
			"getSlipstreamPaths() length = %d, want %d",
			len(got),
			len(want),
		)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf(
				"getSlipstreamPaths()[%d] = %q, want %q",
				i,
				got[i],
				want[i],
			)
		}
	}
}

func assertArgs(
	t *testing.T,
	got []string,
	want []string,
) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf(
			"args length = %d, want %d\n got: %#v\nwant: %#v",
			len(got),
			len(want),
			got,
			want,
		)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf(
				"args[%d] = %q, want %q",
				i,
				got[i],
				want[i],
			)
		}
	}
}

func resolvedSlipstreamDir(t *testing.T) string {
	t.Helper()

	dir := getSlipstreamDir()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf(
			"create Slipstream config directory: %v",
			err,
		)
	}

	return dir
}

func TestSlipstreamEditConfigUpdatesExisting(t *testing.T) {
	service := newTestSlipstreamService(t, &captureStarter{})

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
	service := newTestSlipstreamService(t, &captureStarter{})

	err := service.EditConfig(validSlipstreamConfig(), "does-not-exist")
	if err == nil {
		t.Fatal("EditConfig() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("EditConfig() error = %q, want does not exist", err)
	}
}
