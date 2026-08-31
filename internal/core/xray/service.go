package xray

import (
	"context"
	"fmt"
	"net/netip"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"bgscan/internal/core/fileutil"
	"bgscan/internal/core/process"
)

// XrayService creates, validates, and starts Xray configurations.
type XrayService interface {
	// Binary returns the path of the located Xray executable.
	Binary() string
	// Version returns the Xray version string reported by the binary.
	Version() (string, error)
	// GetOutboundTemplateByName returns the named outbound template.
	GetOutboundTemplateByName(string) (*XrayOutboundsFile, error)
	// GenerateConfig builds a per-target Xray configuration file.
	GenerateConfig(outbound string, ip netip.Addr, port uint16) (string, error)
	// ValidateConfig verifies a configuration file with the located binary.
	ValidateConfig(context.Context, string) error
	// Start runs an Xray process with the given configuration file.
	Start(context.Context, string) (process.Process, error)
}

// xrayService is the default XrayService implementation.
type xrayService struct {
	bin string
}

// NewXrayService creates an Xray service.
//
// It locates the Xray binary before returning; a missing binary is an
// error rather than a per-call lookup later.
func NewXrayService() (XrayService, error) {
	bin, err := FindXrayBinary()
	if err != nil {
		return nil, err
	}

	return &xrayService{bin: bin}, nil
}

// Binary returns the path of the located Xray executable.
func (s *xrayService) Binary() string {
	return s.bin
}

// Version returns the Xray version string reported by the binary.
func (s *xrayService) Version() (string, error) {
	cmd := exec.Command(s.bin, "-version")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("xray version check failed: %w\n%s", err, output)
	}

	return strings.TrimSpace(string(output)), nil
}

func (s *xrayService) GetOutboundTemplateByName(name string) (*XrayOutboundsFile, error) {
	return GetOutboundTemplateByName(name)
}

func (s *xrayService) GenerateConfig(outbound string, ip netip.Addr, port uint16) (string, error) {
	return GenerateConfig(outbound, ip, port)
}

// ValidateConfig verifies a configuration file by executing:
//
//	xray -c <config> --test
//
// If the configuration is invalid, the error contains the full output
// produced by Xray to help diagnose the issue.
func (s *xrayService) ValidateConfig(ctx context.Context, configPath string) error {
	if !fileutil.CheckFileExists(configPath) {
		return fmt.Errorf("config file does not exist: %s", configPath)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.bin, "-c", configPath, "--test")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("xray config validation timed out after 10s (partial output: %s)", string(output))
		}
		return fmt.Errorf("xray config validation failed: %s", string(output))
	}

	return nil
}

// Start runs an Xray process with the given configuration file.
//
// The provided context controls the lifetime of the process: if the context
// is canceled, the Xray process is terminated automatically.
func (s *xrayService) Start(ctx context.Context, configPath string) (process.Process, error) {
	if !fileutil.CheckFileExists(configPath) {
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}

	return process.Start(ctx, s.bin, "-c", configPath)
}

// FindXrayBinary attempts to locate the Xray executable.
func FindXrayBinary() (string, error) {
	return process.FindBinaryInPaths("xray", getXrayPaths())
}

func getXrayPaths() []string {
	base, err := fileutil.BasePath()
	if err != nil {
		return []string{"assets/xray", "xray", ""}
	}

	return []string{
		filepath.Join(base, "assets", "xray"),
		filepath.Join(base, "xray"),
		base,
	}
}
