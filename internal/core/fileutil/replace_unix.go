//go:build !windows

package fileutil

import "os"

func moveReplace(src, dst string) error {
	return os.Rename(src, dst)
}

// Unix has no sharing violations to wait out.
func isTransientLock(error) bool { return false }
