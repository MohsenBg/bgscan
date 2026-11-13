package fileutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// DefaultFileMode is what published data files use unless a caller asks for
// something else.
const DefaultFileMode os.FileMode = 0o644

// WriteFileAtomicFunc writes path in two steps. fn fills a temp file next to
// it, then that file is renamed over path.
//
// Anyone opening path sees the old file or the new one. There is no moment
// where the file is half written. The data is synced before the rename, and
// the directory entry after it, so a crash cannot quietly undo either step.
//
// A mode of 0 means DefaultFileMode.
func WriteFileAtomicFunc(path string, mode os.FileMode, fn func(io.Writer) error) error {
	if path == "" {
		return fmt.Errorf("write file atomically: path is required")
	}
	if mode == 0 {
		mode = DefaultFileMode
	}

	dir := filepath.Dir(path)

	f, tmpPath, err := CreateTempIn(dir, ".bgscan-atomic-*")
	if err != nil {
		return err
	}

	published := false

	defer func() {
		if published {
			return
		}
		_ = f.Close()
		_ = os.Remove(tmpPath)
	}()

	if err := fn(f); err != nil {
		return fmt.Errorf("write temp file for %q: %w", path, err)
	}

	if err := syncFile(f); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("close temp file for %q: %w", path, err)
	}

	if err := os.Chmod(tmpPath, mode); err != nil {
		return fmt.Errorf("chmod temp file for %q: %w", path, err)
	}

	if err := ReplaceFile(tmpPath, path); err != nil {
		return err
	}
	published = true

	return SyncDir(dir)
}

// WriteFileAtomic is WriteFileAtomicFunc for callers that already hold the
// bytes in memory.
func WriteFileAtomic(path string, data []byte, mode os.FileMode) error {
	return WriteFileAtomicFunc(path, mode, func(w io.Writer) error {
		_, err := w.Write(data)
		return err
	})
}

// CopyFileAtomic copies src to dst without ever exposing a partial dst. The
// source is opened read-only and left alone.
func CopyFileAtomic(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file %q: %w", src, err)
	}
	defer func() { _ = in.Close() }()

	return WriteFileAtomicFunc(dst, mode, func(w io.Writer) error {
		if _, err := io.Copy(w, in); err != nil {
			return fmt.Errorf("copy %q: %w", src, err)
		}
		return nil
	})
}
