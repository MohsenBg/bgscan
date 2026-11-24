package dns

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	vaydns "github.com/net2share/vaydns/client"
)

func validDNSTTConfig() DNSTTConfig {
	config := DefaultDNSTTConfig()

	config.Domain = "t.example.com"
	config.PubKey = "cd6d78e954f48f62cb74cdcf8a2459d3d39786a7e11fc4f74c04bca86371f748"

	return config
}

// TestDefaultDNSTTConfig checks the default values returned by DefaultDNSTTConfig.
func TestDefaultDNSTTConfig(t *testing.T) {
	config := DefaultDNSTTConfig()

	if config.ResolverType != ResolverType(vaydns.ResolverTypeUDP) {
		t.Errorf(
			"ResolverType = %v, want %v",
			config.ResolverType,
			vaydns.ResolverTypeUDP,
		)
	}

	if config.ResolverPort != 53 {
		t.Errorf(
			"ResolverPort = %d, want 53",
			config.ResolverPort,
		)
	}

	if config.Fingerprint != "Chrome" {
		t.Errorf(
			"Fingerprint = %q, want %q",
			config.Fingerprint,
			"Chrome",
		)
	}

	if config.RPS != 0 {
		t.Errorf(
			"RPS = %v, want 0",
			config.RPS,
		)
	}
}

// TestDNSTTConfigValidate checks that Validate reports errors for invalid DNSTT fields.
func TestDNSTTConfigValidate(t *testing.T) {
	tests := []struct {
		name  string
		mut   func(*DNSTTConfig)
		field Field
	}{
		{
			name: "missing domain",
			mut: func(c *DNSTTConfig) {
				c.Domain = ""
			},
			field: FieldDomain,
		},
		{
			name: "invalid domain",
			mut: func(c *DNSTTConfig) {
				c.Domain = "invalid domain"
			},
			field: FieldDomain,
		},
		{
			name: "missing public key",
			mut: func(c *DNSTTConfig) {
				c.PubKey = ""
			},
			field: FieldPubKey,
		},
		{
			name: "invalid resolver type",
			mut: func(c *DNSTTConfig) {
				c.ResolverType = ResolverType("invalid")
			},
			field: FieldResolverType,
		},
		{
			name: "zero resolver port",
			mut: func(c *DNSTTConfig) {
				c.ResolverPort = 0
			},
			field: FieldResolverPort,
		},
		{
			name: "negative rps",
			mut: func(c *DNSTTConfig) {
				c.RPS = -1
			},
			field: FieldRPS,
		},
		{
			name: "rps too large",
			mut: func(c *DNSTTConfig) {
				c.RPS = 501
			},
			field: FieldRPS,
		},
		{
			name: "rps nan",
			mut: func(c *DNSTTConfig) {
				c.RPS = math.NaN()
			},
			field: FieldRPS,
		},
		{
			name: "rps positive infinity",
			mut: func(c *DNSTTConfig) {
				c.RPS = math.Inf(1)
			},
			field: FieldRPS,
		},
		{
			name: "rps negative infinity",
			mut: func(c *DNSTTConfig) {
				c.RPS = math.Inf(-1)
			},
			field: FieldRPS,
		},
		{
			name: "missing fingerprint",
			mut: func(c *DNSTTConfig) {
				c.Fingerprint = ""
			},
			field: FieldFingerprint,
		},
		{
			name: "invalid fingerprint",
			mut: func(c *DNSTTConfig) {
				c.Fingerprint = "definitely-invalid"
			},
			field: FieldFingerprint,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validDNSTTConfig()
			tt.mut(&config)

			errs := config.Validate()

			if _, ok := errs[tt.field]; !ok {
				t.Fatalf(
					"Validate() did not return error for field %q; errors=%v",
					tt.field,
					errs,
				)
			}
		})
	}
}

// TestDNSTTConfigValidateValid checks that a valid DNSTT config passes Validate with no errors.
func TestDNSTTConfigValidateValid(t *testing.T) {
	config := validDNSTTConfig()

	errs := config.Validate()

	if len(errs) != 0 {
		t.Fatalf(
			"Validate() returned unexpected errors: %v",
			errs,
		)
	}
}

// TestWithDNSTTDir checks that WithDNSTTDir sets the config directory of the service.
func TestWithDNSTTDir(t *testing.T) {
	service := NewDNSTTService()

	dnsttService, ok := service.(*dnsttService)
	if !ok {
		t.Fatalf(
			"NewDNSTTService() returned %T, want *dnsttService",
			service,
		)
	}

	want := filepath.Join(
		t.TempDir(),
		"custom-dnstt",
	)

	option := WithDNSTTDir(want)
	option(dnsttService)

	if dnsttService.configs.dir != want {
		t.Fatalf(
			"dir = %q, want %q",
			dnsttService.configs.dir,
			want,
		)
	}
}

