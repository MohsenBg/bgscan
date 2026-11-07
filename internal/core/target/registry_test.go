package target

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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

// setBaseDir redirects the application base directory to a temp dir and
// returns that base directory.
func setBaseDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	old := baseDirOverride
	baseDirOverride = dir

	t.Cleanup(func() {
		baseDirOverride = old
	})

	return dir
}

func targetDir(base string, kind Kind) string {
	switch kind {
	case KindDomain:
		return filepath.Join(base, DomainListDir)
	default:
		return filepath.Join(base, IPListDir)
	}
}

func TestResolveTargetFilePath(t *testing.T) {
	base := setBaseDir(t)

	tests := []struct {
		name string
		kind Kind
		want string
	}{
		{"IP", KindIP, filepath.Join(base, IPListDir, "mylist.csv")},
		{"Domain", KindDomain, filepath.Join(base, DomainListDir, "mylist.csv")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveTargetFilePath("mylist", tt.kind)
			if err != nil {
				t.Fatalf("ResolveTargetFilePath() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ResolveTargetFilePath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTargetFileExists(t *testing.T) {
	base := setBaseDir(t)

	ipPath := filepath.Join(base, IPListDir, "mylist.csv")
	mustWriteFile(t, ipPath, "1.1.1.1,1\n")

	if !TargetFileExists("mylist", KindIP) {
		t.Fatalf("TargetFileExists(%q, IP) = false, want true", "mylist")
	}
	if !TargetFileExists("mylist.csv", KindIP) {
		t.Fatalf("TargetFileExists(%q, IP) = false, want true", "mylist.csv")
	}
	if TargetFileExists("mylist", KindDomain) {
		t.Fatalf("TargetFileExists(%q, Domain) = true, want false", "mylist")
	}
	if TargetFileExists("missing", KindIP) {
		t.Fatalf("TargetFileExists(%q, IP) = true, want false", "missing")
	}
}

func TestGetTargetFileInfo_ByName(t *testing.T) {
	base := setBaseDir(t)

	p := filepath.Join(base, IPListDir, "alpha.csv")
	mustWriteFile(t, p, "1.1.1.1,1\n")

	info, err := GetTargetFileInfo("alpha", KindIP)
	if err != nil {
		t.Fatalf("GetTargetFileInfo() error = %v", err)
	}

	if info.Kind != KindIP {
		t.Fatalf("Kind = %v, want %v", info.Kind, KindIP)
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

func TestGetTargetFileInfo_DomainByName(t *testing.T) {
	base := setBaseDir(t)

	p := filepath.Join(base, DomainListDir, "example.csv")
	mustWriteFile(t, p, "example.com,1\n")

	info, err := GetTargetFileInfo("example", KindDomain)
	if err != nil {
		t.Fatalf("GetTargetFileInfo() error = %v", err)
	}

	if info.Kind != KindDomain {
		t.Fatalf("Kind = %v, want %v", info.Kind, KindDomain)
	}
	if info.Path != p {
		t.Fatalf("Path = %q, want %q", info.Path, p)
	}
}

func TestGetTargetFileInfo_ByFilenameWithExt(t *testing.T) {
	base := setBaseDir(t)

	p := filepath.Join(base, IPListDir, "beta.csv")
	mustWriteFile(t, p, "2.2.2.2,0\n")

	info, err := GetTargetFileInfo("beta.csv", KindIP)
	if err != nil {
		t.Fatalf("GetTargetFileInfo() error = %v", err)
	}

	if info.Name != "beta" {
		t.Fatalf("Name = %q, want %q", info.Name, "beta")
	}
	if info.Path != p {
		t.Fatalf("Path = %q, want %q", info.Path, p)
	}
}

func TestGetTargetFileInfo_ByAbsolutePath(t *testing.T) {
	base := setBaseDir(t)

	p := filepath.Join(base, "abs.csv")
	mustWriteFile(t, p, "3.3.3.3,1\n")

	info, err := GetTargetFileInfo(p, KindIP)
	if err != nil {
		t.Fatalf("GetTargetFileInfo() error = %v", err)
	}

	if info.Name != "abs" {
		t.Fatalf("Name = %q, want %q", info.Name, "abs")
	}
	if info.Path != p {
		t.Fatalf("Path = %q, want %q", info.Path, p)
	}
}

func TestGetTargetFileInfo_NotFound(t *testing.T) {
	setBaseDir(t)

	_, err := GetTargetFileInfo("missing", KindIP)
	if !errors.Is(err, ErrTargetFileNotFound) {
		t.Fatalf("GetTargetFileInfo() error = %v, want ErrTargetFileNotFound", err)
	}
}

func TestGetTargetFileInfo_NotRegularFile(t *testing.T) {
	base := setBaseDir(t)

	dirPath := filepath.Join(base, IPListDir, "dir.csv")
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", dirPath, err)
	}

	_, err := GetTargetFileInfo("dir", KindIP)
	if err == nil {
		t.Fatal("expected error for non-regular file, got nil")
	}
}

func TestListFiles_SortsNewestFirstAndFiltersCSV(t *testing.T) {
	base := setBaseDir(t)

	oldTime := time.Now().Add(-2 * time.Hour)
	midTime := time.Now().Add(-1 * time.Hour)
	newTime := time.Now()

	ipOld := filepath.Join(base, IPListDir, "old.csv")
	domainMid := filepath.Join(base, DomainListDir, "mid.csv")
	domainNew := filepath.Join(base, DomainListDir, "new.csv")
	ignore := filepath.Join(base, IPListDir, "ignore.txt")

	mustWriteFile(t, ipOld, "1.1.1.1,1\n")
	mustWriteFile(t, domainMid, "mid.example,1\n")
	mustWriteFile(t, domainNew, "new.example,1\n")
	mustWriteFile(t, ignore, "x\n")

	mustSetModTime(t, ipOld, oldTime)
	mustSetModTime(t, domainMid, midTime)
	mustSetModTime(t, domainNew, newTime)

	got, err := ListFiles()
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	want := []struct {
		name string
		kind Kind
	}{
		{"new", KindDomain},
		{"mid", KindDomain},
		{"old", KindIP},
	}

	if len(got) != len(want) {
		t.Fatalf("len(ListFiles()) = %d, want %d", len(got), len(want))
	}

	for i, w := range want {
		if got[i].Name != w.name || got[i].Kind != w.kind {
			t.Fatalf("got[%d] = (%q, %v), want (%q, %v)", i, got[i].Name, got[i].Kind, w.name, w.kind)
		}
		if filepath.Ext(got[i].Path) != ".csv" {
			t.Fatalf("unexpected non-csv path returned: %q", got[i].Path)
		}
	}
}

func TestListFiles_MissingDirectories(t *testing.T) {
	setBaseDir(t)

	got, err := ListFiles()
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("ListFiles() len = %d, want 0", len(got))
	}
}

func TestResolveTargetFilePathRejectsInvalidNames(t *testing.T) {
	setBaseDir(t)

	invalid := []string{"", "   ", ".", "..", "../x", "../../etc/passwd", "a/b", `a\b`}

	for _, kind := range []Kind{KindIP, KindDomain} {
		for _, name := range invalid {
			path, err := ResolveTargetFilePath(name, kind)
			if err == nil {
				t.Errorf("ResolveTargetFilePath(%q, %v) = %q, want error", name, kind, path)
				continue
			}

			if !strings.Contains(err.Error(), "invalid target list name") {
				t.Errorf("ResolveTargetFilePath(%q, %v) error = %v, want invalid target list name", name, kind, err)
			}

			if _, err := GetTargetFileInfo(name, kind); err == nil {
				t.Errorf("GetTargetFileInfo(%q, %v) should fail", name, kind)
			}
		}
	}
}

func TestResolveTargetFilePathAcceptsValidNames(t *testing.T) {
	base := setBaseDir(t)

	valid := map[string]string{
		"mylist":     "mylist.csv",
		"mylist.csv": "mylist.csv",
		"MY.CSV":     "MY.CSV",
	}

	for _, kind := range []Kind{KindIP, KindDomain} {
		for name, file := range valid {
			path, err := ResolveTargetFilePath(name, kind)
			if err != nil {
				t.Fatalf("ResolveTargetFilePath(%q, %v) error = %v", name, kind, err)
			}

			want := filepath.Join(targetDir(base, kind), file)
			if path != want {
				t.Errorf("ResolveTargetFilePath(%q, %v) = %q, want %q", name, kind, path, want)
			}
		}
	}
}

func TestGetTargetFileInfoAcceptsAbsolutePath(t *testing.T) {
	base := setBaseDir(t)

	abs := filepath.Join(base, "absolute.csv")
	if err := os.WriteFile(abs, []byte("10.0.0.0/8,1\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	info, err := GetTargetFileInfo(abs, KindDomain)
	if err != nil {
		t.Fatalf("GetTargetFileInfo(%q) error = %v", abs, err)
	}

	if info.Path != abs {
		t.Errorf("Path = %q, want %q", info.Path, abs)
	}
	if info.Kind != KindDomain {
		t.Errorf("Kind = %v, want %v", info.Kind, KindDomain)
	}
}
