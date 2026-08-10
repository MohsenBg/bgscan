package dns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	"bgscan/internal/core/process"
)

var ErrSlipstreamTunnelNotRunning = errors.New("slipstream-client process is not running")

// SlipstreamClient manages a slipstream-client tunnel process.
type SlipstreamClient interface {
	RunTunnel(context.Context, string, uint16) (process.Process, error)
}

type slipstreamClient struct {
	bin      string
	domain   string
	certPath string
	dnsPort  uint16

	start ProcessStarter
}

// SlipstreamClientOption configures a Slipstream client.
type SlipstreamClientOption func(*slipstreamClient)

// WithSlipstreamProcessStarter replaces the process launcher.
//
// It allows tests to run without starting a real slipstream-client binary.
func WithSlipstreamProcessStarter(start ProcessStarter) SlipstreamClientOption {
	return func(client *slipstreamClient) {
		if start != nil {
			client.start = start
		}
	}
}

// WithSlipstreamClientBinary uses bin as the slipstream-client executable
// instead of searching the known locations.
func WithSlipstreamClientBinary(bin string) SlipstreamClientOption {
	return func(client *slipstreamClient) {
		if bin != "" {
			client.bin = bin
		}
	}
}

// getSlipstreamPaths returns the locations searched for slipstream-client.
func getSlipstreamPaths() []string {
	exe, err := os.Executable()
	if err != nil {
		return nil
	}

	base := filepath.Dir(exe)

	return []string{
		filepath.Join(base, "assets", "slipstream-client"),
		filepath.Join(base, "assets", "slipstream", "slipstream-client"),
		filepath.Join(base, "slipstream-client"),
		base,
	}
}

// NewSlipstreamClient creates a client for a Slipstream DNS tunnel.
//
// Unless WithSlipstreamClientBinary is provided, it locates slipstream-client
// before returning the client.
func NewSlipstreamClient(
	domain string,
	dnsPort uint16,
	certPath string,
	opts ...SlipstreamClientOption,
) (SlipstreamClient, error) {
	client := &slipstreamClient{
		domain:   domain,
		dnsPort:  dnsPort,
		certPath: certPath,
		start:    process.Start,
	}

	for _, opt := range opts {
		opt(client)
	}

	if client.bin == "" {
		bin, err := FindSlipstreamClient()
		if err != nil {
			return nil, err
		}

		client.bin = bin
	}

	return client, nil
}

// FindSlipstreamClient locates slipstream-client in known locations or PATH.
func FindSlipstreamClient() (string, error) {
	return process.FindBinaryInPaths("slipstream-client", getSlipstreamPaths())
}

// RunTunnel starts a Slipstream DNS tunnel and listens on listenPort.
func (s *slipstreamClient) RunTunnel(
	ctx context.Context,
	ip string,
	listenPort uint16,
) (process.Process, error) {
	args := []string{
		"-d", s.domain,
		"-r", net.JoinHostPort(ip, fmt.Sprint(s.dnsPort)),
		"-l", fmt.Sprint(listenPort),
	}

	if s.certPath != "" {
		args = append(args, "--cert", s.certPath)
	}

	proc, err := s.start(ctx, s.bin, args...)
	if err != nil {
		return nil, err
	}

	return proc, nil
}

// VerifySlipstreamClient verifies that slipstream-client can execute.
func VerifySlipstreamClient() error {
	path, err := FindSlipstreamClient()
	if err != nil {
		return fmt.Errorf("find slipstream-client: %w", err)
	}

	if err := exec.Command(path, "--help").Run(); err != nil {
		return fmt.Errorf("run slipstream-client: %w", err)
	}

	return nil
}
