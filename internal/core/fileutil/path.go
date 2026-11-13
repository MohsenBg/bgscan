package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetCurrentPath returns the working directory as an absolute path.
func GetCurrentPath() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current path: %w", err)
	}
	return path, nil
}

// FileExists reports whether path names a regular file. A directory returns
// false even when it exists.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// StripExt drops the last extension from a name or path.
func StripExt(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}

// HasExt compares the last extension, ignoring case.
func HasExt(name, ext string) bool {
	return strings.EqualFold(filepath.Ext(name), ext)
}

// EnsureParentDir creates the directory that holds path. A bare name like
// "file.txt" has no directory part, so there is nothing to do.
func EnsureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("ensure directory tree %q failed: %w", dir, err)
	}

	return nil
}

// EnsureDir creates path and any missing parents. An empty path is a no-op.
func EnsureDir(path string) error {
	if path == "" {
		return nil
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("create directory %q: %w", path, err)
	}

	return nil
}

// RemoveFile removes path. A file that is already gone is not an error, since
// callers usually want "make sure this is gone" rather than "was it there".
func RemoveFile(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove file %q: %w", path, err)
	}
	return nil
}

// GetOrCreateBaseDir resolves path to an absolute directory and creates it
// when missing. An existing non-directory is an error.
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

// BasePath returns the directory holding the running binary. Under go run the
// binary sits in a throwaway build directory, so the working directory is used
// instead.
func BasePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return basePathFor(exe)
}

func basePathFor(exe string) (string, error) {
	resolved, err := filepath.EvalSymlinks(exe)
	if err == nil {
		exe = resolved
	}

	if isGoRunTempBinary(exe) {
		return os.Getwd()
	}

	return filepath.Dir(exe), nil
}

// isGoRunTempBinary spots the throwaway binary path go run builds under
// /tmp/go-build.../exe/.
func isGoRunTempBinary(path string) bool {
	dir := filepath.Dir(path)
	base := filepath.Base(dir)

	return strings.Contains(path, "go-build") ||
		strings.HasPrefix(base, "exe") && strings.Contains(dir, string(os.PathSeparator)+"b0")
}
