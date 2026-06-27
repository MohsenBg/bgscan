package result

import (
	"os"

	"bgscan/internal/logger"
)

// replaceFile atomically replaces the destination file with the source file.
//
// On Unix-like systems os.Rename provides atomic replacement.
// On Windows, Rename fails if the destination already exists,
// so we fall back to removing the destination first.
func replaceFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// Windows fallback: remove destination then retry rename
	_ = os.Remove(dst)
	return os.Rename(src, dst)
}

// syncDir flushes directory metadata to disk.
//
// This ensures file creations/renames are persisted after a crash.
// On some platforms (notably Windows) this may effectively be a no‑op.
func syncDir(dir string) error {
	df, err := os.Open(dir)
	if err != nil {
		return err
	}

	defer func() {
		if err := df.Close(); err != nil {
			logger.CoreError("error closing file: %v", err)
		}
	}()

	return df.Sync()
}
