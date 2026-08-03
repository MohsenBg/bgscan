//go:build windows

package process

import (
	"os/exec"
)

// setSysProcAttr is a no-op on Windows; process groups and Setpgid are not
// supported.
func setSysProcAttr(cmd *exec.Cmd) {}

// signalTerminate falls back to Kill on Windows because SIGTERM is not
// supported.
func signalTerminate(cmd *exec.Cmd) {
	_ = cmd.Process.Kill()
}

// killProcess forcefully terminates the process on Windows.
func killProcess(cmd *exec.Cmd) {
	_ = cmd.Process.Kill()
}
