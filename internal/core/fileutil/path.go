package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetCurrentPath returns the absolute path of the current working directory.
func GetCurrentPath() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current path: %w", err)
	}
	return path, nil
}

// CheckFileExists checks if a regular file exists at the given path.
func CheckFileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// StripExt removes the extension from a file name or path.
func StripExt(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}

// HasExt checks case-insensitively if a path has the specified extension (e.g., ".json").
func HasExt(name, ext string) bool {
	return strings.EqualFold(filepath.Ext(name), ext)
}

// EnsureDir creates the parent directory for path when needed.
func EnsureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("ensure directory tree %q failed: %w", dir, err)
	}
	return nil
}

// GetOrCreateBaseDir returns an absolute directory path, creating it when absent.
func GetOrCreateBaseDir(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if err == nil {
		if !info.IsDir() {
			return "", os.ErrInvalid
		}
		return absPath, nil
	}

	if !os.IsNotExist(err) {
		return "", err
	}

	if err := os.MkdirAll(absPath, 0o755); err != nil {
		return "", err
	}

	return absPath, nil
}
