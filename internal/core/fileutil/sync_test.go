package fileutil

import "testing"

func TestSyncDir_ExistingDir(t *testing.T) {
	if err := SyncDir(t.TempDir()); err != nil {
		t.Fatalf("SyncDir() on existing dir: %v", err)
	}
}

func TestSyncDir_NonexistentDir(t *testing.T) {
	if err := SyncDir("/nonexistent/path/xyz"); err == nil {
		t.Fatal("expected error for nonexistent dir, got nil")
	}
}

func TestSyncDir_ClosesFile(t *testing.T) {
	tmpDir := t.TempDir()

	// A leaked directory fd would eventually fail here.
	for i := range 100 {
		if err := SyncDir(tmpDir); err != nil {
			t.Fatalf("SyncDir() call %d: %v", i, err)
		}
	}
}
