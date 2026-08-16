//go:build linux || darwin || android

package config

import "syscall"

func getFDLimit() uint64 {
	var rlimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit); err != nil {
		return 1024
	}
	return rlimit.Cur
}
