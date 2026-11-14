package target

import (
	"errors"
	"fmt"
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
	files, err := fileutil.ListFiles(legacyDir, func(name string, _ os.FileInfo) bool {
		return fileutil.HasExt(name, ".csv")
	})
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
	if err := fileutil.EnsureDir(dstDir); err != nil {
		return 0, fmt.Errorf("create target directory: %w", err)
	}

	copied := 0

	for _, file := range files {
		src := file.Path
		dst := filepath.Join(dstDir, file.Name)

		if _, err := os.Stat(dst); err == nil {
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return copied, fmt.Errorf("stat target list %q: %w", dst, err)
		}

		if err := fileutil.CopyFileAtomic(src, dst, fileutil.DefaultFileMode); err != nil {
			return copied, fmt.Errorf("migrate legacy target list %q: %w", file.Name, err)
		}

		copied++
	}

	if copied > 0 {
		if err := fileutil.SyncDir(dstDir); err != nil {
			return copied, err
		}
	}

	return copied, nil
}
