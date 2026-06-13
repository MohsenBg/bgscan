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

// EnsureDir checks and creates the target path's parent directory structures.
func EnsureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("ensure directory tree %q failed: %w", dir, err)
	}
	return nil
}

func GetOrCreateBaseDir(path string) (string, error) {
	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	// Check if exists
	info, err := os.Stat(absPath)
	if err == nil {
		// Exists — make sure it's a directory
		if !info.IsDir() {
			return "", os.ErrInvalid
		}
		return absPath, nil
	}

	// If error is not "not exists", something else is wrong
	if !os.IsNotExist(err) {
		return "", err
	}

	// Create directory (including parents)
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return "", err
	}

	return absPath, nil
}

