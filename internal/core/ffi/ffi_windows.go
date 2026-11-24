//go:build windows

package ffi

import "syscall"

// OpenLibrary loads the named DLL and returns its handle. The handle
// stays valid for the life of the process. The caller does not close
// it.
func OpenLibrary(name string) (uintptr, error) {
	handle, err := syscall.LoadLibrary(name)
	return uintptr(handle), err
}
