package dns

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	vaydns "github.com/net2share/vaydns/client"
	utls "github.com/refraction-networking/utls"
)

func validVayDNSConfig() VayDNSConfig {
	return VayDNSConfig{
		Domain:       "example.com",
		PubKey:       "cd6d78e954f48f62cb74cdcf8a2459d3d39786a7e11fc4f74c04bca86371f748",
		ClientIDSize: 2,
		MaxQnameLen:  253,
		MaxNumLabels: 0,
		MTU:          48,
		RPS:          0,
		RecordType:   TypeTXT,
		ResolverType: ResolverType(vaydns.ResolverTypeUDP),
		ResolverPort: 53,
		Fingerprint:  "chrome",
	}
}

// TestNormalizeConfigName checks that normalizeConfigName strips a .toml extension in any case and keeps other names as they are.
func TestNormalizeConfigName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "default",
			want: "default",
		},
		{
			name: "default.toml",
			want: "default",
		},
		{
			name: "DEFAULT.TOML",
			want: "DEFAULT",
		},
		{
			name: "config.TomL",
			want: "config",
		},
		{
			name: "my-config.toml",
			want: "my-config",
		},
		{
			name: "my.config.toml",
			want: "my.config",
		},
		{
			name: ".toml",
			want: "",
		},
		{
			name: "",
			want: "",
		},
		{
			name: "config.json",
			want: "config.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeConfigName(tt.name); got != tt.want {
				t.Fatalf(
					"normalizeConfigName(%q) = %q, want %q",
					tt.name,
					got,
					tt.want,
				)
			}
		})
	}
}

// TestConfigPath checks the paths configPath builds and the names it rejects.
func TestConfigPath(t *testing.T) {
	service := &vayDNSService{
		configs: newConfigStore[VayDNSConfig]("/tmp/vaydns", "VayDNS"),
	}

	tests := []struct {
		name string
		want string
	}{
		{
			name: "default",
			want: filepath.Join("/tmp/vaydns", "default.toml"),
		},
		{
			name: "default.toml",
			want: filepath.Join("/tmp/vaydns", "default.toml"),
		},
		{
			name: "DEFAULT.TOML",
			want: filepath.Join("/tmp/vaydns", "DEFAULT.toml"),
		},
		{
			name: "my-config",
			want: filepath.Join("/tmp/vaydns", "my-config.toml"),
		},
		{
			name: "my-config.toml",
			want: filepath.Join("/tmp/vaydns", "my-config.toml"),
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

// TestWithVayDNSDir checks that WithVayDNSDir sets the config directory of the service.
func TestWithVayDNSDir(t *testing.T) {
	service := &vayDNSService{
		configs: newConfigStore[VayDNSConfig]("/original", "VayDNS"),
	}

	WithVayDNSDir("/custom")(service)

	if service.configs.dir != "/custom" {
		t.Fatalf(
			"WithVayDNSDir() set dir to %q, want %q",
			service.configs.dir,
			"/custom",
		)
	}
}

// TestWithVayDNSDirEmpty checks that WithVayDNSDir with an empty value keeps the default directory.
func TestWithVayDNSDirEmpty(t *testing.T) {
	service := &vayDNSService{
		configs: newConfigStore[VayDNSConfig]("/original", "VayDNS"),
	}

	WithVayDNSDir("")(service)

	if service.configs.dir != "/original" {
		t.Fatalf(
			"WithVayDNSDir(\"\") changed dir to %q, want %q",
			service.configs.dir,
			"/original",
		)
	}
}

// TestNewVayDNSService checks that NewVayDNSService returns a service with the default config directory.
func TestNewVayDNSService(t *testing.T) {
	service := NewVayDNSService(
		WithVayDNSDir("/custom/vaydns"),
	)

	if service == nil {
		t.Fatal("NewVayDNSService() returned nil")
	}

	concrete, ok := service.(*vayDNSService)
	if !ok {
		t.Fatalf(
			"NewVayDNSService() returned %T, want *vayDNSService",
			service,
		)
	}

	if concrete.configs.dir != "/custom/vaydns" {
		t.Fatalf(
			"service.dir = %q, want %q",
			concrete.configs.dir,
			"/custom/vaydns",
		)
	}
}

// TestSaveConfigAndLoadConfig checks that a VayDNS config saved through the service loads back unchanged.
func TestSaveConfigAndLoadConfig(t *testing.T) {
	dir := t.TempDir()

	service := NewVayDNSService(
		WithVayDNSDir(dir),
	)

	want := validVayDNSConfig()
	want.Domain = "example.com"
	want.PubKey = "cd6d78e954f48f62cb74cdcf8a2459d3d39786a7e11fc4f74c04bca86371f748"
	want.ClientIDSize = 2
	want.MaxQnameLen = 240
	want.MaxNumLabels = 3
	want.MTU = 1200
	want.RPS = 10.5
	want.RecordType = TypeTXT
	want.ResolverType = ResolverType(vaydns.ResolverTypeUDP)
	want.ResolverPort = 53
	want.Fingerprint = "Chrome"

	if err := service.SaveConfig(want, "test"); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	path := filepath.Join(dir, "test.toml")

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file was not created: %v", err)
	}

	got, err := service.LoadConfig("test")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf(
			"LoadConfig() = %#v, want %#v",
			got,
			want,
		)
	}
}

// TestSaveConfigWithExtension checks that SaveConfig accepts a name that already ends in .toml.
func TestSaveConfigWithExtension(t *testing.T) {
	dir := t.TempDir()

	service := NewVayDNSService(
		WithVayDNSDir(dir),
	)

	config := validVayDNSConfig()

	if err := service.SaveConfig(config, "test.toml"); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	expectedPath := filepath.Join(dir, "test.toml")

	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf(
			"expected config at %q: %v",
			expectedPath,
			err,
		)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("got %d files, want 1", len(files))
	}

	if files[0].Name() != "test.toml" {
		t.Fatalf(
			"created file = %q, want %q",
			files[0].Name(),
			"test.toml",
		)
	}
}

