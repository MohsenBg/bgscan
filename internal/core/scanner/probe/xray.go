package probe

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"bgscan/internal/core/config"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/portmgr"
	"bgscan/internal/core/xray"
	"bgscan/internal/logger"
)

// XrayProbe evaluates an IP address by launching a temporary Xray instance.
//
// The probe dynamically generates a runtime Xray configuration targeting the
// provided IP, starts the Xray process, and performs connectivity or bandwidth
// measurements through the local SOCKS5 proxy exposed by that instance.
//
// The probe workflow includes:
//
//  1. Reserving a free local port via PortManager
//  2. Generating a temporary Xray configuration for the target IP
//  3. Starting the Xray process
//  4. Waiting for the local SOCKS5 proxy to become ready
//  5. Measuring connection latency
//  6. Optionally performing download and/or upload bandwidth tests
//
// All operations are context-aware. If the provided context is cancelled,
// the probe stops immediately and any spawned Xray process is terminated.
//
// XrayProbe maintains a ProcessRegistry to track spawned subprocesses and
// ensure they are cleaned up reliably.
type XrayProbe struct {
	// pm manages allocation and release of local ports used by
	// temporary Xray SOCKS5 proxies.
	pm *portmgr.PortManager

	// processRegistry tracks active Xray subprocesses and ensures
	// they are terminated during shutdown.
	processRegistry *ProcessRegistry

	// outbound is the name of the outbound template used when
	// generating runtime Xray configurations.
	outbound string

	// timeout defines the maximum duration for latency and
	// bandwidth measurements.
	timeout time.Duration

	// testMode determines which connectivity tests should run
	// (latency only, download, upload, or both).
	testMode config.ConnectivityTest

	// downloadBytes defines the payload size used for download tests.
	downloadBytes int64

	// uploadBytes defines the payload size used for upload tests.
	uploadBytes int64
}

// NewXrayProbe constructs a new XrayProbe instance.
//
// The outboundName must correspond to a valid outbound template
// registered in the Xray configuration subsystem. If the template
// cannot be found, an error is returned.
func NewXrayProbe(cfg *config.XrayConfig, outboundName string, pm *portmgr.PortManager) (Probe, error) {
	if _, err := xray.GetOutboundTemplateByName(outboundName); err != nil {
		return nil, fmt.Errorf("unknown outbound template %q: %w", outboundName, err)
	}

	return &XrayProbe{
		outbound:        outboundName,
		pm:              pm,
		processRegistry: NewProcessRegistry(),
		timeout:         cfg.Timeout.Duration(),
		testMode:        cfg.ConnectivityTestType,
		downloadBytes:   int64(cfg.DownloadSpeed) * 1024,
		uploadBytes:     int64(cfg.UploadSpeed) * 1024,
	}, nil
}

// Init implements [Probe] and initializes internal background components.
//
// Currently, this starts the internal ProcessRegistry monitor which tracks
// and supervises Xray subprocesses launched by the probe.
func (p *XrayProbe) Init(ctx context.Context) error {
	p.processRegistry.Start(ctx)
	return nil
}

// Run executes a full probe cycle against the provided IP address.
//
// The procedure:
//
//  1. Validate context for early cancellation
//  2. Reserve a local port via PortManager
//  3. Generate a temporary Xray configuration targeting the IP
//  4. Validate the configuration before execution
//  5. Launch the Xray process
//  6. Wait for the local SOCKS5 proxy to become available
//  7. Measure connection latency
//  8. Optionally perform download and/or upload tests
//
// The function guarantees that:
//
//   - The reserved port is always released.
//   - Temporary config files are removed.
//   - The Xray subprocess is terminated before returning.
//
// If the context is cancelled at any point, the probe terminates quickly
// and all associated resources are cleaned up.
func (p *XrayProbe) Run(ctx context.Context, ip string) (*result.IPScanResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Reserve a local port for the temporary SOCKS5 proxy.
	port, err := p.pm.GetPort(ctx)
	if err != nil {
		return nil, err
	}
	defer p.pm.ReleasePort(port)

	// Generate runtime Xray configuration.
	configPath, err := xray.GenerateConfig(p.outbound, ip, port)
	if err != nil {
		return nil, fmt.Errorf("xray config generation failed: %w", err)
	}
	defer func() {
		if err := os.Remove(configPath); err != nil {
			logger.CoreError("failed to remove xray config file: %v", err)
		}
	}()

	// Validate configuration before starting the process.
	if err := xray.ValidateConfig(configPath); err != nil {
		return nil, fmt.Errorf("invalid xray config: %w", err)
	}

	// Start the Xray process.
	proc, err := xray.StartXray(ctx, configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to start xray: %w", err)
	}

	// Track the process for lifecycle management.
	id, err := p.processRegistry.Register(ctx, proc)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := proc.Kill(); err != nil {
			logger.CoreError("failed to terminate xray: %v", err)
		}
		if err := p.processRegistry.Unregister(ctx, id); err != nil {
			logger.CoreError("failed to unregister xray process: %v", err)
		}
	}()

	// Wait until the local SOCKS5 proxy becomes reachable.
	addr := net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port))
	if err := portmgr.WaitPortOpen(ctx, addr, time.Second); err != nil {
		return nil, fmt.Errorf("proxy port did not open for %s: %w", ip, err)
	}

	// Measure proxy latency through Xray.
	latency, err := xray.MeasureLatency(ctx, p.timeout, port)
	if err != nil {
		return nil, fmt.Errorf("latency measurement failed for %s: %w", ip, err)
	}

	res := &result.IPScanResult{
		IP:      ip,
		Latency: latency,
	}

	// Run additional connectivity tests depending on the selected mode.
	switch p.testMode {

	case config.ConnectivityOnly:
		return res, nil

	case config.DownloadSpeedOnly:
		res.Download, err = xray.MeasureDownloadDuration(ctx, p.timeout, p.downloadBytes, port)
		if err != nil {
			return nil, fmt.Errorf("download test failed for %s: %w", ip, err)
		}

	case config.UploadSpeedOnly:
		res.Upload, err = xray.MeasureUploadDuration(ctx, p.timeout, p.uploadBytes, port)
		if err != nil {
			return nil, fmt.Errorf("upload test failed for %s: %w", ip, err)
		}

	case config.Both:
		res.Download, err = xray.MeasureDownloadDuration(ctx, p.timeout, p.downloadBytes, port)
		if err != nil {
			return nil, fmt.Errorf("download test failed for %s: %w", ip, err)
		}

		res.Upload, err = xray.MeasureUploadDuration(ctx, p.timeout, p.uploadBytes, port)
		if err != nil {
			return nil, fmt.Errorf("upload test failed for %s: %w", ip, err)
		}
	}

	return res, nil
}

// Close implements the Probe interface.
//
// Close terminates any active Xray processes that were launched by
// previous probe runs through the internal ProcessRegistry.
//
// The PortManager instance is not affected.
func (p *XrayProbe) Close() error {
	return nil
}
