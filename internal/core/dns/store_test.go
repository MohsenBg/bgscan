package dns

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeStoreConfig is a minimal tunnelConfig for exercising the generic
// config store without a real protocol implementation.
type fakeStoreConfig struct {
	Valid  bool
	Domain string
}

func (c fakeStoreConfig) Validate() map[string]error {
	if c.Valid {
		return nil
	}

	return map[string]error{"valid": errors.New("invalid config")}
}

func newFakeStore(t *testing.T) configStore[fakeStoreConfig] {
	t.Helper()
	return newConfigStore[fakeStoreConfig](t.TempDir(), "Fake")
}

func TestConfigStoreSaveLoadRoundTrip(t *testing.T) {
	store := newFakeStore(t)

	want := fakeStoreConfig{Valid: true, Domain: "example.com"}

	if err := store.SaveConfig(want, "test"); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	got, err := store.LoadConfig("test")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if got != want {
		t.Fatalf("LoadConfig() = %#v, want %#v", got, want)
	}
}

func TestConfigStoreSaveErrors(t *testing.T) {
	store := newFakeStore(t)
	valid := fakeStoreConfig{Valid: true, Domain: "example.com"}

	if err := store.SaveConfig(valid, ""); err == nil {
		t.Error("SaveConfig() with empty name should fail")
	}

	if err := store.SaveConfig(valid, "test"); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	err := store.SaveConfig(valid, "test")
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("SaveConfig() duplicate = %v, want already-exists error", err)
	}

	err = store.SaveConfig(fakeStoreConfig{Valid: false}, "bad")
	if err == nil || !strings.Contains(err.Error(), "invalid Fake config") {
		t.Errorf("SaveConfig() invalid config = %v, want validation error", err)
	}

	err = store.SaveConfig(valid, "bad.toml")
	if err != nil {
		t.Fatalf("SaveConfig() with extension error = %v", err)
	}

	if _, err := store.LoadConfig("bad"); err != nil {
		t.Errorf("LoadConfig() after saving with extension = %v, want same file", err)
	}
}

func TestConfigStoreEdit(t *testing.T) {
	store := newFakeStore(t)
	config := fakeStoreConfig{Valid: true, Domain: "first.com"}

	if err := store.SaveConfig(config, "test"); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	updated := config
	updated.Domain = "second.com"

	if err := store.EditConfig(updated, "test"); err != nil {
		t.Fatalf("EditConfig() error = %v", err)
	}

	got, err := store.LoadConfig("test")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if got.Domain != "second.com" {
		t.Errorf("Domain = %q, want second.com", got.Domain)
	}

	err = store.EditConfig(updated, "missing")
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("EditConfig() missing = %v, want does-not-exist error", err)
	}

	err = store.EditConfig(fakeStoreConfig{Valid: false}, "test")
	if err == nil || !strings.Contains(err.Error(), "invalid Fake config") {
		t.Errorf("EditConfig() invalid = %v, want validation error", err)
	}
}

