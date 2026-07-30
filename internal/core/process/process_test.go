package process

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestStartAndWait(t *testing.T) {
	ctx := context.Background()

	// Use the current test binary as the child process.
	p, err := Start(
		ctx,
		os.Args[0],
		"-test.run=TestHelperProcess",
	)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	err = p.Kill()
	if err != nil {
		t.Fatalf("Kill failed: %v", err)
	}

	if err := p.Wait(); err == nil {
		t.Fatal("expected process exit error")
	}
}

func TestProcessKillWhenNotRunning(t *testing.T) {
	p := &process{}

	err := p.Kill()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProcessStopGracefullyWhenNotRunning(t *testing.T) {
	p := &process{}

	err := p.StopGracefully(time.Second)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFindBinaryInPaths(t *testing.T) {
	dir := t.TempDir()

	name := "testbinary"

	path := filepath.Join(dir, name)

	if err := os.WriteFile(path, nil, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := FindBinaryInPaths(name, []string{dir})
	if err != nil {
		t.Fatalf("FindBinaryInPaths failed: %v", err)
	}

	want, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFindBinaryInPathsNotFound(t *testing.T) {
	_, err := FindBinaryInPaths(
		"does-not-exist",
		nil,
	)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEnsureExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix permission bits are not supported on windows")
	}

	path := filepath.Join(
		t.TempDir(),
		"binary",
	)

	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	err := EnsureExecutable(path)
	if err != nil {
		t.Fatalf("EnsureExecutable failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if info.Mode()&0o111 == 0 {
		t.Fatal("file is not executable")
	}
}

func TestEnsureExecutableAlreadyExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix permission bits are not supported on windows")
	}

	path := filepath.Join(
		t.TempDir(),
		"binary",
	)

	if err := os.WriteFile(path, nil, 0o755); err != nil {
		t.Fatal(err)
	}

	err := EnsureExecutable(path)
	if err != nil {
		t.Fatalf("EnsureExecutable failed: %v", err)
	}
}

func TestEnsureExecutableMissingFile(t *testing.T) {
	err := EnsureExecutable(
		filepath.Join(
			t.TempDir(),
			"missing",
		),
	)

	if err == nil {
		t.Fatal("expected error")
	}
}
