package slipstreamprobe

import (
	"context"
	"fmt"
	"time"

	"bgscan/internal/core/dns"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/portmgr"
	"bgscan/internal/core/scanner/probe"
	"bgscan/internal/logger"
)

// SlipstreamConfig holds parameters required to establish and verify
// a Slipstream DNS tunnel.
type SlipstreamConfig struct {
	Domain   string
	CertPath string
	DNSPort  uint16
	Timeout  time.Duration
}

// SlipstreamProbe performs connectivity verification by creating a
// Slipstream DNS tunnel toward a target IP and testing the resulting
// local SOCKS5 proxy.
type SlipstreamProbe struct {
	pm              *portmgr.PortManager
	processRegistry *probe.ProcessRegistry
	config          SlipstreamConfig
}

// NewSlipstreamProbe creates a new SlipstreamProbe instance, validating
// that workerCount is positive.
func NewSlipstreamProbe(workerCount int, config SlipstreamConfig, pm *portmgr.PortManager) (probe.Probe, error) {
	if workerCount <= 0 {
		return nil, fmt.Errorf("worker count must be positive, got %d", workerCount)
	}

	return &SlipstreamProbe{
		pm:              pm,
		processRegistry: probe.NewProcessRegistry(),
		config:          config,
	}, nil
}

// Schema returns the result schema for Slipstream probes.
func (s *SlipstreamProbe) Schema() result.ResultSchema {
	return Schema
}

// Init implements [probe.Probe] and starts the internal ProcessRegistry
// used to track and manage Slipstream tunnel processes.
func (s *SlipstreamProbe) Init(ctx context.Context) error {
	s.processRegistry.Start(ctx)
	return nil
}

// Run implements [probe.Probe] and establishes a Slipstream tunnel toward
// the target IP, verifying connectivity via the resulting local SOCKS5 proxy.
//
// On success, it returns a SlipstreamResult containing the target IP and
// post-tunnel validation latency.
func (s *SlipstreamProbe) Run(ctx context.Context, ip string) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	localPort, err := s.pm.GetPort(ctx)
	if err != nil {
		return nil, err
	}
	defer s.pm.ReleasePort(localPort)

	client, err := dns.NewSlipstreamClient(
		s.config.Domain,
		s.config.DNSPort,
		s.config.CertPath,
	)
	if err != nil {
		return nil, fmt.Errorf("slipstream client init failed: %w", err)
	}

	proc, err := client.RunTunnel(ctx, ip, localPort)
	if err != nil {
		return nil, fmt.Errorf("failed to start slipstream tunnel: %w", err)
	}

	id, err := s.processRegistry.Register(ctx, proc)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := s.processRegistry.Unregister(ctx, id); err != nil {
			logger.CoreError("error unregistering process: %v", err)
		}
		_ = client.StopTunnel(context.Background())
	}()

	localProxyAddr := fmt.Sprintf("127.0.0.1:%d", localPort)

	if err := portmgr.WaitPortOpen(ctx, localProxyAddr, time.Second); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("proxy port did not open for %s: %w", ip, err)
	}

	start := time.Now()

	ok := dns.TestProxy(ctx, localProxyAddr, s.config.Timeout)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if !ok {
		return nil, fmt.Errorf("slipstream handshake failed for %s", ip)
	}

	return SlipstreamResult{
		IP:      ip,
		Latency: time.Since(start),
		Port:    localPort,
	}, nil
}

// Close implements [probe.Probe]. It is a no-op as SlipstreamProbe holds
// no long-lived resources outside of the per-Run scope.
func (s *SlipstreamProbe) Close() error {
	return nil
}
