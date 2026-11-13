//go:build windows

package fileutil

// SyncDir does nothing on Windows. The rename already used
// MOVEFILE_WRITE_THROUGH, and a directory handle cannot be fsynced.
func SyncDir(string) error { return nil }
