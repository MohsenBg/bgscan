package migration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteVersion_ThenRead(t *testing.T) {
	dir := t.TempDir()

	if err := WriteVersion(dir, 2); err != nil {
		t.Fatalf("WriteVersion: %v", err)
	}

	version, found, err := ReadVersion(dir)
	if err != nil {
		t.Fatalf("ReadVersion: %v", err)
	}
	if !found {
		t.Fatal("marker should be found after WriteVersion")
	}
	if version != 2 {
		t.Fatalf("version = %d, want 2", version)
	}
}

func TestWriteVersion_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "result")

	if err := WriteVersion(dir, 1); err != nil {
		t.Fatalf("WriteVersion: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, FormatMarker)); err != nil {
		t.Fatalf("marker not created: %v", err)
	}
}

func TestWriteVersion_ContentIsOneInteger(t *testing.T) {
	dir := t.TempDir()

	if err := WriteVersion(dir, 3); err != nil {
		t.Fatalf("WriteVersion: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, FormatMarker))
	if err != nil {
		t.Fatalf("read marker: %v", err)
	}
	if got := string(data); got != "3\n" {
		t.Fatalf("marker content = %q, want %q", got, "3\n")
	}
}

func TestWriteVersion_RejectsNonPositive(t *testing.T) {
	for _, version := range []int{0, -1} {
		if err := WriteVersion(t.TempDir(), version); err == nil {
			t.Fatalf("WriteVersion(%d) should fail", version)
		}
	}
}

func TestReadVersion_MissingMarker(t *testing.T) {
	version, found, err := ReadVersion(t.TempDir())
	if err != nil {
		t.Fatalf("ReadVersion: %v", err)
	}
	if found {
		t.Fatal("found should be false when the marker is missing")
	}
	if version != 0 {
		t.Fatalf("version = %d, want 0", version)
	}
}

func TestReadVersion_EmptyMarker(t *testing.T) {
	dir := t.TempDir()
	writeMarker(t, dir, "")

	if _, found, err := ReadVersion(dir); err == nil {
		t.Fatalf("expected error, got found=%v", found)
	}
}

func TestReadVersion_MalformedMarker(t *testing.T) {
	dir := t.TempDir()
	writeMarker(t, dir, "two\n")

	if _, _, err := ReadVersion(dir); err == nil {
		t.Fatal("expected error for non-integer marker")
	}
}

func TestReadVersion_RejectsZero(t *testing.T) {
	dir := t.TempDir()
	writeMarker(t, dir, "0\n")

	if _, _, err := ReadVersion(dir); err == nil {
		t.Fatal("expected error for version zero")
	}
}

func TestReadVersion_TrimsWhitespace(t *testing.T) {
	dir := t.TempDir()
	writeMarker(t, dir, "  4  \n")

	version, found, err := ReadVersion(dir)
	if err != nil {
		t.Fatalf("ReadVersion: %v", err)
	}
	if !found || version != 4 {
		t.Fatalf("got version=%d found=%v, want 4 true", version, found)
	}
}

func TestReadVersion_RequiresDirectory(t *testing.T) {
	if _, _, err := ReadVersion(""); err == nil {
		t.Fatal("expected error for empty directory")
	}
}

func writeMarker(t *testing.T, dir, content string) {
	t.Helper()

	path := filepath.Join(dir, FormatMarker)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}
}
