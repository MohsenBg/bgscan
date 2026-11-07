package target

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
)

// LegacyIPListDir is the pre-targets location for IP lists. Older releases
// shipped their bundle there, and bgscan-builder versions that predate the
// targets layout still write generated CSV files into it.
const LegacyIPListDir = "ips"

// MigrateLegacyIPLists migrates CSV files from ips/ to targets/ips/.
// Existing files are preserved, making repeated startup calls safe.
func MigrateLegacyIPLists() (int, error) {
	base, err := getBaseRoot()
	if err != nil {
		return 0, err
	}

	legacyDir := filepath.Join(base, LegacyIPListDir)
	entries, err := os.ReadDir(legacyDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, fmt.Errorf("read legacy target directory: %w", err)
	}

	dstDir, err := getIPBaseDir()
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return 0, fmt.Errorf("create target directory: %w", err)
	}

	copied := 0

	for _, entry := range entries {
		if entry.IsDir() || !fileutil.HasExt(entry.Name(), ".csv") {
			continue
		}

		src := filepath.Join(legacyDir, entry.Name())
		dst := filepath.Join(dstDir, entry.Name())

		if _, err := os.Stat(dst); err == nil {
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return copied, fmt.Errorf("stat target list %q: %w", dst, err)
		}

		if err := copyFileAtomic(src, dst); err != nil {
			return copied, fmt.Errorf("migrate legacy target list %q: %w", entry.Name(), err)
		}

		copied++
	}

	return copied, nil
}

// copyFileAtomic copies src to dst through a temporary file in dst's
// directory, then renames it into place.
func copyFileAtomic(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	tmp, err := os.CreateTemp(filepath.Dir(dst), ".bgscan-migrate-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if _, err := io.Copy(tmp, in); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Chmod(tmpPath, 0o644); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, dst); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	return nil
}
