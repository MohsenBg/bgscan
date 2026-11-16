package migration

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
	"github.com/MohsenBg/bgscan/internal/core/netutil"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

// MigrateTargets upgrades target lists under baseDir to the canonical layout.
func (m *migrationService) MigrateTargets() (TargetReport, error) {
	var report TargetReport

	markerDir := filepath.Join(m.baseDir, filepath.Dir(target.IPListDir))
	version, found, err := ReadVersion(markerDir)
	if err != nil {
		return report, err
	}
	if found && version > TargetVersion {
		return report, fmt.Errorf(
			"target lists use format version %d, newer than this build supports (%d)",
			version, TargetVersion,
		)
	}
	if found && version == TargetVersion {
		return report, nil
	}

	if err := m.moveV1Targets(); err != nil {
		return report, err
	}

	// Domains only exist in V2, so only IP lists need upgrading.
	dir := filepath.Join(m.baseDir, target.IPListDir)
	files, err := fileutil.ListFiles(dir, func(name string, _ os.FileInfo) bool {
		return fileutil.HasExt(name, ".csv")
	})
	if err != nil {
		// A legacy install has ips/ but no targets/ yet, and a fresh install
		// may have neither. Neither is an error.
		if errors.Is(err, os.ErrNotExist) {
			return report, WriteVersion(markerDir, TargetVersion)
		}
		return report, fmt.Errorf("list target files: %w", err)
	}

	for _, file := range files {
		upgraded, err := m.migrateTargetFile(file.Path)
		if err != nil {
			return report, err
		}
		if upgraded {
			report.Upgraded++
		}
	}

	if err := WriteVersion(markerDir, TargetVersion); err != nil {
		return report, err
	}
	return report, nil
}

// moveV1Targets moves files from legacy "ips/" into "targets/ips/".
func (m *migrationService) moveV1Targets() error {
	srcDir := filepath.Join(m.baseDir, legacyIPDir)

	files, err := fileutil.ListFiles(srcDir, func(name string, _ os.FileInfo) bool {
		return fileutil.HasExt(name, ".csv")
	})
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("list legacy target files: %w", err)
	}
	if len(files) == 0 {
		return nil
	}

	dstDir := filepath.Join(m.baseDir, target.IPListDir)
	if err := fileutil.EnsureDir(dstDir); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	copied := false
	for _, file := range files {
		dst := filepath.Join(dstDir, file.Name)
		if fileutil.FileExists(dst) {
			continue
		}
		if err := fileutil.CopyFileAtomic(file.Path, dst, fileutil.DefaultFileMode); err != nil {
			return fmt.Errorf("copy legacy target list %q: %w", file.Name, err)
		}
		_ = fileutil.RemoveFile(file.Path)
		copied = true
	}

	if copied {
		if err := fileutil.SyncDir(dstDir); err != nil {
			return fmt.Errorf("sync target directory: %w", err)
		}
	}
	return nil
}

// migrateTargetFile upgrades a single list in place. It reports whether the
// file was rewritten.
func (m *migrationService) migrateTargetFile(path string) (bool, error) {
	version, hasRows, err := detectTargetVersion(path)
	if err != nil {
		return false, fmt.Errorf("read target list %q: %w", path, err)
	}

	if !hasRows {
		if err := fileutil.RemoveFile(path); err != nil {
			return false, fmt.Errorf("remove empty target list %q: %w", path, err)
		}
		return false, nil
	}

	if version == targetV2 {
		return false, nil
	}

	return upgradeTargetV1(path)
}

// detectTargetVersion peeks at the first few rows to decide the layout.
// A small number of unknown rows (header/blank lines) is tolerated.
func detectTargetVersion(path string) (targetVersion, bool, error) {
	var (
		version targetVersion
		hasRows bool
		skipped int
	)

	err := fileutil.StreamCSVWithOffset(path, target.DefaultCSVConfig, func(rec []string, _ int64) error {
		hasRows = true

		switch {
		case isTargetV2Record(rec):
			version = targetV2
			return io.EOF
		case isTargetV1Record(rec):
			version = targetV1
			return io.EOF
		}

		skipped++
		if skipped >= 5 {
			return io.EOF
		}
		return nil
	})
	if err != nil {
		return targetUnknown, hasRows, err
	}
	return version, hasRows, nil
}

func isTargetV1Record(rec []string) bool {
	_, err := parseTargetV1(rec)
	return err == nil
}

func isTargetV2Record(rec []string) bool {
	_, err := parseTargetV2(rec)
	return err == nil
}

// parseTargetV1 reads the old IP-only layout: value or value,enable.
// It is deliberately strict about column count so a canonical row cannot be
// mistaken for V1 by confusing the port with the enable flag.
func parseTargetV1(rec []string) (target.Entry, error) {
	if len(rec) < 1 || len(rec) > 2 {
		return target.Entry{}, fmt.Errorf("expected 1 or 2 fields, got %d", len(rec))
	}

	ip, ok := netutil.ParseIPOrCIDR(rec[0])
	if !ok {
		return target.Entry{}, fmt.Errorf("invalid IP or CIDR: %q", rec[0])
	}

	enable := 1
	if len(rec) == 2 {
		value, err := strconv.Atoi(strings.TrimSpace(rec[1]))
		if err != nil || (value != 0 && value != 1) {
			return target.Entry{}, fmt.Errorf("invalid enable value %q", rec[1])
		}
		enable = value
	}
	return target.New(ip, enable), nil
}

// parseTargetV2 requires the canonical four columns.
func parseTargetV2(rec []string) (target.Entry, error) {
	if len(rec) != 4 {
		return target.Entry{}, fmt.Errorf("expected 4 fields, got %d", len(rec))
	}
	return target.ParseRecord(rec)
}

// upgradeTargetV1 rewrites a legacy IP-only list into the canonical layout,
// dropping rows that cannot be parsed. If nothing valid remains, the file is
// removed.
func upgradeTargetV1(path string) (bool, error) {
	dir := filepath.Dir(path)

	tmp, tmpPath, err := fileutil.CreateTempIn(dir, ".bgscan-migrate-*")
	if err != nil {
		return false, fmt.Errorf("create temp file for %q: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		_ = fileutil.RemoveFile(tmpPath)
		return false, fmt.Errorf("close temp file for %q: %w", path, err)
	}
	defer func() { _ = fileutil.RemoveFile(tmpPath) }()

	var valid int
	err = target.WriteCSV(tmpPath, func(write func(target.Entry) error) error {
		return fileutil.StreamCSVWithOffset(path, target.DefaultCSVConfig, func(rec []string, _ int64) error {
			if len(rec) < 1 || len(rec) > 2 {
				return nil
			}
			entry, err := parseTargetV1(rec)
			if err != nil {
				return nil
			}
			valid++
			return write(entry)
		})
	})
	if err != nil {
		return false, fmt.Errorf("read target list %q: %w", path, err)
	}

	if valid == 0 {
		if err := fileutil.RemoveFile(path); err != nil {
			return false, fmt.Errorf("remove target list %q: %w", path, err)
		}
		return false, nil
	}

	if err := os.Chmod(tmpPath, fileutil.DefaultFileMode); err != nil {
		return false, fmt.Errorf("chmod temp file for %q: %w", path, err)
	}
	if err := fileutil.ReplaceFile(tmpPath, path); err != nil {
		return false, fmt.Errorf("replace target list %q: %w", path, err)
	}
	if err := fileutil.SyncDir(dir); err != nil {
		return false, err
	}
	return true, nil
}
