//go:build !windows

package process

import (
	"os/exec"
	"syscall"
)

// setSysProcAttr starts each command in its own process group so cleanup can
// signal the command and any children together.
func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// signalTerminate sends SIGTERM to the command's process group.
func signalTerminate(cmd *exec.Cmd) {
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
}

// killProcess sends SIGKILL to the command's process group.
func killProcess(cmd *exec.Cmd) {
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
