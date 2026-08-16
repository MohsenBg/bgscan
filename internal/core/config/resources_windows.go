//go:build windows

package config

// Windows doesn't have RLIMIT_NOFILE / POSIX rlimits.
// Return a safe, conservative default instead.
// https://stackoverflow.com/questions/729162/windows-equivalent-of-ulimit-n
func getFDLimit() uint64 {
	return 1024
}
