package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func makeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestIsGoRunTempBinary(t *testing.T) {
	sep := string(os.PathSeparator)
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"go-build in path", filepath.Join(sep+"tmp", "go-build123", "b001", "app.test"), true},
		{"exe dir under b0", filepath.Join(sep+"tmp", "x", "b001", "exe", "app"), true},
		{"exe dir but no b0", filepath.Join(sep+"tmp", "x", "exe", "app"), false},
		{"normal install", filepath.Join(sep+"usr", "local", "bin", "app"), false},
		{"b0 dir but not exe", filepath.Join(sep+"tmp", "b001", "app"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isGoRunTempBinary(tt.path); got != tt.want {
				t.Fatalf("isGoRunTempBinary(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestBasePathFor_RealBinary(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "app")
	makeFile(t, bin)

	got, err := basePathFor(bin)
	if err != nil {
		t.Fatal(err)
	}

	want, _ := filepath.EvalSymlinks(dir)
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBasePathFor_Symlink(t *testing.T) {
	realDir := t.TempDir()
	bin := filepath.Join(realDir, "app")
	makeFile(t, bin)

	link := filepath.Join(t.TempDir(), "app-link")
	if err := os.Symlink(bin, link); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}

	got, err := basePathFor(link)
	if err != nil {
		t.Fatal(err)
	}

	want, _ := filepath.EvalSymlinks(realDir)
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBasePathFor_GoRunBinary(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "go-build123", "b001", "exe", "app")
	makeFile(t, bin)

	got, err := basePathFor(bin)
	if err != nil {
		t.Fatal(err)
	}

	wd, _ := os.Getwd()
	if got != wd {
		t.Fatalf("got %q, want working dir %q", got, wd)
	}
}

func TestBasePathFor_UnresolvableFallsBack(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "does-not-exist")

	got, err := basePathFor(missing)
	if err != nil {
		t.Fatal(err)
	}

	// EvalSymlinks fails, so the original path's directory is used.
	if got != dir {
		t.Fatalf("got %q, want %q", got, dir)
	}
}

func TestBasePath_Smoke(t *testing.T) {
	got, err := BasePath()
	if err != nil {
		t.Fatal(err)
	}
	if got == "" {
		t.Fatal("empty base path")
	}
}
