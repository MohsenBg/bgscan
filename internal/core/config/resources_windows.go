//go:build windows

package config

// getFDLimit returns a safe, conservative default: Windows has no
// RLIMIT_NOFILE / POSIX rlimits.
// https://stackoverflow.com/questions/729162/windows-equivalent-of-ulimit-n
func getFDLimit() uint64 {
	return 1024
}
