package fileutil

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// ═══════════════════════════════════════════════════════════
// TOML File Operations
// ═══════════════════════════════════════════════════════════

// WriteTOMLFile writes data to a TOML file at the specified path.
// The directory will be created if it does not exist.
func WriteTOMLFile(path string, data any) error {
	if err := EnsureDir(path); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create TOML file %s: %w", path, err)
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("encode TOML to %s: %w", path, err)
	}

	return nil
}

// WriteTOMLFileIfNotExist writes TOML only if the file does not already exist.
func WriteTOMLFileIfNotExist(path string, data any) error {
	if CheckFileExists(path) {
		return nil
	}
	return WriteTOMLFile(path, data)
}

// GetTOMLFile reads a TOML file and unmarshals it into dest.
func GetTOMLFile(path string, dest any) error {
	if !CheckFileExists(path) {
		return fmt.Errorf("TOML file does not exist: %s", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open TOML file %s: %w", path, err)
	}
	defer f.Close()

	dec := toml.NewDecoder(f)
	if err := dec.Decode(dest); err != nil {
		return fmt.Errorf("decode TOML from %s: %w", path, err)
	}

	return nil
}

// GetTOMLFileOrDefault loads TOML from file and falls back to defaultValue on error.
// If fallback is used, the default value is persisted to disk.
func GetTOMLFileOrDefault(path string, dest any, defaultValue any) error {
	if err := GetTOMLFile(path, dest); err == nil {
		return nil
	}

	// populate dest from defaultValue
	b, err := toml.Marshal(defaultValue)
	if err != nil {
		return fmt.Errorf("marshal default TOML value: %w", err)
	}

	if err := toml.Unmarshal(b, dest); err != nil {
		return fmt.Errorf("unmarshal default TOML value: %w", err)
	}

	// persist default for next run (best effort)
	_ = WriteTOMLFile(path, defaultValue)

	return nil
}

// UpdateTOMLFile reads a TOML file, applies updateFn, and writes it back.
func UpdateTOMLFile(
	path string,
	dest any,
	updateFn func(any) error,
) error {

	if err := GetTOMLFile(path, dest); err != nil {
		return fmt.Errorf("load TOML file %s: %w", path, err)
	}

	if err := updateFn(dest); err != nil {
		return fmt.Errorf("update TOML value: %w", err)
	}

	if err := WriteTOMLFile(path, dest); err != nil {
		return fmt.Errorf("write updated TOML file %s: %w", path, err)
	}

	return nil
}
