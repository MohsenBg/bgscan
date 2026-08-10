package xray

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"bgscan/internal/core/fileutil"
	"bgscan/internal/core/process"
)

// FindXrayBinary attempts to locate the Xray executable.
func FindXrayBinary() (string, error) {
	return process.FindBinaryInPaths("xray", getXrayPaths())
}

func XrayVersion() (string, error) {
	xrayBin, err := FindXrayBinary()
	if err != nil {
		return "", fmt.Errorf("xray binary not found: %w", err)
	}

	cmd := exec.Command(xrayBin, "-version")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("xray version check failed: %w\n%s", err, output)
	}

	version := strings.TrimSpace(string(output))
	return version, nil
}

// ValidateConfig verifies that a configuration file is valid
// by executing:
//
//	xray -c <config> --test
//
// If the configuration is invalid, the error returned will contain
// the full output produced by Xray to help diagnose the issue.
func ValidateConfig(ctx context.Context, configPath string) error {
	if !fileutil.CheckFileExists(configPath) {
		return fmt.Errorf("config file does not exist: %s", configPath)
	}

	xrayBin, err := FindXrayBinary()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, xrayBin, "-c", configPath, "--test")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("xray config validation timed out after 10s (partial output: %s)", string(output))
		}
		return fmt.Errorf("xray config validation failed: %s", string(output))
	}

	return nil
}

// StartXray launches an Xray process using the provided configuration.
//
// The process is started asynchronously and returned as an XrayProcess
// instance so the caller can manage its lifecycle.
//
// The provided context controls the lifetime of the process. If the
// context is canceled, the Xray process will be terminated automatically.
func StartXray(ctx context.Context, configPath string) (process.Process, error) {
	if !fileutil.CheckFileExists(configPath) {
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}

	xrayBin, err := FindXrayBinary()
	if err != nil {
		return nil, err
	}

	return process.Start(ctx, xrayBin, "-c", configPath)
}

func getXrayPaths() []string {
	exe, err := os.Executable()
	if err != nil {
		return []string{"assets/xray", "xray", ""}
	}

	base := filepath.Dir(exe)

	return []string{
		filepath.Join(base, "assets", "xray"),
		filepath.Join(base, "xray"),
		base,
	}
}
