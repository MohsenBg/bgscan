package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

func TestMigrateTargets_V1OneColumn(t *testing.T) {
	base := t.TempDir()
	path := targetPath(base, target.KindIP, "one.csv")
	mustWriteFile(t, path, "1.1.1.1\n2.2.2.2/24\n")

	report, err := newService(t, base).MigrateTargets()
	if err != nil {
		t.Fatalf("MigrateTargets: %v", err)
	}
	if report.Upgraded != 1 {
		t.Fatalf("Upgraded = %d, want 1", report.Upgraded)
	}

	want := "1.1.1.1/32,0,,1\n2.2.2.0/24,0,,1\n"
	if got := mustReadFile(t, path); got != want {
		t.Fatalf("content = %q, want %q", got, want)
	}
}

func TestMigrateTargets_V1TwoColumns(t *testing.T) {
	base := t.TempDir()
	path := targetPath(base, target.KindIP, "two.csv")
	mustWriteFile(t, path, "1.1.1.1,1\n2.2.2.2,0\n")

	report, err := newService(t, base).MigrateTargets()
	if err != nil {
		t.Fatalf("MigrateTargets: %v", err)
	}
	if report.Upgraded != 1 {
		t.Fatalf("Upgraded = %d, want 1", report.Upgraded)
	}

	want := "1.1.1.1/32,0,,1\n2.2.2.2/32,0,,0\n"
	if got := mustReadFile(t, path); got != want {
		t.Fatalf("content = %q, want %q", got, want)
	}
}

func TestMigrateTargets_V2Untouched(t *testing.T) {
	base := t.TempDir()
	path := targetPath(base, target.KindIP, "canonical.csv")
	canonical := "1.1.1.1/32,443,/x,1\n"
	mustWriteFile(t, path, canonical)

	report, err := newService(t, base).MigrateTargets()
	if err != nil {
		t.Fatalf("MigrateTargets: %v", err)
	}
	if report.Upgraded != 0 {
		t.Fatalf("Upgraded = %d, want 0", report.Upgraded)
	}
	if got := mustReadFile(t, path); got != canonical {
		t.Fatalf("canonical file was rewritten: %q", got)
	}
}

func TestMigrateTargets_DropsInvalidRows(t *testing.T) {
	base := t.TempDir()
	path := targetPath(base, target.KindIP, "mixed.csv")
	mustWriteFile(t, path, "1.1.1.1,1\nbad-ip,1\n2.2.2.2,1\n")

	report, err := newService(t, base).MigrateTargets()
	if err != nil {
		t.Fatalf("MigrateTargets: %v", err)
	}
	if report.Upgraded != 1 {
		t.Fatalf("Upgraded = %d, want 1", report.Upgraded)
	}

	want := "1.1.1.1/32,0,,1\n2.2.2.2/32,0,,1\n"
	if got := mustReadFile(t, path); got != want {
		t.Fatalf("content = %q, want %q", got, want)
	}
}

func TestMigrateTargets_DropsInvalidEnableValues(t *testing.T) {
	base := t.TempDir()
	path := targetPath(base, target.KindIP, "enable.csv")
	mustWriteFile(t, path, "1.1.1.1,2\n2.2.2.2,1\n")

	if _, err := newService(t, base).MigrateTargets(); err != nil {
		t.Fatalf("MigrateTargets: %v", err)
	}

	if got, want := mustReadFile(t, path), "2.2.2.2/32,0,,1\n"; got != want {
		t.Fatalf("content = %q, want %q", got, want)
	}
}

