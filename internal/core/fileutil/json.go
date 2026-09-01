package fileutil

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/MohsenBg/bgscan/internal/logger"
)

// WriteJSONFile encodes value as JSON and writes it to path.
// It creates the parent directory when needed.
func WriteJSONFile(path string, value any) error {
	if err := EnsureFileDir(path); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create JSON file %q: %w", path, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.CoreError("close JSON file %q: %v", path, err)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("encode JSON file %q: %w", path, err)
	}

	return nil
}

// WriteJSONFileIfNotExist writes value only when path does not exist.
func WriteJSONFileIfNotExist(path string, value any) error {
	if CheckFileExists(path) {
		return nil
	}

	return WriteJSONFile(path, value)
}

// ReadJSONFile decodes the JSON file at path into a new value of type T.
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

// ReadJSONFileOrDefault returns the value in path, or defaultValue if it cannot
// be read or decoded. It does not write defaultValue to disk.
func ReadJSONFileOrDefault[T any](path string, defaultValue T) T {
	value, err := ReadJSONFile[T](path)
	if err != nil {
		return defaultValue
	}

	return value
}

// UpdateJSONFile reads a JSON value, passes it to update, then saves the
// returned value.
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
