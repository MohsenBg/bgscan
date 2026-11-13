//go:build windows

package fileutil

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

func moveReplace(src, dst string) error {
	from, err := windows.UTF16PtrFromString(src)
	if err != nil {
		return err
	}

	to, err := windows.UTF16PtrFromString(dst)
	if err != nil {
		return err
	}

	flags := uint32(windows.MOVEFILE_REPLACE_EXISTING | windows.MOVEFILE_WRITE_THROUGH)
	if err := windows.MoveFileEx(from, to, flags); err != nil {
		return &os.LinkError{Op: "rename", Old: src, New: dst, Err: err}
	}

	return nil
}

// isTransientLock reports errors that mean "someone has the file open, try
// again". Access denied and the two sharing errors cover the usual cases.
func isTransientLock(err error) bool {
	return errors.Is(err, windows.ERROR_ACCESS_DENIED) ||
		errors.Is(err, windows.ERROR_SHARING_VIOLATION) ||
		errors.Is(err, windows.ERROR_LOCK_VIOLATION)
}