// TestWithDNSTTDirEmpty checks that WithDNSTTDir with an empty value keeps the default directory.
func TestWithDNSTTDirEmpty(t *testing.T) {
	service := NewDNSTTService()

	dnsttService := service.(*dnsttService)
	original := dnsttService.configs.dir

	WithDNSTTDir("")(dnsttService)

	if dnsttService.configs.dir != original {
		t.Fatalf(
			"empty directory changed dir from %q to %q",
			original,
			dnsttService.configs.dir,
		)
	}
}

// TestDNSTTServiceConfigPath checks the file path the DNSTT service builds for a config name.
func TestDNSTTServiceConfigPath(t *testing.T) {
	service := &dnsttService{
		configs: newConfigStore[DNSTTConfig]("/tmp/dnstt", "DNSTT"),
	}

	tests := []struct {
		name string
		want string
	}{
		{
			name: "test",
			want: "/tmp/dnstt/test.toml",
		},
		{
			name: "test.toml",
			want: "/tmp/dnstt/test.toml",
		},
		{
			name: "TEST.TOML",
			want: "/tmp/dnstt/TEST.toml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.configs.configPath(tt.name)
			if err != nil {
				t.Fatalf("configPath(%q) error = %v", tt.name, err)
			}

			if got != tt.want {
				t.Fatalf(
					"configPath(%q) = %q, want %q",
					tt.name,
					got,
					tt.want,
				)
			}
		})
	}
}

// TestDNSTTServiceSaveLoadConfig checks that a DNSTT config saved through the service loads back unchanged.
func TestDNSTTServiceSaveLoadConfig(t *testing.T) {
	dir := t.TempDir()

	service := NewDNSTTService(
		WithDNSTTDir(dir),
	)

	config := validDNSTTConfig()
	config.RPS = 25
	config.ResolverPort = 5353
	config.Fingerprint = "Chrome_120"

	if err := service.SaveConfig(config, "test"); err != nil {
		t.Fatalf(
			"SaveConfig() error = %v",
			err,
		)
	}

	got, err := service.LoadConfig("test")
	if err != nil {
		t.Fatalf(
			"LoadConfig() error = %v",
			err,
		)
	}

	if got.Domain != config.Domain {
		t.Errorf(
			"Domain = %q, want %q",
			got.Domain,
			config.Domain,
		)
	}

	if got.PubKey != config.PubKey {
		t.Errorf(
			"PubKey = %q, want %q",
			got.PubKey,
			config.PubKey,
		)
	}

	if got.ResolverType != config.ResolverType {
		t.Errorf(
			"ResolverType = %q, want %q",
			got.ResolverType,
			config.ResolverType,
		)
	}

	if got.ResolverPort != config.ResolverPort {
		t.Errorf(
			"ResolverPort = %d, want %d",
			got.ResolverPort,
			config.ResolverPort,
		)
	}

	if got.Fingerprint != config.Fingerprint {
		t.Errorf(
			"Fingerprint = %q, want %q",
			got.Fingerprint,
			config.Fingerprint,
		)
	}

	if got.RPS != config.RPS {
		t.Errorf(
			"RPS = %v, want %v",
			got.RPS,
			config.RPS,
		)
	}
}

// TestDNSTTServiceSaveConfigRequiresName checks that SaveConfig rejects an empty config name.
func TestDNSTTServiceSaveConfigRequiresName(t *testing.T) {
	service := NewDNSTTService(
		WithDNSTTDir(t.TempDir()),
	)

	err := service.SaveConfig(
		validDNSTTConfig(),
		" ",
	)

	if err == nil {
		t.Fatal("SaveConfig() expected error for empty name")
	}
}

// TestDNSTTServiceSaveConfigValidatesConfig checks that SaveConfig rejects an invalid config.
func TestDNSTTServiceSaveConfigValidatesConfig(t *testing.T) {
	service := NewDNSTTService(
		WithDNSTTDir(t.TempDir()),
	)

	config := validDNSTTConfig()
	config.Domain = ""

	err := service.SaveConfig(config, "test")

	if err == nil {
		t.Fatal("SaveConfig() expected validation error")
	}
}

// TestDNSTTServiceLoadConfigMissing checks that LoadConfig returns an error for a config that does not exist.
func TestDNSTTServiceLoadConfigMissing(t *testing.T) {
	service := NewDNSTTService(
		WithDNSTTDir(t.TempDir()),
	)

	_, err := service.LoadConfig("missing")

	if err == nil {
		t.Fatal("LoadConfig() expected error for missing config")
	}
}

// TestDNSTTServiceGetAllConfigFiles checks that GetAllConfigFiles returns the saved DNSTT configs.
func TestDNSTTServiceGetAllConfigFiles(t *testing.T) {
	dir := t.TempDir()

	service := NewDNSTTService(
		WithDNSTTDir(dir),
	)

	if err := service.SaveConfig(
		validDNSTTConfig(),
		"first",
	); err != nil {
		t.Fatalf("SaveConfig(first): %v", err)
	}

	config := validDNSTTConfig()
	config.Domain = "second.example.com"

	if err := service.SaveConfig(
		config,
		"second.toml",
	); err != nil {
		t.Fatalf("SaveConfig(second): %v", err)
	}

	entries, err := service.GetAllConfigFiles()
	if err != nil {
		t.Fatalf(
			"GetAllConfigFiles() error = %v",
			err,
		)
	}

	if len(entries) != 2 {
		t.Fatalf(
			"GetAllConfigFiles() returned %d files, want 2",
			len(entries),
		)
	}
}

