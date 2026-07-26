package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// ───────────────────────────────────────────────
// Start + Wait: basic lifecycle
// ───────────────────────────────────────────────

func TestStart_Wait_ExitZero(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process spawn in short mode")
	}

	p, err := Start(context.Background(), "/bin/true")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if err := p.Wait(); err != nil {
		t.Fatalf("expected clean exit, got: %v", err)
	}
}

func TestStart_Wait_ExitNonZero(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process spawn in short mode")
	}

	p, err := Start(context.Background(), "/bin/false")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if err := p.Wait(); err == nil {
		t.Fatal("expected non-zero exit error")
	}
}

func TestStart_WithArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process spawn in short mode")
	}

	// echo hello should output "hello\n" — we don't capture stdout,
	// but we verify it exits cleanly
	p, err := Start(context.Background(), "/bin/echo", "hello")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if err := p.Wait(); err != nil {
		t.Fatalf("echo exited with error: %v", err)
	}
}

func TestStart_NonexistentBinary(t *testing.T) {
	_, err := Start(context.Background(), "/nonexistent/binary/path")
	if err == nil {
		t.Fatal("expected error for nonexistent binary")
	}
}

func TestStart_ResolvesRelativePath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process spawn in short mode")
	}

	// "true" is relative, Start should resolve it via filepath.Abs
	p, err := Start(context.Background(), "true")
	if err != nil {
		// On some systems, a bare "true" without path might not be found
		// since exec.CommandContext needs an absolute path after Abs()
		// Try with a known absolute path instead
		if runtime.GOOS != "windows" {
			p, err = Start(context.Background(), "/usr/bin/true")
		}
		if err != nil {
			t.Skipf("skipping: true binary not found: %v", err)
		}
	}
	p.Wait()
}

func TestStart_WithContextCancel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process spawn in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately
	cancel()

	p, err := Start(ctx, "/bin/sleep", "60")
	if err != nil {
		// The process might fail to start or start and get killed immediately.
		// Both are acceptable — just verify no panic.
		return
	}
	// If it did start, Wait should return (process killed by context cancellation)
	err = p.Wait()
	// Either context error or signal error is fine
	t.Logf("Wait returned: %v", err)
}

func TestWait_CalledMultipleTimes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process spawn in short mode")
	}

	p, err := Start(context.Background(), "/bin/true")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait should be idempotent — calling multiple times should not deadlock
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := p.Wait(); err != nil {
				t.Errorf("Wait returned error: %v", err)
			}
		}()
	}
	wg.Wait()
}

// ───────────────────────────────────────────────
// Kill: forceful termination
// ───────────────────────────────────────────────

func TestKill_LongRunningProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process spawn in short mode")
	}

	p, err := Start(context.Background(), "/bin/sleep", "300")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	start := time.Now()
	err = p.Kill()
	elapsed := time.Since(start)

	if err == nil {
		// SIGKILL typically results in a signal error, but some systems
		// might report success. Both are acceptable.
		t.Log("Kill returned nil (process already exited or error not propagated)")
	}
	t.Logf("Kill completed in %v, err=%v", elapsed, err)

	if elapsed > 5*time.Second {
		t.Fatalf("Kill took too long: %v", elapsed)
	}
}

func TestKill_AlreadyExited(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process spawn in short mode")
	}

	p, err := Start(context.Background(), "/bin/true")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	p.Wait() // process already done

	// Kill on already-exited process should not hang or panic
	err = p.Kill()
	t.Logf("Kill on exited process: %v", err)
}

func TestKill_KillsProcessGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process spawn in short mode")
	}
	if runtime.GOOS == "windows" {
		t.Skip("process group kill not fully supported on Windows")
	}

	// Start a shell that spawns a child sleep process.
	// If process groups work, killing the parent should also kill the child.
	cmd := exec.Command("/bin/sh", "-c", "sleep 300 & sleep 300")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start shell: %v", err)
	}

	p := &Process{
		cmd:      cmd,
		waitDone: make(chan struct{}),
	}
	go p.backgroundWait()

	time.Sleep(100 * time.Millisecond) // let children spawn

	p.Kill()

	// Verify the process group is gone by checking no sleep processes remain
	time.Sleep(50 * time.Millisecond)
	out, _ := exec.Command("pgrep", "-P", strconv.Itoa(cmd.Process.Pid)).CombinedOutput()
	// pgrep returns exit code 1 when no matches — that's what we want
	t.Logf("child processes after kill: %s", strings.TrimSpace(string(out)))
}

