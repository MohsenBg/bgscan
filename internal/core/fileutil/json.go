package fileutil

import (
	"encoding/json"
	"fmt"
	"os"
)

// WriteJSONFile writes value as indented JSON, creating the parent directory
// if it is missing.
func WriteJSONFile(path string, value any) error {
	if err := EnsureParentDir(path); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create JSON file %q: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("encode JSON file %q: %w", path, err)
	}

	return nil
}

// WriteJSONFileIfNotExist creates path only when it is missing. An existing
// file is left alone, even if it holds something else.
func WriteJSONFileIfNotExist(path string, value any) error {
	if FileExists(path) {
		return nil
	}

	return WriteJSONFile(path, value)
}

// ReadJSONFile decodes path into a fresh T.
func ReadJSONFile[T any](path string) (T, error) {
	var value T

	data, err := os.ReadFile(path)
	if err != nil {
		return value, fmt.Errorf("read JSON file %q: %w", path, err)
	}

	if err := json.Unmarshal(data, &value); err != nil {
		return value, fmt.Errorf("decode JSON file %q: %w", path, err)
	}

	return value, nil
}

// ReadJSONFileOrDefault returns the decoded value, or defaultValue when the
// file is missing or broken. The default is not written back to disk.
func ReadJSONFileOrDefault[T any](path string, defaultValue T) T {
	value, err := ReadJSONFile[T](path)
	if err != nil {
		return defaultValue
	}

	return value
}

// UpdateJSONFile reads the file, hands the value to update, and writes the
// result back. A failed update leaves the file untouched.
func UpdateJSONFile[T any](path string, update func(T) (T, error)) error {
	value, err := ReadJSONFile[T](path)
	if err != nil {
		return fmt.Errorf("read JSON file %q: %w", path, err)
	}

	value, err = update(value)
	if err != nil {
		return fmt.Errorf("update JSON file %q: %w", path, err)
	}

	if err := WriteJSONFile(path, value); err != nil {
		return fmt.Errorf("write JSON file %q: %w", path, err)
	}

	return nil
}