// TestLoadConfigNotFound checks that LoadConfig returns an error for a missing config.
func TestLoadConfigNotFound(t *testing.T) {
	service := NewVayDNSService(
		WithVayDNSDir(t.TempDir()),
	)

	_, err := service.LoadConfig("missing")

	if err == nil {
		t.Fatal("LoadConfig() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "load VayDNS config") {
		t.Fatalf(
			"LoadConfig() error = %q, want load context",
			err,
		)
	}
}

// TestLoadConfigInvalidTOML checks that LoadConfig returns an error for a file that is not valid TOML.
func TestLoadConfigInvalidTOML(t *testing.T) {
	dir := t.TempDir()

	service := NewVayDNSService(
		WithVayDNSDir(dir),
	)

	path := filepath.Join(dir, "broken.toml")

	content := []byte(`
Domain = [
this is not valid TOML
`)

	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := service.LoadConfig("broken")

	if err == nil {
		t.Fatal("LoadConfig() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "load VayDNS config") {
		t.Fatalf(
			"LoadConfig() error = %q, want load context",
			err,
		)
	}
}

// TestSaveConfigCreatesExpectedFile checks that SaveConfig creates the expected .toml file on disk.
func TestSaveConfigCreatesExpectedFile(t *testing.T) {
	dir := t.TempDir()

	service := NewVayDNSService(
		WithVayDNSDir(dir),
	)

	config := validVayDNSConfig()

	if err := service.SaveConfig(config, "my-config.toml"); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("got %d files, want 1", len(files))
	}

	if got := files[0].Name(); got != "my-config.toml" {
		t.Fatalf(
			"created file = %q, want %q",
			got,
			"my-config.toml",
		)
	}
}

// TestEditConfigUpdatesExisting checks that EditConfig replaces the stored VayDNS config.
func TestEditConfigUpdatesExisting(t *testing.T) {
	dir := t.TempDir()

	service := NewVayDNSService(
		WithVayDNSDir(dir),
	)

	config := validEditableVayDNSConfig()
	if err := service.SaveConfig(config, "my-tunnel"); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	updated := config
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

// TestEditConfigMissingConfigReturnsError checks that EditConfig fails for a config that does not exist.
func TestEditConfigMissingConfigReturnsError(t *testing.T) {
	service := NewVayDNSService(
		WithVayDNSDir(t.TempDir()),
	)

	err := service.EditConfig(validEditableVayDNSConfig(), "does-not-exist")
	if err == nil {
		t.Fatal("EditConfig() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("EditConfig() error = %q, want does not exist", err)
	}
}

func validEditableVayDNSConfig() VayDNSConfig {
	config := validVayDNSConfig()
	config.PubKey = "cd6d78e954f48f62cb74cdcf8a2459d3d39786a7e11fc4f74c04bca86371f748"
	return config
}

// TestGetAllConfigFiles checks that GetAllConfigFiles returns the saved configs.
func TestGetAllConfigFiles(t *testing.T) {
	dir := t.TempDir()

	service := NewVayDNSService(
		WithVayDNSDir(dir),
	)

	first := validVayDNSConfig()
	first.Domain = "first.example.com"

	second := validVayDNSConfig()
	second.Domain = "second.example.com"

	configs := map[string]VayDNSConfig{
		"first.toml":  first,
		"second.toml": second,
	}

	for name, config := range configs {
		if err := service.SaveConfig(config, name); err != nil {
			t.Fatalf(
				"SaveConfig(%q) error = %v",
				name,
				err,
			)
		}
	}

	if err := os.WriteFile(
		filepath.Join(dir, "readme.txt"),
		[]byte("not a config"),
		0o644,
	); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(dir, "broken.toml"),
		[]byte("this is invalid [["),
		0o644,
	); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	entries, err := service.GetAllConfigFiles()
	if err != nil {
		t.Fatalf("GetAllConfigFiles() error = %v", err)
	}

	var names []string

	for _, entry := range entries {
		names = append(names, entry.Name)
	}

	wantNames := []string{
		"first",
		"second",
	}

	if !reflect.DeepEqual(names, wantNames) {
		t.Fatalf(
			"GetAllConfigFiles() = %#v, want %#v",
			names,
			wantNames,
		)
	}
}

// TestGetAllConfigFilesUsesConfiguredDirectory checks that GetAllConfigFiles reads from the directory set by WithVayDNSDir.
func TestGetAllConfigFilesUsesConfiguredDirectory(t *testing.T) {
	dir := t.TempDir()

	service := NewVayDNSService(
		WithVayDNSDir(dir),
	)

	config := validVayDNSConfig()

	if err := service.SaveConfig(config, "custom.toml"); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	entries, err := service.GetAllConfigFiles()
	if err != nil {
		t.Fatalf("GetAllConfigFiles() error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf(
			"GetAllConfigFiles() returned %d entries, want 1",
			len(entries),
		)
	}

	if entries[0].Name != "custom" {
		t.Fatalf(
			"entry.Name = %q, want %q",
			entries[0].Name,
			"custom",
		)
	}
}

// TestGetAllConfigFilesEmptyDirectory checks that GetAllConfigFiles on an empty directory returns no files and no error.
func TestGetAllConfigFilesEmptyDirectory(t *testing.T) {
	service := NewVayDNSService(
		WithVayDNSDir(t.TempDir()),
	)

	entries, err := service.GetAllConfigFiles()
	if err != nil {
		t.Fatalf("GetAllConfigFiles() error = %v", err)
	}

	if len(entries) != 0 {
		t.Fatalf(
			"GetAllConfigFiles() returned %d entries, want 0",
			len(entries),
		)
	}
}

// TestGetAllConfigFilesMissingDirectory checks that GetAllConfigFiles creates a missing directory and returns no files.
func TestGetAllConfigFilesMissingDirectory(t *testing.T) {
	service := NewVayDNSService(
		WithVayDNSDir(filepath.Join(t.TempDir(), "missing")),
	)

	entries, err := service.GetAllConfigFiles()
	if err != nil {
		t.Fatalf("GetAllConfigFiles() error = %v", err)
	}

	if len(entries) != 0 {
		t.Fatalf("GetAllConfigFiles() returned %d entries, want 0", len(entries))
	}
}

// TestParseClientHelloID checks that parseClientHelloID resolves a known fingerprint label.
func TestParseClientHelloID(t *testing.T) {
	tests := []struct {
		name        string
		fingerprint string
		want        utls.ClientHelloID
		wantErr     bool
	}{
		{
			name:        "empty fingerprint",
			fingerprint: "",
			wantErr:     true,
		},
		{
			name:        "unknown fingerprint",
			fingerprint: "definitely-not-a-real-fingerprint",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseClientHelloID(tt.fingerprint)

			if tt.wantErr {
				if err == nil {
					t.Fatal("parseClientHelloID() error = nil, want error")
				}

				return
			}

			if err != nil {
				t.Fatalf(
					"parseClientHelloID() error = %v",
					err,
				)
			}

			if got != tt.want {
				t.Fatalf(
					"parseClientHelloID() = %#v, want %#v",
					got,
					tt.want,
				)
			}
		})
	}
}

// TestParseClientHelloIDKnownFingerprints checks that parseClientHelloID resolves every label the vaydns library lists.
func TestParseClientHelloIDKnownFingerprints(t *testing.T) {
	clients := vaydns.UTLSClientHelloIDMap()

	if len(clients) == 0 {
		t.Fatal("UTLSClientHelloIDMap() returned no fingerprints")
	}

	for _, client := range clients {
		if client.ID == nil {
			continue
		}

		t.Run(client.Label, func(t *testing.T) {
			got, err := parseClientHelloID(client.Label)
			if err != nil {
				t.Fatalf(
					"parseClientHelloID(%q) error = %v",
					client.Label,
					err,
				)
			}

			if got != *client.ID {
				t.Fatalf(
					"parseClientHelloID(%q) = %#v, want %#v",
					client.Label,
					got,
					*client.ID,
				)
			}
		})
	}
}

// TestParseClientHelloIDCaseInsensitive checks that parseClientHelloID ignores letter case and spaces.
func TestParseClientHelloIDCaseInsensitive(t *testing.T) {
	clients := vaydns.UTLSClientHelloIDMap()

	for _, client := range clients {
		if client.ID == nil || client.Label == "" {
			continue
		}

		if client.Label == strings.ToUpper(client.Label) {
			continue
		}

		t.Run(client.Label, func(t *testing.T) {
			got, err := parseClientHelloID(
				strings.ToUpper(client.Label),
			)
			if err != nil {
				t.Fatalf(
					"parseClientHelloID(%q) error = %v",
					client.Label,
					err,
				)
			}

			if got != *client.ID {
				t.Fatalf(
					"parseClientHelloID(%q) = %#v, want %#v",
					client.Label,
					got,
					*client.ID,
				)
			}
		})
	}
}

// TestParseClientHelloIDError checks that parseClientHelloID returns an error for an unknown label.
func TestParseClientHelloIDError(t *testing.T) {
	const fingerprint = "invalid-fingerprint"

	_, err := parseClientHelloID(fingerprint)
	if err == nil {
		t.Fatal("parseClientHelloID() error = nil, want error")
	}

	if !strings.Contains(err.Error(), fingerprint) {
		t.Fatalf(
			"error = %q, want fingerprint %q in error",
			err,
			fingerprint,
		)
	}
}

// TestRenameConfig checks that RenameConfig moves a config to the new name.
func TestRenameConfig(t *testing.T) {
	dir := t.TempDir()

	service := NewVayDNSService(
		WithVayDNSDir(dir),
	)

	if err := service.SaveConfig(validEditableVayDNSConfig(), "old-name"); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	if err := service.RenameConfig("old-name", "new-name"); err != nil {
		t.Fatalf("RenameConfig() error = %v", err)
	}

	if _, err := service.LoadConfig("new-name"); err != nil {
		t.Fatalf("LoadConfig(new-name) error = %v", err)
	}

	if _, err := service.LoadConfig("old-name"); err == nil {
		t.Fatal("LoadConfig(old-name) expected error after rename")
	}
}

// TestEditConfigThenRename checks that an edited config keeps its new values after a rename.
func TestEditConfigThenRename(t *testing.T) {
	dir := t.TempDir()

	service := NewVayDNSService(
		WithVayDNSDir(dir),
	)

	if err := service.SaveConfig(validEditableVayDNSConfig(), "my-tunnel"); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	updated := validEditableVayDNSConfig()
	updated.Domain = "updated.example.com"

	if err := service.EditConfig(updated, "my-tunnel"); err != nil {
		t.Fatalf("EditConfig() error = %v", err)
	}

	if err := service.RenameConfig("my-tunnel", "renamed-tunnel"); err != nil {
		t.Fatalf("RenameConfig() error = %v", err)
	}

	got, err := service.LoadConfig("renamed-tunnel")
	if err != nil {
		t.Fatalf("LoadConfig(renamed-tunnel) error = %v", err)
	}

	if got.Domain != "updated.example.com" {
		t.Fatalf("Domain = %q, want %q", got.Domain, "updated.example.com")
	}
}

// TestRenameConfigDestinationAlreadyExists checks that RenameConfig fails when the new name is already in use.
func TestRenameConfigDestinationAlreadyExists(t *testing.T) {
	dir := t.TempDir()

	service := NewVayDNSService(
		WithVayDNSDir(dir),
	)

	config := validEditableVayDNSConfig()

	if err := service.SaveConfig(config, "a"); err != nil {
		t.Fatalf("SaveConfig(a) error = %v", err)
	}

	if err := service.SaveConfig(config, "b"); err != nil {
		t.Fatalf("SaveConfig(b) error = %v", err)
	}

	err := service.RenameConfig("a", "b")
	if err == nil {
		t.Fatal("RenameConfig() expected error for existing destination")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("RenameConfig() error = %q, want already exists", err)
	}
}

// TestValidatePrivateKey checks validatePrivateKey on valid, malformed, and empty keys.
func TestValidatePrivateKey(t *testing.T) {
	if err := validatePrivateKey(""); err == nil {
		t.Fatal("validatePrivateKey() error = nil, want error for empty key")
	}

	if err := validatePrivateKey("not a valid key"); err == nil {
		t.Fatal("validatePrivateKey() error = nil, want error for invalid key")
	}
}

// TestVayDNSClientIDSizeRange checks the client ID size bounds enforced by VayDNSConfig.Validate.
func TestVayDNSClientIDSizeRange(t *testing.T) {
	config := validVayDNSConfig()
	config.PubKey = "cd6d78e954f48f62cb74cdcf8a2459d3d39786a7e11fc4f74c04bca86371f748"

	config.ClientIDSize = 0
	errs := config.Validate()
	if _, ok := errs[FieldClientIDSize]; !ok {
		t.Fatal("client id size 0 should be invalid (range 1-8)")
	}

	config.ClientIDSize = 1
	if _, ok := config.Validate()["client_id_size"]; ok {
		t.Fatal("client id size 1 should be valid")
	}

	config.ClientIDSize = 8
	if _, ok := config.Validate()["client_id_size"]; ok {
		t.Fatal("client id size 8 should be valid")
	}

	config.ClientIDSize = 9
	if _, ok := config.Validate()["client_id_size"]; !ok {
		t.Fatal("client id size 9 should be invalid (range 1-8)")
	}
}