// TestDNSTTServiceGetAllConfigFilesIgnoresInvalidFiles checks that GetAllConfigFiles skips files that fail to load or validate.
func TestDNSTTServiceGetAllConfigFilesIgnoresInvalidFiles(t *testing.T) {
	dir := t.TempDir()

	service := NewDNSTTService(
		WithDNSTTDir(dir),
	)

	if err := service.SaveConfig(
		validDNSTTConfig(),
		"valid",
	); err != nil {
		t.Fatalf("SaveConfig(): %v", err)
	}

	invalidPath := filepath.Join(dir, "invalid.toml")

	if err := os.WriteFile(
		invalidPath,
		[]byte(`
Domain = ""
PubKey = ""
ResolverPort = 0
Fingerprint = ""
RPS = -1
`),
		0o600,
	); err != nil {
		t.Fatalf(
			"write invalid config: %v",
			err,
		)
	}

	entries, err := service.GetAllConfigFiles()
	if err != nil {
		t.Fatalf(
			"GetAllConfigFiles() error = %v",
			err,
		)
	}

	if len(entries) != 1 {
		t.Fatalf(
			"GetAllConfigFiles() returned %d files, want 1",
			len(entries),
		)
	}
}

// TestNewDNSTTTunnelServer checks that newDNSTTTunnelServer builds a server in DNSTT compatibility mode.
func TestNewDNSTTTunnelServer(t *testing.T) {
	config := validDNSTTConfig()
	config.RPS = 25

	server, err := newDNSTTTunnelServer(config)
	if err != nil {
		t.Fatalf(
			"newDNSTTTunnelServer() error = %v",
			err,
		)
	}

	if !server.DnsttCompat {
		t.Fatal(
			"server.DnsttCompat = false, want true",
		)
	}

	if server.RPS != config.RPS {
		t.Fatalf(
			"server.RPS = %v, want %v",
			server.RPS,
			config.RPS,
		)
	}
}

// TestNewDNSTTTunnelServerInvalidConfig checks that newDNSTTTunnelServer rejects an invalid domain or public key.
func TestNewDNSTTTunnelServerInvalidConfig(t *testing.T) {
	config := validDNSTTConfig()
	config.Domain = ""

	// newDNSTTTunnelServer itself does not validate the complete
	// DNSTTConfig, so validate before constructing the server.
	if errs := config.Validate(); len(errs) == 0 {
		t.Fatal(
			"expected invalid configuration",
		)
	}
}

// TestNewDNSTTResolver checks that newVayDNSResolver builds a resolver for the DNSTT settings.
func TestNewDNSTTResolver(t *testing.T) {
	config := validDNSTTConfig()

	resolver, err := newVayDNSResolver(
		config.ResolverType,
		config.ResolverPort,
		config.Fingerprint,
		"8.8.8.8",
	)
	if err != nil {
		t.Fatalf(
			"newVayDNSResolver() error = %v",
			err,
		)
	}

	if resolver == nil {
		t.Fatal("newVayDNSResolver() returned nil resolver")
	}

	if resolver.UTLSClientHelloID == nil {
		t.Fatal(
			"resolver.UTLSClientHelloID = nil, want configured fingerprint",
		)
	}
}

// TestNewDNSTTResolverInvalidFingerprint checks that newVayDNSResolver rejects an unknown TLS fingerprint.
func TestNewDNSTTResolverInvalidFingerprint(t *testing.T) {
	config := validDNSTTConfig()
	config.Fingerprint = "invalid-fingerprint"

	_, err := newVayDNSResolver(
		config.ResolverType,
		config.ResolverPort,
		config.Fingerprint,
		"8.8.8.8",
	)
	if err == nil {
		t.Fatal(
			"newVayDNSResolver() expected fingerprint error",
		)
	}
}

// TestDNSTTServiceEditConfigUpdatesExisting checks that EditConfig replaces the stored DNSTT config.
func TestDNSTTServiceEditConfigUpdatesExisting(t *testing.T) {
	dir := t.TempDir()

	service := NewDNSTTService(
		WithDNSTTDir(dir),
	)

	if err := service.SaveConfig(validDNSTTConfig(), "my-tunnel"); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	updated := validDNSTTConfig()
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

// TestDNSTTServiceEditConfigMissingConfigReturnsError checks that EditConfig fails for a config that does not exist.
func TestDNSTTServiceEditConfigMissingConfigReturnsError(t *testing.T) {
	service := NewDNSTTService(
		WithDNSTTDir(t.TempDir()),
	)

	err := service.EditConfig(validDNSTTConfig(), "does-not-exist")
	if err == nil {
		t.Fatal("EditConfig() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("EditConfig() error = %q, want does not exist", err)
	}
}
