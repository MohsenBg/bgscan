package netutil

import "net"

// IsTimeout reports whether the given error represents a network timeout.
func IsTimeout(err error) bool {
	if ne, ok := err.(net.Error); ok {
		return ne.Timeout()
	}
	return false
}
