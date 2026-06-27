package probe

import (
	"context"
	"fmt"
	"time"

	"bgscan/internal/core/dns"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/portmgr"
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
	processRegistry *ProcessRegistry
	config          SlipstreamConfig
}

// NewSlipstreamProbe creates a new SlipstreamProbe instance.
func NewSlipstreamProbe(workerCount int, config SlipstreamConfig, pm *portmgr.PortManager) (Probe, error) {
	if workerCount <= 0 {
		return nil, fmt.Errorf("worker count must be positive, got %d", workerCount)
	}

	return &SlipstreamProbe{
		pm:              pm,
		processRegistry: NewProcessRegistry(),
		config:          config,
	}, nil
}

func (s *SlipstreamProbe) Init(ctx context.Context) error {
	s.processRegistry.Start(ctx)
	return nil
}

// Run establishes a Slipstream tunnel to the given IP and verifies it
// by issuing an HTTP request through the resulting local SOCKS5 proxy.
func (s *SlipstreamProbe) Run(ctx context.Context, ip string) (*result.IPScanResult, error) {
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

	return &result.IPScanResult{
		IP:      ip,
		Latency: time.Since(start),
	}, nil
}

func (s *SlipstreamProbe) Close() error {
	return nil
}
