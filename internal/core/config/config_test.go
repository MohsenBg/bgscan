package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func newTestStore(t *testing.T) Store {
	t.Helper()

	return NewStore(WithSettingsDir(t.TempDir()))
}

func defaultScannerConfig() ScannerConfig {
	return ScannerConfig{
		General: DefaultGeneralConfig(),
		Writer:  DefaultWriterConfig(),
		ICMP:    DefaultICMPConfig(),
		TCP:     DefaultTCPConfig(),
		HTTP:    DefaultHTTPConfig(),
		Xray:    DefaultXrayConfig(),
		DNS:     DefaultDNSConfig(),
	}
}

// TestNewStore_UsesBasePath verifies that under `go test` (BasePath falls
// back to the working directory) the default store lives in ./settings.
func TestNewStore_UsesDefaultDirectory(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	want := filepath.Join(wd, settingsDir)

	store := NewStore()
	if store.dir != want {
		t.Fatalf("expected directory %q, got %q", want, store.dir)
	}
}

func TestNewStore_UsesConfiguredDirectory(t *testing.T) {
	dir := t.TempDir()

	store := NewStore(WithSettingsDir(dir))

	if store.dir != dir {
		t.Fatalf("expected directory %q, got %q", dir, store.dir)
	}
}

func TestStoreLoad_CreatesDefaultFiles(t *testing.T) {
	store := newTestStore(t)

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	want := defaultScannerConfig()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected config:\nwant: %+v\ngot:  %+v", want, got)
	}

	files := []string{
		generalFile,
		writerFile,
		icmpFile,
		tcpFile,
		httpFile,
		xrayFile,
		dnsFile,
	}

	for _, filename := range files {
		path := store.path(filename)

		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected settings file %q: %v", path, err)
			continue
		}

		if info.IsDir() {
			t.Errorf("expected file at %q, found directory", path)
		}
	}
}

func TestStoreLoad_ReturnsErrorForInvalidFile(t *testing.T) {
	store := newTestStore(t)
	path := store.path(generalFile)

	if err := os.WriteFile(path, []byte("invalid = ["), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error for invalid TOML")
	}
}

func TestStoreSave_WritesAllConfigurations(t *testing.T) {
	store := newTestStore(t)
	want := defaultScannerConfig()

	if err := store.SaveGeneral(want.General); err != nil {
		t.Fatalf("SaveGeneral() error: %v", err)
	}
	if err := store.SaveWriter(want.Writer); err != nil {
		t.Fatalf("SaveWriter() error: %v", err)
	}
	if err := store.SaveICMP(want.ICMP); err != nil {
		t.Fatalf("SaveICMP() error: %v", err)
	}
	if err := store.SaveTCP(want.TCP); err != nil {
		t.Fatalf("SaveTCP() error: %v", err)
	}
	if err := store.SaveHTTP(want.HTTP); err != nil {
		t.Fatalf("SaveHTTP() error: %v", err)
	}
	if err := store.SaveXray(want.Xray); err != nil {
		t.Fatalf("SaveXray() error: %v", err)
	}
	if err := store.SaveDNS(want.DNS); err != nil {
		t.Fatalf("SaveDNS() error: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("saved config was not loaded correctly:\nwant: %+v\ngot:  %+v", want, got)
	}
}

func TestStoreSave_ReturnsErrorWhenDirectoryIsAFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "not-a-directory")

	if err := os.WriteFile(path, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}

	store := NewStore(WithSettingsDir(path))

	if err := store.SaveTCP(DefaultTCPConfig()); err == nil {
		t.Fatal("expected save error")
	}
}
