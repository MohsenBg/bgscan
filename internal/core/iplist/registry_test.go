package iplist

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func chdirTempDir(t *testing.T) string {
	t.Helper()

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	tmpDir := t.TempDir()
	resolved, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", tmpDir, err)
	}
	tmpDir = resolved
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir(%q): %v", tmpDir, err)
	}

	t.Cleanup(func() {
		_ = os.Chdir(oldWD)
	})

	return tmpDir
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
}

func mustSetModTime(t *testing.T, path string, modTime time.Time) {
	t.Helper()

	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("Chtimes(%q): %v", path, err)
	}
}

// TestGetIPFilePath verifies that application-relative paths resolve from
// the working directory while running under `go test` (BasePath fallback).
func TestGetIPFilePath(t *testing.T) {
	tmpDir := chdirTempDir(t)

	got, err := GetIPFilePath("mylist")
	if err != nil {
		t.Fatalf("GetIPFilePath() error = %v", err)
	}

	want := filepath.Join(tmpDir, "ips", "mylist.csv")
	if got != want {
		t.Fatalf("GetIPFilePath() = %q, want %q", got, want)
	}
}

func TestFileExists(t *testing.T) {
	tmpDir := chdirTempDir(t)

	p := filepath.Join(tmpDir, "ips", "mylist.csv")
	mustWriteFile(t, p, "1.1.1.1,1\n")

	if !FileExists("mylist") {
		t.Fatalf("FileExists(%q) = false, want true", "mylist")
	}
	if !FileExists("mylist.csv") {
		t.Fatalf("FileExists(%q) = false, want true", "mylist.csv")
	}
	if FileExists("missing") {
		t.Fatalf("FileExists(%q) = true, want false", "missing")
	}
}

func TestGetIPFileInfo_ByName(t *testing.T) {
	tmpDir := chdirTempDir(t)

	p := filepath.Join(tmpDir, "ips", "alpha.csv")
	mustWriteFile(t, p, "1.1.1.1,1\n")

	info, err := GetIPFileInfo("alpha")
	if err != nil {
		t.Fatalf("GetIPFileInfo() error = %v", err)
	}

	if info.Name != "alpha" {
		t.Fatalf("Name = %q, want %q", info.Name, "alpha")
	}
	if info.Path != p {
		t.Fatalf("Path = %q, want %q", info.Path, p)
	}
	if info.Size == 0 {
		t.Fatalf("Size = 0, want > 0")
	}
}

func TestGetIPFileInfo_ByFilenameWithExt(t *testing.T) {
	tmpDir := chdirTempDir(t)

	p := filepath.Join(tmpDir, "ips", "beta.csv")
	mustWriteFile(t, p, "2.2.2.2,0\n")

	info, err := GetIPFileInfo("beta.csv")
	if err != nil {
		t.Fatalf("GetIPFileInfo() error = %v", err)
	}

	if info.Name != "beta" {
		t.Fatalf("Name = %q, want %q", info.Name, "beta")
	}
	if info.Path != p {
		t.Fatalf("Path = %q, want %q", info.Path, p)
	}
}

func TestGetIPFileInfo_ByAbsolutePath(t *testing.T) {
	tmpDir := chdirTempDir(t)

	p := filepath.Join(tmpDir, "abs.csv")
	mustWriteFile(t, p, "3.3.3.3,1\n")

	info, err := GetIPFileInfo(p)
	if err != nil {
		t.Fatalf("GetIPFileInfo() error = %v", err)
	}

	if info.Name != "abs" {
		t.Fatalf("Name = %q, want %q", info.Name, "abs")
	}
	if info.Path != p {
		t.Fatalf("Path = %q, want %q", info.Path, p)
	}
}

func TestGetIPFileInfo_NotFound(t *testing.T) {
	chdirTempDir(t)

	_, err := GetIPFileInfo("missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetIPFileInfo_NotRegularFile(t *testing.T) {
	chdirTempDir(t)

	dirPath := filepath.Join("ips", "dir.csv")
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", dirPath, err)
	}

	_, err := GetIPFileInfo("dir")
	if err == nil {
		t.Fatal("expected error for non-regular file, got nil")
	}
}

func TestListIPFiles_SortsNewestFirstAndFiltersCSV(t *testing.T) {
	base := setBaseDir(t)

	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	oldTime := time.Now().Add(-2 * time.Hour)
	midTime := time.Now().Add(-1 * time.Hour)
	newTime := time.Now()

	mustWriteFile(t, filepath.Join(base, "old.csv"), "1.1.1.1,1\n")
	mustWriteFile(t, filepath.Join(base, "mid.csv"), "2.2.2.2,1\n")
	mustWriteFile(t, filepath.Join(base, "new.csv"), "3.3.3.3,1\n")
	mustWriteFile(t, filepath.Join(base, "ignore.txt"), "x\n")

	mustSetModTime(t, filepath.Join(base, "old.csv"), oldTime)
	mustSetModTime(t, filepath.Join(base, "mid.csv"), midTime)
	mustSetModTime(t, filepath.Join(base, "new.csv"), newTime)

	got, err := ListIPFiles()
	if err != nil {
		t.Fatalf("ListIPFiles() error = %v", err)
	}

	wantNames := []string{"new", "mid", "old"}

	if len(got) != len(wantNames) {
		t.Fatalf(
			"len(ListIPFiles()) = %d, want %d",
			len(got),
			len(wantNames),
		)
	}

	for i, want := range wantNames {
		if got[i].Name != want {
			t.Fatalf(
				"got[%d].Name = %q, want %q",
				i,
				got[i].Name,
				want,
			)
		}
	}

	for _, file := range got {
		if filepath.Ext(file.Path) != ".csv" {
			t.Fatalf("unexpected non-csv path returned: %q", file.Path)
		}
	}
}

func TestListIPFiles_EmptyDir(t *testing.T) {
	base := setBaseDir(t)

	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	got, err := ListIPFiles()
	if err != nil {
		t.Fatalf("ListIPFiles() error = %v", err)
	}

	if len(got) != 0 {
		t.Fatalf(
			"ListIPFiles() len = %d, want 0",
			len(got),
		)
	}
}

// setBaseDir redirects the application base directory to a temp dir and
// returns the expected IP list directory inside it.
func setBaseDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	old := baseDirOverride
	baseDirOverride = dir

	t.Cleanup(func() {
		baseDirOverride = old
	})

	return filepath.Join(dir, IPListDir)
}
