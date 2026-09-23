package netutil

import (
	"errors"
	"syscall"
)

// unreachableErrnos are local OS errnos meaning the host itself cannot
// route to the destination, as opposed to the destination refusing us.
var unreachableErrnos = []syscall.Errno{
	syscall.ENETUNREACH,
	syscall.EHOSTUNREACH,
	syscall.EADDRNOTAVAIL,
	syscall.EAFNOSUPPORT,
}

// IsUnreachable reports whether err indicates the local system cannot route
// to the destination (no route to network/host, address not available, or
// unsupported address family). Such failures are properties of the scanning
// environment, not the scanned target.
func IsUnreachable(err error) bool {
	var errno syscall.Errno
	if !errors.As(err, &errno) {
		return false
	}
	for _, u := range unreachableErrnos {
		if errno == u {
			return true
		}
	}
	return false
}
