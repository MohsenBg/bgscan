package fileutil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CreateTempIn creates a temp file in dir.
//
// The directory is the point. rename is only atomic inside one filesystem, and
// the system temp directory is usually a different one. A file that will be
// published has to start life next to its destination.
func CreateTempIn(dir, pattern string) (*os.File, string, error) {
	if dir == "" {
		return nil, "", fmt.Errorf("create temp file: directory is required")
	}

	if err := EnsureDir(dir); err != nil {
		return nil, "", err
	}

	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return nil, "", fmt.Errorf("create temp file in %q: %w", dir, err)
	}

	absPath, err := filepath.Abs(f.Name())
	if err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return nil, "", fmt.Errorf("resolve temp file absolute path: %w", err)
	}

	return f, absPath, nil
}

// CreateSystemTempFile creates a temp file in the system temp directory.
//
// Use it for scratch data only. Anything that will be renamed into place needs
// CreateTempIn instead. The caller closes and removes the file.
func CreateSystemTempFile(pattern string) (*os.File, string, error) {
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

// CreateSystemTempJSONFile writes data as indented JSON to a system temp file
// and returns its path.
func CreateSystemTempJSONFile(pattern string, data any) (string, error) {
	f, path, err := CreateSystemTempFile(pattern)
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
