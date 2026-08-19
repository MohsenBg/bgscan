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
		PubKey:       "test-public-key",
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

func TestConfigPath(t *testing.T) {
	service := &vayDNSService{
		dir: "/tmp/vaydns",
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
			if got := service.configPath(tt.name); got != tt.want {
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

func TestWithVayDNSDir(t *testing.T) {
	service := &vayDNSService{
		dir: "/original",
	}

	WithVayDNSDir("/custom")(service)

	if service.dir != "/custom" {
		t.Fatalf(
			"WithVayDNSDir() set dir to %q, want %q",
			service.dir,
			"/custom",
		)
	}
}

func TestWithVayDNSDirEmpty(t *testing.T) {
	service := &vayDNSService{
		dir: "/original",
	}

	WithVayDNSDir("")(service)

	if service.dir != "/original" {
		t.Fatalf(
			"WithVayDNSDir(\"\") changed dir to %q, want %q",
			service.dir,
			"/original",
		)
	}
}

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

	if concrete.dir != "/custom/vaydns" {
		t.Fatalf(
			"service.dir = %q, want %q",
			concrete.dir,
			"/custom/vaydns",
		)
	}
}

func TestSaveConfigAndLoadConfig(t *testing.T) {
	dir := t.TempDir()

	service := NewVayDNSService(
		WithVayDNSDir(dir),
	)

	want := validVayDNSConfig()
	want.Domain = "example.com"
	want.PubKey = "test-public-key"
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