func TestMigrateTargets_DeletesFileWithNoValidRows(t *testing.T) {
	base := t.TempDir()
	path := targetPath(base, target.KindIP, "useless.csv")
	mustWriteFile(t, path, "bad-ip,1\nnope,0\n")

	if _, err := newService(t, base).MigrateTargets(); err != nil {
		t.Fatalf("MigrateTargets: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("file with no valid rows was not deleted: %v", err)
	}
}

func TestMigrateTargets_DeletesEmptyFile(t *testing.T) {
	base := t.TempDir()
	path := targetPath(base, target.KindIP, "empty.csv")
	mustWriteFile(t, path, "")

	if _, err := newService(t, base).MigrateTargets(); err != nil {
		t.Fatalf("MigrateTargets: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("empty file was not deleted: %v", err)
	}
}

func TestMigrateTargets_IgnoresNonCSVFiles(t *testing.T) {
	base := t.TempDir()
	ipDir := filepath.Join(base, target.IPListDir)

	mustWriteFile(t, filepath.Join(ipDir, "notes.txt"), "1.1.1.1,1\n")
	mustWriteFile(t, filepath.Join(ipDir, "list.csv.default"), "1.1.1.1,1\n")

	if _, err := newService(t, base).MigrateTargets(); err != nil {
		t.Fatalf("MigrateTargets: %v", err)
	}

	if got := mustReadFile(t, filepath.Join(ipDir, "notes.txt")); got != "1.1.1.1,1\n" {
		t.Fatalf("non-CSV file was changed: %q", got)
	}
	if got := mustReadFile(t, filepath.Join(ipDir, "list.csv.default")); got != "1.1.1.1,1\n" {
		t.Fatalf(".default file was changed: %q", got)
	}
}

func TestMigrateTargets_LeavesDomainDirectoryAlone(t *testing.T) {
	base := t.TempDir()
	path := targetPath(base, target.KindDomain, "domains.csv")
	content := "example.com,1\n"
	mustWriteFile(t, path, content)

	if _, err := newService(t, base).MigrateTargets(); err != nil {
		t.Fatalf("MigrateTargets: %v", err)
	}

	if got := mustReadFile(t, path); got != content {
		t.Fatalf("domain list was changed: %q", got)
	}
}

func TestMigrateTargets_WritesMarkerAfterPass(t *testing.T) {
	base := t.TempDir()
	mustWriteFile(t, targetPath(base, target.KindIP, "list.csv"), "1.1.1.1,1\n")

	if _, err := newService(t, base).MigrateTargets(); err != nil {
		t.Fatalf("MigrateTargets: %v", err)
	}

	marker := filepath.Join(base, filepath.Dir(target.IPListDir), FormatMarker)
	if got, want := mustReadFile(t, marker), "2\n"; got != want {
		t.Fatalf("marker = %q, want %q", got, want)
	}
}

func TestMigrateTargets_MarkerStopsSecondPass(t *testing.T) {
	base := t.TempDir()
	path := targetPath(base, target.KindIP, "list.csv")
	mustWriteFile(t, path, "1.1.1.1/32,0,,1\n")

	if _, err := newService(t, base).MigrateTargets(); err != nil {
		t.Fatalf("first pass: %v", err)
	}

	legacy := "1.1.1.1,1\n"
	mustWriteFile(t, path, legacy)

	report, err := newService(t, base).MigrateTargets()
	if err != nil {
		t.Fatalf("second pass: %v", err)
	}
	if report.Upgraded != 0 {
		t.Fatalf("Upgraded = %d, want 0 once the marker is current", report.Upgraded)
	}
	if got := mustReadFile(t, path); got != legacy {
		t.Fatalf("file was rewritten after the marker was current: %q", got)
	}
}

func TestMigrateTargets_RejectsNewerMarkerVersion(t *testing.T) {
	base := t.TempDir()
	if err := WriteVersion(filepath.Join(base, filepath.Dir(target.IPListDir)), TargetVersion+1); err != nil {
		t.Fatalf("WriteVersion: %v", err)
	}

	if _, err := newService(t, base).MigrateTargets(); err == nil {
		t.Fatal("expected an error for a newer format version")
	}
}

func TestMigrateTargets_MovesLegacyIPLists(t *testing.T) {
	base := t.TempDir()
	legacy := filepath.Join(base, legacyIPDir, "legacy.csv")
	mustWriteFile(t, legacy, "1.1.1.1,1\n")
	mustWriteFile(t, filepath.Join(base, legacyIPDir, "notes.txt"), "ignore me\n")
	mustWriteFile(t, filepath.Join(base, legacyIPDir, "list.csv.default"), "1.1.1.1,1\n")

	if _, err := newService(t, base).MigrateTargets(); err != nil {
		t.Fatalf("MigrateTargets: %v", err)
	}

	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatalf("legacy CSV should be consumed: %v", err)
	}
	if got := mustReadFile(t, filepath.Join(base, legacyIPDir, "notes.txt")); got != "ignore me\n" {
		t.Fatalf("non-CSV file was touched: %q", got)
	}
	if got := mustReadFile(t, filepath.Join(base, legacyIPDir, "list.csv.default")); got != "1.1.1.1,1\n" {
		t.Fatalf(".default template was touched: %q", got)
	}

	dst := targetPath(base, target.KindIP, "legacy.csv")
	if got, want := mustReadFile(t, dst), "1.1.1.1/32,0,,1\n"; got != want {
		t.Fatalf("migrated list = %q, want %q", got, want)
	}
}

func TestMigrateTargets_LegacyMoveDoesNotOverwrite(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, legacyIPDir, "same.csv")
	mustWriteFile(t, src, "1.1.1.1,1\n")

	dst := targetPath(base, target.KindIP, "same.csv")
	mustWriteFile(t, dst, "9.9.9.9/32,0,,1\n")

	if _, err := newService(t, base).MigrateTargets(); err != nil {
		t.Fatalf("MigrateTargets: %v", err)
	}

	if got, want := mustReadFile(t, dst), "9.9.9.9/32,0,,1\n"; got != want {
		t.Fatalf("existing list was overwritten: %q", got)
	}
	if _, err := os.Stat(src); err != nil {
		t.Fatalf("legacy file should be kept when the destination exists: %v", err)
	}
}

func TestMigrateTargets_LegacyMoveSkippedWhenMarkerCurrent(t *testing.T) {
	base := t.TempDir()
	if err := WriteVersion(filepath.Join(base, filepath.Dir(target.IPListDir)), TargetVersion); err != nil {
		t.Fatalf("WriteVersion: %v", err)
	}

	src := filepath.Join(base, legacyIPDir, "late.csv")
	mustWriteFile(t, src, "1.1.1.1,1\n")

	if _, err := newService(t, base).MigrateTargets(); err != nil {
		t.Fatalf("MigrateTargets: %v", err)
	}

	if _, err := os.Stat(src); err != nil {
		t.Fatalf("legacy file was consumed after the marker was current: %v", err)
	}
	if _, err := os.Stat(targetPath(base, target.KindIP, "late.csv")); !os.IsNotExist(err) {
		t.Fatalf("legacy file was copied after the marker was current: %v", err)
	}
}

func newService(t *testing.T, base string) MigrationService {
	t.Helper()

	svc, err := New(WithBaseDir(base))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return svc
}

func targetPath(base string, kind target.Kind, name string) string {
	dir := target.IPListDir
	if kind == target.KindDomain {
		dir = target.DomainListDir
	}
	return filepath.Join(base, dir, name)
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()

	if err := fileutil.EnsureDir(filepath.Dir(path)); err != nil {
		t.Fatalf("create directory for %q: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %q: %v", path, err)
	}
	return string(data)
}
