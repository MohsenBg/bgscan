package result

import (
	"os"
	"path/filepath"
	"testing"
)

// ---------- replaceFile ----------

func TestReplaceFile_NewDestination(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src.csv")
	dst := filepath.Join(tmpDir, "dst.csv")

	_ = os.WriteFile(src, []byte("hello"), 0o644)

	err := replaceFile(src, dst)
	if err != nil {
		t.Fatalf("replaceFile() error = %v", err)
	}

	// src should be gone.
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

func TestReplaceFile_OverwriteExisting(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src.csv")
	dst := filepath.Join(tmpDir, "dst.csv")

	_ = os.WriteFile(src, []byte("new content"), 0o644)
	_ = os.WriteFile(dst, []byte("old content"), 0o644)

	err := replaceFile(src, dst)
	if err != nil {
		t.Fatalf("replaceFile() error = %v", err)
	}

	data, _ := os.ReadFile(dst)
	if string(data) != "new content" {
		t.Errorf("dst content = %q, want %q", string(data), "new content")
	}
}

func TestReplaceFile_NonexistentSource(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "nonexistent.csv")
	dst := filepath.Join(tmpDir, "dst.csv")

	err := replaceFile(src, dst)
	if err == nil {
		t.Fatal("expected error for nonexistent source, got nil")
	}
}

// ---------- syncDir ----------

func TestSyncDir_ExistingDir(t *testing.T) {
	tmpDir := t.TempDir()

	err := syncDir(tmpDir)
	if err != nil {
		t.Fatalf("syncDir() on existing dir: %v", err)
	}
}

func TestSyncDir_NonexistentDir(t *testing.T) {
	err := syncDir("/nonexistent/path/xyz")
	if err == nil {
		t.Fatal("expected error for nonexistent dir, got nil")
	}
}

func TestSyncDir_ClosesFile(t *testing.T) {
	tmpDir := t.TempDir()
	// syncDir should close the directory fd even on success.
	// This test verifies no fd leak by calling it many times.
	for i := range 100 {
		err := syncDir(tmpDir)
		if err != nil {
			t.Fatalf("syncDir() call %d: %v", i, err)
		}
	}
}
