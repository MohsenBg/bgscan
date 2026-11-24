//go:build !windows

package ffi

import "github.com/ebitengine/purego"

// OpenLibrary loads the named shared library and returns its handle.
// The handle stays valid for the life of the process. The caller does
// not close it.
func OpenLibrary(name string) (uintptr, error) {
	return purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
}
