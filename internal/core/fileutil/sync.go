package fileutil

import (
	"fmt"
	"os"
)

// syncFile pushes f's data to stable storage.
func syncFile(f *os.File) error {
	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync file %q: %w", f.Name(), err)
	}
	return nil
}
