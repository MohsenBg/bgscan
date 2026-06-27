package fileutil

import (
	"encoding/json"
	"fmt"
	"os"

	"bgscan/internal/logger"
)

// ═══════════════════════════════════════════════════════════
// Json Files Operations
// ═══════════════════════════════════════════════════════════

// WriteJSONFile writes data to a JSON file.
// The directory will be created if it does not exist.
// Existing files are overwritten.
func WriteJSONFile(path string, data any) error {
	if err := EnsureDir(path); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create json file %s: %w", path, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger.CoreError("error closing file: %v", err)
		}
	}()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	return nil
}

// WriteJSONFileIfNotExist writes JSON only if the file does not already exist.
func WriteJSONFileIfNotExist(path string, data any) error {
	if CheckFileExists(path) {
		return nil
	}
	return WriteJSONFile(path, data)
}

// GetJSONFile reads a JSON file and unmarshals it into dest.
func GetJSONFile(path string, dest any) error {
	if !CheckFileExists(path) {
		return fmt.Errorf("json file does not exist: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read json file %s: %w", path, err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("unmarshal json %s: %w", path, err)
	}

	return nil
}

// GetJSONFileOrDefault loads JSON from file or falls back to defaultValue.
// If loading fails, the default is copied into dest and written to disk.
func GetJSONFileOrDefault(path string, dest any, defaultValue any) error {
	if err := GetJSONFile(path, dest); err == nil {
		return nil
	}

	// copy default into destination
	tmp, err := json.Marshal(defaultValue)
	if err != nil {
		return fmt.Errorf("marshal default json: %w", err)
	}

	if err := json.Unmarshal(tmp, dest); err != nil {
		return fmt.Errorf("apply default json: %w", err)
	}

	// write default file for next run
	_ = WriteJSONFile(path, defaultValue)

	return nil
}

// UpdateJSONFile loads a JSON file, applies an update function,
// then writes the result back to disk.
func UpdateJSONFile(path string, dest any, updateFn func(any) error) error {
	if err := GetJSONFile(path, dest); err != nil {
		return err
	}

	if err := updateFn(dest); err != nil {
		return fmt.Errorf("update json data: %w", err)
	}

	return WriteJSONFile(path, dest)
}
