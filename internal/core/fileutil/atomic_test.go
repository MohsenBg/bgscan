package fileutil

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestWriteFileAtomic_CreatesFile covers write file atomic creates file.
func TestWriteFileAtomic_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "out.csv")

	if err := WriteFileAtomic(path, []byte("a,b\n"), DefaultFileMode); err != nil {
		t.Fatalf("WriteFileAtomic: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "a,b\n" {
		t.Fatalf("content = %q, want %q", data, "a,b\n")
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat: %v", err)
		}
		if got := info.Mode().Perm(); got != DefaultFileMode {
			t.Fatalf("mode = %o, want %o", got, DefaultFileMode)
		}
	}

	assertNoTempFiles(t, dir)
}

// TestWriteFileAtomic_ReplacesExisting covers write file atomic replaces existing.
func TestWriteFileAtomic_ReplacesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.csv")

	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	if err := WriteFileAtomic(path, []byte("new"), DefaultFileMode); err != nil {
		t.Fatalf("WriteFileAtomic: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "new" {
		t.Fatalf("content = %q, want %q", data, "new")
	}
}

// TestWriteFileAtomic_FunctionErrorKeepsOriginal covers write file atomic function error keeps original.
func TestWriteFileAtomic_FunctionErrorKeepsOriginal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.csv")

	if err := os.WriteFile(path, []byte("keep"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	wantErr := errors.New("boom")
	err := WriteFileAtomicFunc(path, DefaultFileMode, func(w io.Writer) error {
		if _, err := w.Write([]byte("partial")); err != nil {
			return err
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "keep" {
		t.Fatalf("content = %q, want %q", data, "keep")
	}

	assertNoTempFiles(t, dir)
}

// TestCopyFileAtomic covers copy file atomic.
func TestCopyFileAtomic(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.csv")
	dst := filepath.Join(dir, "dst.csv")

	if err := os.WriteFile(src, []byte("payload"), 0o640); err != nil {
		t.Fatalf("seed source: %v", err)
	}

	if err := CopyFileAtomic(src, dst, 0o600); err != nil {
		t.Fatalf("CopyFileAtomic: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile(dst): %v", err)
	}
	if string(data) != "payload" {
		t.Fatalf("content = %q, want %q", data, "payload")
	}

	if _, err := os.Stat(src); err != nil {
		t.Fatalf("source should still exist: %v", err)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(dst)
		if err != nil {
			t.Fatalf("Stat(dst): %v", err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Fatalf("mode = %o, want 600", got)
		}
	}
}

func assertNoTempFiles(t *testing.T, dir string) {
	t.Helper()

	matches, err := filepath.Glob(filepath.Join(dir, ".bgscan-atomic-*"))
	if err != nil {
		t.Fatalf("Glob: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("temp files left behind: %v", matches)
	}
}
