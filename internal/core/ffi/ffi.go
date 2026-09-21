package ffi

import "unsafe"

// GoString converts a NUL-terminated C string at p to a Go string.
//
// The returned string is copied and does not depend on the lifetime of the
// C string.
func GoString(p uintptr) string {
	if p == 0 {
		return ""
	}

	ptr := *(*unsafe.Pointer)(unsafe.Pointer(&p))
	n := 0

	for *(*byte)(unsafe.Add(ptr, n)) != 0 {
		n++
	}

	return string(unsafe.Slice((*byte)(ptr), n))
}

// CString returns a NUL-terminated byte slice suitable for passing to C.
//
// The returned slice is owned by Go and must remain alive while C uses it.
func CString(s string) []byte {
	return append([]byte(s), 0)
}

// OptionalCString returns nil for an empty string, otherwise a pointer to a
// NUL-terminated C string. The backing slice is added to keepAlive so it
// remains reachable by Go while C is using it.
func OptionalCString(s string, keepAlive *[]any) *byte {
	if s == "" {
		return nil
	}

	b := CString(s)
	*keepAlive = append(*keepAlive, b)

	return &b[0]
}
