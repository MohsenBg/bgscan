package fileutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// helperProcess markers for the subprocess variants of the BasePath tests.
const (
	basePathHelperEnv       = "BGSCAN_BASEPATH_HELPER"
	basePathHelperReal      = "real"
	basePathHelperSymlinked = "symlinked"
)

// initialWorkingDir captures the working directory before any test can
// chdir into a temp dir that later gets removed.
var initialWorkingDir = func() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return wd
}()

// runHelperBinary copies the test binary out of the Go build temp
// directory so it behaves like a real built executable, runs it as a
// subprocess in helper mode, and returns the reported base path.
func runHelperBinary(t *testing.T, mode string) (string, string) {
	t.Helper()

	exe, err := os.Executable()
	if err != nil {
		t.Skipf("cannot determine test binary: %v", err)
	}

	dir := t.TempDir()
	bin := filepath.Join(dir, "bgscan-testapp")

	data, err := os.ReadFile(exe)
	if err != nil {
		t.Fatalf("read test binary: %v", err)
	}
	if err := os.WriteFile(bin, data, 0o755); err != nil {
		t.Fatalf("write copy of test binary: %v", err)
	}

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), basePathHelperEnv+"="+mode)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("helper subprocess failed: %v", err)
	}
	return string(out), dir
}

// TestBasePath_UnderGoTest verifies the development rule: when the test
// binary is a temporary `go test` build, BasePath falls back to the
// working directory so application-relative resources resolve from the
// project tree.
func TestBasePath_UnderGoTest(t *testing.T) {
	wd := initialWorkingDir
	if wd == "" {
		t.Skip("working directory unavailable")
	}

	got, err := BasePath()
	if err != nil {
		t.Fatalf("BasePath() error = %v", err)
	}

	resolvedWD, err := filepath.EvalSymlinks(wd)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", wd, err)
	}

	if got != resolvedWD && got != wd {
		t.Fatalf("BasePath() = %q, want working directory %q", got, wd)
	}
}

// TestBasePath_RealBinary verifies that for a real built executable,
// BasePath returns the directory containing the binary — even when the
// process is started from a different working directory.
func TestBasePath_RealBinary(t *testing.T) {
	t.Chdir(t.TempDir())

	got, binDir := runHelperBinary(t, "real")

	resolvedBinDir, err := filepath.EvalSymlinks(binDir)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", binDir, err)
	}

	if got != resolvedBinDir && got != binDir {
		t.Fatalf("BasePath() = %q, want executable directory %q", got, binDir)
	}
}

// TestBasePath_SymlinkedBinary verifies that a symlink pointing at the
// executable resolves to the real binary's directory before deriving the
// base path.
func TestBasePath_SymlinkedBinary(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Skipf("cannot determine test binary: %v", err)
	}

	realDir := t.TempDir()
	bin := filepath.Join(realDir, "bgscan-testapp")
	data, err := os.ReadFile(exe)
	if err != nil {
		t.Fatalf("read test binary: %v", err)
	}
	if err := os.WriteFile(bin, data, 0o755); err != nil {
		t.Fatalf("write test binary: %v", err)
	}

	linkDir := t.TempDir()
	link := filepath.Join(linkDir, "bgscan-link")
	if err := os.Symlink(bin, link); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}

	cmd := exec.Command(link)
	cmd.Env = append(os.Environ(), basePathHelperEnv+"=symlinked")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("helper subprocess failed: %v", err)
	}
	got := string(out)

	resolvedRealDir, err := filepath.EvalSymlinks(realDir)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", realDir, err)
	}
	resolvedLinkDir, err := filepath.EvalSymlinks(linkDir)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", linkDir, err)
	}

	if got != resolvedRealDir && got != realDir {
		t.Fatalf(
			"BasePath() = %q, want resolved binary directory %q (not symlink dir %q)",
			got, resolvedRealDir, resolvedLinkDir,
		)
	}
}
