package fileutil

import (
	"errors"
	"path/filepath"
	"testing"
)

type testConfig struct {
	Name    string `toml:"name"`
	Enabled bool   `toml:"enabled"`
	Count   int    `toml:"count"`
}

func TestWriteTOMLFile_WritesAndCreatesDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.toml")

	want := testConfig{
		Name:    "bgscan",
		Enabled: true,
		Count:   10,
	}

	if err := WriteTOMLFile(path, want); err != nil {
		t.Fatalf("WriteTOMLFile() error: %v", err)
	}

	got, err := ReadTOMLFile[testConfig](path)
	if err != nil {
		t.Fatalf("ReadTOMLFile() error: %v", err)
	}

	if got != want {
		t.Fatalf("unexpected config: want %+v, got %+v", want, got)
	}
}

func TestWriteTOMLFile_OverwritesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")

	if err := WriteTOMLFile(path, testConfig{Name: "old"}); err != nil {
		t.Fatal(err)
	}

	if err := WriteTOMLFile(path, testConfig{Name: "new"}); err != nil {
		t.Fatal(err)
	}

	got, err := ReadTOMLFile[testConfig](path)
	if err != nil {
		t.Fatal(err)
	}

	if got.Name != "new" {
		t.Fatalf("expected name %q, got %q", "new", got.Name)
	}
}

func TestWriteTOMLFileIfNotExist_CreatesMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")

	if err := WriteTOMLFileIfNotExist(path, testConfig{Name: "created"}); err != nil {
		t.Fatal(err)
	}

	got, err := ReadTOMLFile[testConfig](path)
	if err != nil {
		t.Fatal(err)
	}

	if got.Name != "created" {
		t.Fatalf("expected name %q, got %q", "created", got.Name)
	}
}

func TestWriteTOMLFileIfNotExist_DoesNotOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")

	if err := WriteTOMLFile(path, testConfig{Name: "original"}); err != nil {
		t.Fatal(err)
	}

	if err := WriteTOMLFileIfNotExist(path, testConfig{Name: "changed"}); err != nil {
		t.Fatal(err)
	}

	got, err := ReadTOMLFile[testConfig](path)
	if err != nil {
		t.Fatal(err)
	}

	if got.Name != "original" {
		t.Fatalf("expected original value, got %q", got.Name)
	}
}

func TestReadTOMLFile_ReturnsErrorForMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.toml")

	_, err := ReadTOMLFile[testConfig](path)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestReadTOMLFile_ReturnsErrorForInvalidTOML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.toml")

	if err := WriteTextFile(path, "invalid = ["); err != nil {
		t.Fatal(err)
	}

	_, err := ReadTOMLFile[testConfig](path)
	if err == nil {
		t.Fatal("expected TOML decode error")
	}
}

func TestReadTOMLFileOrDefault_UsesExistingValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")

	if err := WriteTOMLFile(path, testConfig{Name: "existing"}); err != nil {
		t.Fatal(err)
	}

	got := ReadTOMLFileOrDefault(path, testConfig{Name: "default"})

	if got.Name != "existing" {
		t.Fatalf("expected existing value, got %q", got.Name)
	}
}

func TestReadTOMLFileOrDefault_ReturnsDefaultForMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")

	want := testConfig{
		Name:  "default",
		Count: 5,
	}

	got := ReadTOMLFileOrDefault(path, want)

	if got != want {
		t.Fatalf("expected default %+v, got %+v", want, got)
	}
}

func TestUpdateTOMLFile_UpdatesValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")

	if err := WriteTOMLFile(path, testConfig{Count: 1}); err != nil {
		t.Fatal(err)
	}

	err := UpdateTOMLFile(path, func(cfg testConfig) (testConfig, error) {
		cfg.Count++
		return cfg, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := ReadTOMLFile[testConfig](path)
	if err != nil {
		t.Fatal(err)
	}

	if got.Count != 2 {
		t.Fatalf("expected count 2, got %d", got.Count)
	}
}

func TestUpdateTOMLFile_ReturnsUpdateError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")

	if err := WriteTOMLFile(path, testConfig{}); err != nil {
		t.Fatal(err)
	}

	expected := errors.New("update failed")

	err := UpdateTOMLFile(path, func(testConfig) (testConfig, error) {
		return testConfig{}, expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected update error %v, got %v", expected, err)
	}
}
