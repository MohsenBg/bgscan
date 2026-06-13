package fileutil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CreateTmpFile builds a temporary file using a naming pattern.
// Caller is responsible for closing and removing the returned *os.File.
func CreateTmpFile(pattern string) (*os.File, string, error) {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return nil, "", fmt.Errorf("create temp file: %w", err)
	}

	absPath, err := filepath.Abs(f.Name())
	if err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return nil, "", fmt.Errorf("resolve temp file absolute path: %w", err)
	}

	return f, absPath, nil
}

// CreateTmpJSONFile marshals data into an un-indented temporary file and returns its path.
func CreateTmpJSONFile(pattern string, data any) (string, error) {
	f, path, err := CreateTmpFile(pattern)
	if err != nil {
		return "", err
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	if err := enc.Encode(data); err != nil {
		_ = f.Close() // CRITICAL: Close descriptor before trying to drop from file index allocation table
		_ = os.Remove(path)
		return "", fmt.Errorf("encode json to temp file: %w", err)
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("close temp file: %w", err)
	}

	return path, nil
}