func TestConfigStoreGetAllConfigFiles(t *testing.T) {
	dir := t.TempDir()
	store := newConfigStore[fakeStoreConfig](dir, "Fake")

	if err := store.SaveConfig(fakeStoreConfig{Valid: true, Domain: "a.com"}, "first"); err != nil {
		t.Fatalf("SaveConfig(first) error = %v", err)
	}

	if err := store.SaveConfig(fakeStoreConfig{Valid: true, Domain: "b.com"}, "second"); err != nil {
		t.Fatalf("SaveConfig(second) error = %v", err)
	}

	// An invalid config and a non-TOML file must both be skipped.
	if err := os.WriteFile(
		filepath.Join(dir, "invalid.toml"),
		[]byte("Valid = false\n"),
		0o600,
	); err != nil {
		t.Fatalf("write invalid config: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(dir, "notes.txt"),
		[]byte("not a config"),
		0o600,
	); err != nil {
		t.Fatalf("write non-config: %v", err)
	}

	files, err := store.GetAllConfigFiles()
	if err != nil {
		t.Fatalf("GetAllConfigFiles() error = %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("GetAllConfigFiles() = %d files, want 2", len(files))
	}

	names := map[string]bool{}
	for _, file := range files {
		names[file.Name] = true

		if file.Config.Domain == "" {
			t.Errorf("file %q has empty config", file.Name)
		}
		if file.Path == "" {
			t.Errorf("file %q has empty path", file.Name)
		}
		if file.CreatedAt.IsZero() {
			t.Errorf("file %q has zero CreatedAt", file.Name)
		}
	}

	for _, want := range []string{"first", "second"} {
		if !names[want] {
			t.Errorf("config %q missing from results", want)
		}
	}
}

func TestConfigStoreGetAllConfigFilesEmptyDir(t *testing.T) {
	store := newFakeStore(t)

	files, err := store.GetAllConfigFiles()
	if err != nil {
		t.Fatalf("GetAllConfigFiles() error = %v", err)
	}

	if len(files) != 0 {
		t.Fatalf("GetAllConfigFiles() = %d files, want 0", len(files))
	}
}

func TestConfigStoreValidateAllConfigs(t *testing.T) {
	dir := t.TempDir()
	store := newConfigStore[fakeStoreConfig](dir, "Fake")

	if err := store.SaveConfig(fakeStoreConfig{Valid: true}, "ok"); err != nil {
		t.Fatalf("SaveConfig(ok) error = %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(dir, "invalid.toml"),
		[]byte("Valid = false\n"),
		0o600,
	); err != nil {
		t.Fatalf("write invalid config: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(dir, "broken.toml"),
		[]byte("not [valid toml"),
		0o600,
	); err != nil {
		t.Fatalf("write broken config: %v", err)
	}

	results, err := store.ValidateAllConfigs()
	if err != nil {
		t.Fatalf("ValidateAllConfigs() error = %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("ValidateAllConfigs() = %d results, want 2", len(results))
	}

	byName := map[string]ConfigValidationResult{}
	for _, result := range results {
		name := normalizeConfigName(result.File.Name)
		byName[name] = result
	}

	invalid, ok := byName["invalid"]
	if !ok {
		t.Fatal("invalid.toml not reported")
	}
	if _, ok := invalid.Errors["valid"]; !ok {
		t.Errorf("invalid.toml errors = %v, want 'valid' key", invalid.Errors)
	}

	broken, ok := byName["broken"]
	if !ok {
		t.Fatal("broken.toml not reported")
	}
	if len(broken.Errors) != 0 {
		t.Errorf("broken.toml errors = %v, want empty map for unloadable file", broken.Errors)
	}

	if _, ok := byName["ok"]; ok {
		t.Error("valid config should not be reported")
	}
}

func TestConfigStoreRenameConfig(t *testing.T) {
	store := newFakeStore(t)
	config := fakeStoreConfig{Valid: true, Domain: "example.com"}

	if err := store.SaveConfig(config, "old"); err != nil {
		t.Fatalf("SaveConfig(old) error = %v", err)
	}

	if err := store.SaveConfig(config, "other"); err != nil {
		t.Fatalf("SaveConfig(other) error = %v", err)
	}

	if err := store.RenameConfig("old", "new"); err != nil {
		t.Fatalf("RenameConfig() error = %v", err)
	}

	if _, err := store.LoadConfig("new"); err != nil {
		t.Errorf("LoadConfig(new) error = %v, want renamed config", err)
	}

	if _, err := store.LoadConfig("old"); err == nil {
		t.Error("LoadConfig(old) expected error after rename")
	}

	err := store.RenameConfig("new", "other")
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("RenameConfig() to existing = %v, want already-exists error", err)
	}

	err = store.RenameConfig("missing", "elsewhere")
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("RenameConfig() missing = %v, want does-not-exist error", err)
	}
}

func TestTunnelConfigDir(t *testing.T) {
	got := tunnelConfigDir("someproto")

	wantSuffix := filepath.Join("assets", "dns-tunneling", "someproto")
	if !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("tunnelConfigDir() = %q, want suffix %q", got, wantSuffix)
	}
}
