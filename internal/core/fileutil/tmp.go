package fileutil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CreateTmpFile creates a temporary file using pattern.
// The caller must close and remove the returned file.
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

// CreateTmpJSONFile writes data as JSON to a temporary file and returns its path.
func CreateTmpJSONFile(pattern string, data any) (string, error) {
	f, path, err := CreateTmpFile(pattern)
	if err != nil {
		return "", err
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	if err := enc.Encode(data); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("encode json to temp file: %w", err)
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("close temp file: %w", err)
	}

	return path, nil
}