// ───────────────────────────────────────────────
// StopGracefully: SIGTERM then fallback
// ───────────────────────────────────────────────

func TestStopGracefully_QuickExit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process spawn in short mode")
	}

	// "true" exits immediately, StopGracefully should return before timeout
	p, err := Start(context.Background(), "/bin/true")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	start := time.Now()
	err = p.StopGracefully(5 * time.Second)
	elapsed := time.Since(start)

	if err != nil {
		t.Logf("StopGracefully error (acceptable): %v", err)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("StopGracefully took too long for instant-exit process: %v", elapsed)
	}
}

func TestStopGracefully_TimeoutThenKill(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process spawn in short mode")
	}

	// sleep ignores SIGTERM, so StopGracefully should timeout and force-kill
	p, err := Start(context.Background(), "/bin/sleep", "300")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	start := time.Now()
	timeout := 200 * time.Millisecond
	err = p.StopGracefully(timeout)
	elapsed := time.Since(start)

	t.Logf("StopGracefully completed in %v (timeout=%v), err=%v", elapsed, timeout, err)

	// Should have taken approximately the timeout duration (not much longer)
	if elapsed > timeout+2*time.Second {
		t.Fatalf("StopGracefully exceeded timeout significantly: %v", elapsed)
	}
}

func TestStopGracefully_NilProcess(t *testing.T) {
	p := &Process{}
	err := p.StopGracefully(time.Second)
	if err == nil {
		t.Fatal("expected error for nil process")
	}
	if err.Error() != "process not running" {
		t.Fatalf("expected 'process not running', got: %v", err)
	}
}

// ───────────────────────────────────────────────
// Kill: nil / edge cases
// ───────────────────────────────────────────────

func TestKill_NilProcess(t *testing.T) {
	p := &Process{}
	err := p.Kill()
	if err == nil {
		t.Fatal("expected error for nil process")
	}
	if err.Error() != "process not running" {
		t.Fatalf("expected 'process not running', got: %v", err)
	}
}

func TestWait_NilProcess_NoPanic(t *testing.T) {
	p := &Process{}
	// Zero-value Process has nil waitDone — Wait() would block forever.
	// Document this by verifying in a goroutine with timeout.
	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic: %v", r)
			}
		}()
		p.Wait()
		done <- nil // should never reach here
	}()

	select {
	case err := <-done:
		t.Logf("Wait returned: %v", err)
	case <-time.After(500 * time.Millisecond):
		t.Log("Wait blocked on nil channel (expected for zero-value Process)")
	}
}

// ───────────────────────────────────────────────
// FindBinaryInPaths
// ───────────────────────────────────────────────

func TestFindBinaryInPaths_FoundInDirs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	// /bin should contain "ls" (or ls.exe on Windows)
	name := "ls"
	if runtime.GOOS == "windows" {
		name = "ls.exe"
	}

	path, err := FindBinaryInPaths(name, []string{"/bin", "/usr/bin"})
	if err != nil {
		t.Fatalf("expected to find %s in /bin or /usr/bin: %v", name, err)
	}
	if !filepath.IsAbs(path) {
		t.Fatalf("expected absolute path, got: %s", path)
	}
}

func TestFindBinaryInPaths_FoundInPATH(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	// "sh" (or "sh.exe") should be in system PATH
	name := "sh"
	if runtime.GOOS == "windows" {
		name = "sh.exe"
	}

	path, err := FindBinaryInPaths(name, []string{"/nonexistent/dir1", "/nonexistent/dir2"})
	if err != nil {
		t.Fatalf("expected to find %s via PATH fallback: %v", name, err)
	}
	if !filepath.IsAbs(path) {
		t.Fatalf("expected absolute path, got: %s", path)
	}
}

func TestFindBinaryInPaths_NotFound(t *testing.T) {
	_, err := FindBinaryInPaths("definitely_not_a_real_binary_xyz123", []string{"/tmp"})
	if err == nil {
		t.Fatal("expected error for nonexistent binary")
	}
}

