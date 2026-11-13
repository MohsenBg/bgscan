//go:build !windows

package fileutil

import (
	"fmt"
	"os"
)

// SyncDir flushes a directory entry. Without it a successful rename can still
// vanish in a power cut. Some filesystems refuse the call, which is why the
// error is reported instead of ignored.
func SyncDir(dir string) error {
	df, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("open directory %q: %w", dir, err)
	}
	defer func() { _ = df.Close() }()

	if err := df.Sync(); err != nil {
		return fmt.Errorf("sync directory %q: %w", dir, err)
	}

	return nil
}
