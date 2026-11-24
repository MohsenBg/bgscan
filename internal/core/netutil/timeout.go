package netutil

import "net"

// IsTimeout reports whether err is a network error with a timeout.
// IsTimeout returns false for nil errors and non-network errors.
func IsTimeout(err error) bool {
	if ne, ok := err.(net.Error); ok {
		return ne.Timeout()
	}
	return false
}