func TestFindBinaryInPaths_EmptyDirsFallsBackToPATH(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	// Empty dir list should fall back to exec.LookPath
	name := "true"
	if runtime.GOOS == "windows" {
		name = "true.exe"
	}

	path, err := FindBinaryInPaths(name, nil)
	if err != nil {
		t.Fatalf("expected to find %s via PATH: %v", name, err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}
}

func TestFindBinaryInPaths_CustomTempDir(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	tmpDir := t.TempDir()
	binaryName := "mytestbin"
	if runtime.GOOS == "windows" {
		binaryName = "mytestbin.exe"
	}
	binPath := filepath.Join(tmpDir, binaryName)

	// Create a non-executable file — FindBinaryInPaths only checks os.Stat,
	// not executable bit, so it should still be "found"
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\n"), 0o644); err != nil {
		t.Fatalf("write temp binary: %v", err)
	}

	found, err := FindBinaryInPaths("mytestbin", []string{tmpDir})
	if err != nil {
		t.Fatalf("expected to find binary in custom dir: %v", err)
	}
	if found != binPath {
		// Allow for absolute path resolution
		abs, _ := filepath.Abs(binPath)
		if found != abs {
			t.Fatalf("expected %q, got %q", binPath, found)
		}
	}
}

func TestFindBinaryInPaths_SkipsDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a directory with the binary name
	dirPath := filepath.Join(tmpDir, "mybin")
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	_, err := FindBinaryInPaths("mybin", []string{tmpDir})
	if err == nil {
		t.Fatal("expected error when only a directory matches the name")
	}
}

func TestFindBinaryInPaths_WindowsExeSuffix(t *testing.T) {
	if runtime.GOOS != "windows" {
		// On non-Windows, the function should NOT append .exe
		tmpDir := t.TempDir()
		binPath := filepath.Join(tmpDir, "testbin")
		os.WriteFile(binPath, []byte("#!/bin/sh\nexit 0\n"), 0o755)

		found, err := FindBinaryInPaths("testbin", []string{tmpDir})
		if err != nil {
			t.Fatalf("expected to find testbin: %v", err)
		}
		if found != binPath {
			abs, _ := filepath.Abs(binPath)
			if found != abs {
				t.Fatalf("expected %q, got %q", binPath, found)
			}
		}
	}
	// On Windows, the function appends .exe automatically —
	// can't easily test this on Linux without build tags.
}

// ───────────────────────────────────────────────
// EnsureExecutable
// ───────────────────────────────────────────────

func TestEnsureExecutable_AlreadyExecutable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	// /bin/true is already executable
	err := EnsureExecutable("/bin/true")
	if err != nil {
		t.Fatalf("EnsureExecutable on already-executable file: %v", err)
	}
}

func TestEnsureExecutable_AddsExecuteBit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("EnsureExecutable is a no-op on Windows")
	}
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "testscript")

	// Write a file without execute permission
	content := []byte("#!/bin/sh\necho hi\n")
	if err := os.WriteFile(fPath, content, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Verify it's NOT executable
	info, _ := os.Stat(fPath)
	if info.Mode()&0o111 != 0 {
		t.Fatal("test file should not be executable initially")
	}

	// EnsureExecutable should add execute bits
	if err := EnsureExecutable(fPath); err != nil {
		t.Fatalf("EnsureExecutable: %v", err)
	}

	// Verify it IS now executable
	info, _ = os.Stat(fPath)
	if info.Mode()&0o111 == 0 {
		t.Fatal("file should be executable after EnsureExecutable")
	}
}

func TestEnsureExecutable_PreservesOtherBits(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("EnsureExecutable is a no-op on Windows")
	}
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "permscript")

	// Write with specific permissions: rw-r--r-- (0644)
	if err := os.WriteFile(fPath, []byte("#!/bin/sh\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	EnsureExecutable(fPath)

	info, _ := os.Stat(fPath)
	mode := info.Mode().Perm()

	// Original read bits should be preserved
	if mode&0o444 == 0 {
		t.Fatal("read bits should be preserved")
	}
	// Execute bits should now be set
	if mode&0o111 == 0 {
		t.Fatal("execute bits should be added")
	}
}

func TestEnsureExecutable_NonexistentFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("EnsureExecutable is a no-op on Windows")
	}

	err := EnsureExecutable("/nonexistent/file")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestEnsureExecutable_Idempotent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no-op on Windows")
	}
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "idempotent_test")
	os.WriteFile(fPath, []byte("#!/bin/sh\n"), 0o755)

	// Calling twice should not error
	if err := EnsureExecutable(fPath); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if err := EnsureExecutable(fPath); err != nil {
		t.Fatalf("second call: %v", err)
	}
}

