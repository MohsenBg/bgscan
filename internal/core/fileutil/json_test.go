package fileutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestWriteJSONFile_WritesAndCreatesDirectory covers write jsonfile writes and creates directory.
func TestWriteJSONFile_WritesAndCreatesDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")

	input := map[string]any{
		"name":    "bgscan",
		"version": 2,
	}

	if err := WriteJSONFile(path, input); err != nil {
		t.Fatalf("WriteJSONFile() error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected JSON content")
	}
}

// TestWriteJSONFile_OverwritesExistingFile covers write jsonfile overwrites existing file.
func TestWriteJSONFile_OverwritesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	if err := WriteJSONFile(path, map[string]string{"value": "old"}); err != nil {
		t.Fatal(err)
	}

	if err := WriteJSONFile(path, map[string]string{"value": "new"}); err != nil {
		t.Fatal(err)
	}

	result, err := ReadJSONFile[map[string]string](path)
	if err != nil {
		t.Fatal(err)
	}

	if result["value"] != "new" {
		t.Fatalf("expected %q, got %q", "new", result["value"])
	}
}

// TestReadJSONFile_ReadsJSON covers read jsonfile reads json.
func TestReadJSONFile_ReadsJSON(t *testing.T) {
	type data struct {
		Name string `json:"name"`
	}

	path := filepath.Join(t.TempDir(), "data.json")
	want := data{Name: "test"}

	if err := WriteJSONFile(path, want); err != nil {
		t.Fatal(err)
	}

	got, err := ReadJSONFile[data](path)
	if err != nil {
		t.Fatalf("ReadJSONFile() error: %v", err)
	}

	if got != want {
		t.Fatalf("unexpected result: want %+v, got %+v", want, got)
	}
}

// TestReadJSONFile_ReturnsErrorWhenMissing covers read jsonfile returns error when missing.
func TestReadJSONFile_ReturnsErrorWhenMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")

	_, err := ReadJSONFile[map[string]any](path)
	if err == nil {
		t.Fatal("expected error")
	}
}

// TestReadJSONFile_ReturnsErrorForInvalidJSON covers read jsonfile returns error for invalid json.
func TestReadJSONFile_ReturnsErrorForInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.json")

	if err := os.WriteFile(path, []byte("{invalid json"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadJSONFile[map[string]any](path)
	if err == nil {
		t.Fatal("expected JSON decode error")
	}
}

// TestWriteJSONFileIfNotExist_CreatesMissingFile covers write jsonfile if not exist creates missing file.
func TestWriteJSONFileIfNotExist_CreatesMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	if err := WriteJSONFileIfNotExist(path, map[string]string{"value": "created"}); err != nil {
		t.Fatal(err)
	}

	result, err := ReadJSONFile[map[string]string](path)
	if err != nil {
		t.Fatal(err)
	}

	if result["value"] != "created" {
		t.Fatalf("expected %q, got %q", "created", result["value"])
	}
}

// TestWriteJSONFileIfNotExist_DoesNotOverwrite covers write jsonfile if not exist does not overwrite.
func TestWriteJSONFileIfNotExist_DoesNotOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	if err := WriteJSONFile(path, map[string]string{"value": "original"}); err != nil {
		t.Fatal(err)
	}

	if err := WriteJSONFileIfNotExist(path, map[string]string{"value": "changed"}); err != nil {
		t.Fatal(err)
	}

	result, err := ReadJSONFile[map[string]string](path)
	if err != nil {
		t.Fatal(err)
	}

	if result["value"] != "original" {
		t.Fatalf("expected original value, got %q", result["value"])
	}
}

// TestReadJSONFileOrDefault_UsesExistingValue covers read jsonfile or default uses existing value.
func TestReadJSONFileOrDefault_UsesExistingValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	if err := WriteJSONFile(path, map[string]string{"value": "existing"}); err != nil {
		t.Fatal(err)
	}

	got := ReadJSONFileOrDefault(path, map[string]string{"value": "default"})

	if got["value"] != "existing" {
		t.Fatalf("expected %q, got %q", "existing", got["value"])
	}
}

// TestReadJSONFileOrDefault_ReturnsDefaultForMissingFile covers read jsonfile or default returns default for missing file.
func TestReadJSONFileOrDefault_ReturnsDefaultForMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	want := map[string]string{"value": "default"}

	got := ReadJSONFileOrDefault(path, want)

	if got["value"] != "default" {
		t.Fatalf("expected default value, got %q", got["value"])
	}

	if FileExists(path) {
		t.Fatal("default value should not be written to disk")
	}
}

// TestUpdateJSONFile_UpdatesData covers update jsonfile updates data.
func TestUpdateJSONFile_UpdatesData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	if err := WriteJSONFile(path, map[string]int{"count": 1}); err != nil {
		t.Fatal(err)
	}

	err := UpdateJSONFile(path, func(data map[string]int) (map[string]int, error) {
		data["count"]++
		return data, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := ReadJSONFile[map[string]int](path)
	if err != nil {
		t.Fatal(err)
	}

	if result["count"] != 2 {
		t.Fatalf("expected count 2, got %d", result["count"])
	}
}

// TestUpdateJSONFile_ReturnsUpdateError covers update jsonfile returns update error.
func TestUpdateJSONFile_ReturnsUpdateError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	if err := WriteJSONFile(path, map[string]string{"x": "y"}); err != nil {
		t.Fatal(err)
	}

	expected := errors.New("update failed")

	err := UpdateJSONFile(path, func(map[string]string) (map[string]string, error) {
		return nil, expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected error %v, got %v", expected, err)
	}
}
