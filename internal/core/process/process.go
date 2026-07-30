package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Process represents a running external process that can be terminated
// gracefully or forcefully.
type Process interface {
	StopGracefully(timeout time.Duration) error
	Kill() error
	Wait() error
}

// process wraps an exec.Cmd and manages waiting and shutdown.
type process struct {
	cmd      *exec.Cmd
	waitOnce sync.Once
	waitErr  error
	waitDone chan struct{}
}

// Start launches a binary with the provided arguments.
//
// The returned Process owns the started command and should be stopped by the
// caller when it is no longer needed.
func Start(ctx context.Context, binary string, args ...string) (Process, error) {
	if !filepath.IsAbs(binary) {
		abs, err := filepath.Abs(binary)
		if err != nil {
			return nil, fmt.Errorf("resolve binary path: %w", err)
		}
		binary = abs
	}

	cmd := exec.CommandContext(ctx, binary, args...)
	setSysProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start process: %w", err)
	}

	p := &process{
		cmd:      cmd,
		waitDone: make(chan struct{}),
	}

	go p.backgroundWait()

	return p, nil
}

func (p *process) backgroundWait() {
	p.waitOnce.Do(func() {
		if p.cmd != nil && p.cmd.Process != nil {
			p.waitErr = p.cmd.Wait()
		}

		close(p.waitDone)
	})
}

// StopGracefully sends a termination request and waits for the process to exit.
//
// If the process does not exit before timeout, it is forcefully killed.
func (p *process) StopGracefully(timeout time.Duration) error {
	if p.cmd == nil || p.cmd.Process == nil {
		return fmt.Errorf("process not running")
	}

	select {
	case <-p.waitDone:
		return nil
	default:
	}

	signalTerminate(p.cmd)

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-p.waitDone:
		return nil

	case <-timer.C:
		return p.Kill()
	}
}

// Kill forcefully terminates the process and waits until it exits.
//
// It does not return the child's exit status; use Wait when that information
// is needed.
func (p *process) Kill() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return fmt.Errorf("process not running")
	}

	select {
	case <-p.waitDone:
		return nil
	default:
	}

	killProcess(p.cmd)

	<-p.waitDone

	return nil
}

// Wait blocks until the process exits and returns its exit status.
func (p *process) Wait() error {
	<-p.waitDone
	return p.waitErr
}

// FindBinaryInPaths searches the provided directories and then the system PATH
// for a binary.
//
// The returned path is always absolute. On Windows, ".exe" is added when the
// binary name does not already contain it.
func FindBinaryInPaths(binary string, dirs []string) (string, error) {
	if runtime.GOOS == "windows" && !strings.HasSuffix(binary, ".exe") {
		binary += ".exe"
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(dir, binary)

		info, err := os.Stat(fullPath)
		if err == nil && !info.IsDir() {
			return filepath.Abs(fullPath)
		}
	}

	path, err := exec.LookPath(binary)
	if err != nil {
		return "", fmt.Errorf("binary %q not found", binary)
	}

	return filepath.Abs(path)
}

// EnsureExecutable adds executable permissions to a file on Unix systems.
//
// Windows does not use Unix executable bits, so this function does nothing
// there.
func EnsureExecutable(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.Mode()&0o111 != 0 {
		return nil
	}

	return os.Chmod(path, info.Mode()|0o755)
}
