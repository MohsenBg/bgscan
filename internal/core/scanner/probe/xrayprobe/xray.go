package xrayprobe

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"bgscan/internal/core/config"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/portmgr"
	"bgscan/internal/core/scanner/probe"
	"bgscan/internal/core/speedtest"
	"bgscan/internal/core/xray"
	"bgscan/internal/logger"
)

// XrayProbe evaluates an IP by spawning a temporary Xray outbound on a local
// port and measuring latency and (optionally) download/upload throughput
// through the resulting proxy. Each Run allocates a port, generates a config,
// validates it, starts Xray, and tears down every resource on return.
type XrayProbe struct {
	pm              *portmgr.PortManager
	processRegistry *probe.ProcessRegistry
	outbound        string
	latencyTimeout  time.Duration
	transferTimeout time.Duration
	testMode        config.ConnectivityTest
	downloadBytes   int64
	uploadBytes     int64
	minDownload     speedtest.BitsPerSec
	minUpload       speedtest.BitsPerSec
	maxLatency      time.Duration
}

// NewXrayProbe builds an XrayProbe. outboundName must be a known outbound
// template name. The configured download/upload kbps limits are converted to
// byte budgets and per-direction min-speed thresholds for the speed test.
func NewXrayProbe(cfg *config.XrayConfig, outboundName string, pm *portmgr.PortManager) (probe.Probe, error) {
	if _, err := xray.GetOutboundTemplateByName(outboundName); err != nil {
		return nil, fmt.Errorf("unknown outbound template %q: %w", outboundName, err)
	}

	speedTestSeconds := int64(cfg.Timeout.Duration().Seconds())

	// cfg.DownloadSpeed and cfg.UploadSpeed are in kbps.
	// Bytes = kbps * 1000 bits/kbit / 8 bits/byte * seconds
	downloadBytes := int64(cfg.DownloadSpeed) * 1000 / 8 * speedTestSeconds
	uploadBytes := int64(cfg.UploadSpeed) * 1000 / 8 * speedTestSeconds

	return &XrayProbe{
		outbound:        outboundName,
		pm:              pm,
		processRegistry: probe.NewProcessRegistry(),
		latencyTimeout:  cfg.Timeout.Duration(),
		transferTimeout: cfg.Timeout.Duration(),
		testMode:        cfg.ConnectivityTestType,
		downloadBytes:   downloadBytes,
		uploadBytes:     uploadBytes,
		minDownload:     speedtest.BitsPerSec(cfg.DownloadSpeed) * speedtest.Kbps,
		minUpload:       speedtest.BitsPerSec(cfg.UploadSpeed) * speedtest.Kbps,
	}, nil
}

// Schema returns the Xray result schema.
func (p *XrayProbe) Schema() result.ResultSchema { return Schema }

// Init starts the process registry used to track spawned Xray processes.
func (p *XrayProbe) Init(ctx context.Context) error {
	p.processRegistry.Start(ctx)
	return nil
}

// Run probes ip by starting a temporary Xray instance on an allocated local
// port, waiting for the proxy to open, then running latency and (per
// testMode) download/upload speed tests through it. On failure the error is
// wrapped with the phase that failed.
func (p *XrayProbe) Run(ctx context.Context, ip string) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	port, err := p.pm.GetPort(ctx)
	if err != nil {
		return nil, err
	}
	defer p.pm.ReleasePort(port)

	configPath, err := xray.GenerateConfig(p.outbound, ip, port)
	if err != nil {
		return nil, fmt.Errorf("xray config generation failed: %w", err)
	}
	defer func() {
		if err := os.Remove(configPath); err != nil {
			logger.CoreError("failed to remove xray config file: %v", err)
		}
	}()

	if err := xray.ValidateConfig(configPath); err != nil {
		return nil, fmt.Errorf("invalid xray config: %w", err)
	}

	proc, err := xray.StartXray(ctx, configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to start xray: %w", err)
	}

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

	addr := net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port))
	if err := portmgr.WaitPortOpen(ctx, addr, time.Second); err != nil {
		return nil, fmt.Errorf("proxy port did not open for %s: %w", ip, err)
	}

	latResult, err := speedtest.MeasureLatency(ctx, speedtest.LatencyConfig{
		Timeout:    p.latencyTimeout,
		MaxLatency: p.maxLatency,
		ProxyPort:  port,
	})
	if err != nil {
		return nil, fmt.Errorf("latency measurement failed for %s: %w", ip, err)
	}

	res := XrayResult{
		IP:      ip,
		Latency: latResult.RTT,
	}

	switch p.testMode {

	case config.ConnectivityOnly:
		return res, nil

	case config.DownloadSpeedOnly:
		dlResult, err := speedtest.MeasureDownloadSpeed(ctx, speedtest.DownloadConfig{
			Bytes:     p.downloadBytes,
			Timeout:   p.transferTimeout,
			MinSpeed:  p.minDownload,
			ProxyPort: port,
		})
		if err != nil {
			return nil, fmt.Errorf("download test failed for %s: %w", ip, err)
		}
		res.Download = dlResult.Speed

	case config.UploadSpeedOnly:
		ulResult, err := speedtest.MeasureUploadSpeed(ctx, speedtest.UploadConfig{
			Bytes:     p.uploadBytes,
			Timeout:   p.transferTimeout,
			MinSpeed:  p.minUpload,
			ProxyPort: port,
		})
		if err != nil {
			return nil, fmt.Errorf("upload test failed for %s: %w", ip, err)
		}
		res.Upload = ulResult.Speed

	case config.Both:
		dlResult, err := speedtest.MeasureDownloadSpeed(ctx, speedtest.DownloadConfig{
			Bytes:     p.downloadBytes,
			Timeout:   p.transferTimeout,
			MinSpeed:  p.minDownload,
			ProxyPort: port,
		})
		if err != nil {
			return nil, fmt.Errorf("download test failed for %s: %w", ip, err)
		}
		res.Download = dlResult.Speed

		ulResult, err := speedtest.MeasureUploadSpeed(ctx, speedtest.UploadConfig{
			Bytes:     p.uploadBytes,
			Timeout:   p.transferTimeout,
			MinSpeed:  p.minUpload,
			ProxyPort: port,
		})
		if err != nil {
			return nil, fmt.Errorf("upload test failed for %s: %w", ip, err)
		}
		res.Upload = ulResult.Speed
	}

	return res, nil
}

// Close is a no-op; per-Run teardown handles resource cleanup.
func (p *XrayProbe) Close() error {
	return nil
}