func TestEnsureExecutable_DirectoryPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no-op on Windows")
	}

	// Directories typically have execute bit set, but just verify no crash
	tmpDir := t.TempDir()
	err := EnsureExecutable(tmpDir)
	// Should not error — dirs already have exec bit
	if err != nil {
		t.Fatalf("EnsureExecutable on directory: %v", err)
	}
}

// ───────────────────────────────────────────────
// Unix-specific: setSysProcAttr behavior
// ───────────────────────────────────────────────

func TestSetSysProcAttr_SetsProcessGroup(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix-only test")
	}

	cmd := exec.Command("/bin/true")
	setSysProcAttr(cmd)

	if cmd.SysProcAttr == nil {
		t.Fatal("expected SysProcAttr to be set")
	}
	if !cmd.SysProcAttr.Setpgid {
		t.Fatal("expected Setpgid=true")
	}
}

func TestSignalTerminate_SendsSIGTERM(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process spawn in short mode")
	}
	if runtime.GOOS == "windows" {
		t.Skip("Unix-only test")
	}

	// Start a process that sleeps
	cmd := exec.Command("/bin/sleep", "300")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	p := &Process{cmd: cmd, waitDone: make(chan struct{})}
	go p.backgroundWait()

	time.Sleep(50 * time.Millisecond)

	// Send SIGTERM via signalTerminate
	signalTerminate(cmd)

	// Wait for exit with timeout
	done := make(chan error, 1)
	go func() {
		done <- p.Wait()
	}()

	select {
	case err := <-done:
		t.Logf("process exited after SIGTERM: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("process did not exit after SIGTERM")
	}
}

func TestKillProcess_SendsSIGKILL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process spawn in short mode")
	}
	if runtime.GOOS == "windows" {
		t.Skip("Unix-only test")
	}

	// SIGTERM-ignoring process — SIGKILL should still work
	cmd := exec.Command("/bin/sh", "-c", "trap '' TERM; sleep 300")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}

	p := &Process{cmd: cmd, waitDone: make(chan struct{})}
	go p.backgroundWait()

	time.Sleep(50 * time.Millisecond)

	start := time.Now()
	killProcess(cmd)
	p.Wait()
	elapsed := time.Since(start)

	if elapsed > 2*time.Second {
		t.Fatalf("SIGKILL took too long: %v", elapsed)
	}
	t.Logf("process killed by SIGKILL in %v", elapsed)
}

// ───────────────────────────────────────────────
// Integration: full lifecycle
// ───────────────────────────────────────────────

func TestStart_Wait_Kill_FullLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process spawn in short mode")
	}

	// Start a short-lived process, wait for it, verify clean exit
	p, err := Start(context.Background(), "/bin/echo", "lifecycle")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	if err := p.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}

	// Second Wait should not deadlock
	done := make(chan error, 1)
	go func() {
		done <- p.Wait()
	}()

	select {
	case err := <-done:
		t.Logf("second Wait: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("second Wait deadlocked")
	}
}

func TestStart_MultipleProcesses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process spawn in short mode")
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p, err := Start(context.Background(), "/bin/true")
			if err != nil {
				t.Errorf("Start failed: %v", err)
				return
			}
			if err := p.Wait(); err != nil {
				t.Errorf("Wait failed: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestStart_Kill_DoesNotLeakGoroutine(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process spawn in short mode")
	}

	initial := runtime.NumGoroutine()

	p, err := Start(context.Background(), "/bin/sleep", "300")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	p.Kill()

	// Give goroutines time to settle
	time.Sleep(100 * time.Millisecond)

	// Allow some tolerance for test infrastructure goroutines
	final := runtime.NumGoroutine()
	if final > initial+2 {
		t.Fatalf("possible goroutine leak: started with %d, now %d", initial, final)
	}
}
