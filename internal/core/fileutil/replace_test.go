package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReplaceFile_NewDestination covers replace file new destination.
func TestReplaceFile_NewDestination(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src.csv")
	dst := filepath.Join(tmpDir, "dst.csv")

	_ = os.WriteFile(src, []byte("hello"), 0o644)

	if err := ReplaceFile(src, dst); err != nil {
		t.Fatalf("ReplaceFile() error = %v", err)
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("source file should be removed after rename")
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile(dst): %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("dst content = %q, want %q", string(data), "hello")
	}
}

// TestReplaceFile_OverwriteExisting covers replace file overwrite existing.
func TestReplaceFile_OverwriteExisting(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src.csv")
	dst := filepath.Join(tmpDir, "dst.csv")

	_ = os.WriteFile(src, []byte("new content"), 0o644)
	_ = os.WriteFile(dst, []byte("old content"), 0o644)

	if err := ReplaceFile(src, dst); err != nil {
		t.Fatalf("ReplaceFile() error = %v", err)
	}

	data, _ := os.ReadFile(dst)
	if string(data) != "new content" {
		t.Errorf("dst content = %q, want %q", string(data), "new content")
	}
}

// TestReplaceFile_NonexistentSource covers replace file nonexistent source.
func TestReplaceFile_NonexistentSource(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "nonexistent.csv")
	dst := filepath.Join(tmpDir, "dst.csv")

	if err := ReplaceFile(src, dst); err == nil {
		t.Fatal("expected error for nonexistent source, got nil")
	}
}
