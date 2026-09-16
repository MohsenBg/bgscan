package xrayprobe

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/netip"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner/portmgr"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/core/speedtest"
	"github.com/MohsenBg/bgscan/internal/core/xray"
	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/xtls/xray-core/shared"
)

const bytesPerKbpsSecond float64 = 1000.0 / 8.0

// XrayService is the probe's view of xray.XrayService.
type XrayService interface {
	GetOutboundTemplateByName(string) (*xray.XrayOutboundsFile, error)
	GenerateConfig(outbound string, ip netip.Addr, port uint16) (*xray.XrayConfig, error)
	ValidateConfig(context.Context, *xray.XrayConfig) error
	Start(context.Context, *xray.XrayConfig) (xray.Instance, error)
	CloseShared() error
}

// XrayProbe validates connectivity and performance through a temporary local
// Xray SOCKS proxy configured for a target IP.
type XrayProbe struct {
	pm    portmgr.Manager
	xray  XrayService
	speed speedtest.Service

	outbound        string
	latencyTimeout  time.Duration
	transferTimeout time.Duration
	testMode        config.ConnectivityTest
	downloadBytes   int64
	uploadBytes     int64
	minDownload     speedtest.BitsPerSec
	minUpload       speedtest.BitsPerSec
}

// Option configures an XrayProbe.
type Option func(*XrayProbe)

// WithXrayService uses service for Xray configuration and instance operations.
func WithXrayService(service XrayService) Option {
	return func(p *XrayProbe) {
		if service != nil {
			p.xray = service
		}
	}
}

// WithSpeedTester uses tester for proxy latency and speed measurements.
func WithSpeedTester(tester speedtest.Service) Option {
	return func(p *XrayProbe) {
		if tester != nil {
			p.speed = tester
		}
	}
}

// NewXrayProbe creates an Xray probe for outboundName.
func NewXrayProbe(
	cfg *config.XrayConfig,
	outboundName string,
	pm portmgr.Manager,
	opts ...Option,
) (probe.Probe, error) {
	if cfg == nil {
		return nil, fmt.Errorf("xray config is nil")
	}

	if pm == nil {
		return nil, fmt.Errorf("port manager is nil")
	}

	if outboundName == "" {
		return nil, fmt.Errorf("outbound template name is empty")
	}

	timeout := cfg.Timeout.Duration()
	if timeout <= 0 {
		return nil, fmt.Errorf("xray timeout must be positive")
	}

	speedTestTimeout := cfg.SpeedTestTimeout.Duration()
	if speedTestTimeout <= 0 {
		return nil, fmt.Errorf("xray speed test timeout must be positive")
	}

	byteFactor := bytesPerKbpsSecond * speedTestTimeout.Seconds()
	downloadBytes := math.Round(float64(cfg.DownloadSpeed) * byteFactor)
	uploadBytes := math.Round(float64(cfg.UploadSpeed) * byteFactor)

	p := &XrayProbe{
		pm:    pm,
		speed: speedtest.NewService(),

		outbound:        outboundName,
		latencyTimeout:  timeout,
		transferTimeout: speedTestTimeout,
		testMode:        cfg.ConnectivityTestType,
		downloadBytes:   int64(downloadBytes),
		uploadBytes:     int64(uploadBytes),
		minDownload:     speedtest.BitsPerSec(cfg.DownloadSpeed) * speedtest.Kbps,
		minUpload:       speedtest.BitsPerSec(cfg.UploadSpeed) * speedtest.Kbps,
	}

	for _, opt := range opts {
		opt(p)
	}

	if p.xray == nil {
		// Scanner tuning, applied once process-wide at creation. The
		// connect cap follows the latency timeout from settings, so a
		// user retune can never strand dials beyond the check budget.
		p.xray = xray.NewXrayService(xray.WithScanTuning(shared.ScanTuning{
			DialTimeout:        timeout,
			MaxConcurrentDials: cfg.MaxConcurrentDials,
			DialMaxAttempts:    cfg.DialMaxAttempts,
		}))
	}

	if _, err := p.xray.GetOutboundTemplateByName(outboundName); err != nil {
		return nil, fmt.Errorf("unknown outbound template %q: %w", outboundName, err)
	}

	return p, nil
}

// Schema returns the result schema emitted by the probe.
func (p *XrayProbe) Schema() result.ResultSchema {
	return Schema
}

// Init validates the outbound template once; per-IP configs only swap the
// address, so Run skips per-target validation.
func (p *XrayProbe) Init(ctx context.Context) error {
	port, err := p.pm.Get(ctx)
	if err != nil {
		return fmt.Errorf("lease port for Xray config validation: %w", err)
	}
	p.pm.Release(port)

	loopback := netip.AddrFrom4([4]byte{127, 0, 0, 1})

	cfg, err := p.xray.GenerateConfig(p.outbound, loopback, port)
	if err != nil {
		return fmt.Errorf("generate Xray config for validation: %w", err)
	}

	if err := p.xray.ValidateConfig(ctx, cfg); err != nil {
		return fmt.Errorf("outbound template %q is invalid: %w", p.outbound, err)
	}

	return nil
}

// Run starts a temporary Xray instance for ip and performs the configured
// connectivity, download, and upload checks.
func (p *XrayProbe) Run(ctx context.Context, ip netip.Addr) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	port, err := p.pm.Get(ctx)
	if err != nil {
		return nil, err
	}
	defer p.pm.Release(port)

	cfg, err := p.xray.GenerateConfig(p.outbound, ip, port)
	if err != nil {
		return nil, fmt.Errorf("generate Xray config: %w", err)
	}

	inst, err := p.xray.Start(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("start Xray: %w", err)
	}

	// Always closed via defer below.
	defer func() {
		if err := inst.Close(); err != nil {
			logger.CoreError("terminate Xray: %v", err)
		}
	}()

	addr := net.JoinHostPort("127.0.0.1", fmt.Sprint(port))

	if err := p.pm.WaitOpen(ctx, addr, time.Second); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}

		return nil, fmt.Errorf("wait for Xray proxy: %w", err)
	}

	latency, err := p.speed.MeasureLatency(ctx, speedtest.LatencyConfig{
		Timeout:   p.latencyTimeout,
		ProxyPort: port,
	})
	if err != nil {
		return nil, fmt.Errorf("measure latency for %s: %w", ip, err)
	}

	result := XrayResult{
		IP:      ip,
		Latency: latency.RTT,
	}

	switch p.testMode {
	case config.ConnectivityOnly:
		return result, nil

	case config.DownloadSpeedOnly:
		download, err := p.measureDownload(ctx, port)
		if err != nil {
			return nil, fmt.Errorf("measure download for %s: %w", ip, err)
		}

		result.Download = download.Speed

	case config.UploadSpeedOnly:
		upload, err := p.measureUpload(ctx, port)
		if err != nil {
			return nil, fmt.Errorf("measure upload for %s: %w", ip, err)
		}

		result.Upload = upload.Speed

	case config.Both:
		download, err := p.measureDownload(ctx, port)
		if err != nil {
			return nil, fmt.Errorf("measure download for %s: %w", ip, err)
		}
		result.Download = download.Speed

		upload, err := p.measureUpload(ctx, port)
		if err != nil {
			return nil, fmt.Errorf("measure upload for %s: %w", ip, err)
		}
		result.Upload = upload.Speed
	}

	return result, nil
}

func (p *XrayProbe) measureDownload(
	ctx context.Context,
	port uint16,
) (speedtest.SpeedResult, error) {
	return p.speed.MeasureDownloadSpeed(ctx, speedtest.DownloadConfig{
		Bytes:     p.downloadBytes,
		Timeout:   p.transferTimeout,
		MinSpeed:  p.minDownload,
		ProxyPort: port,
	})
}

func (p *XrayProbe) measureUpload(
	ctx context.Context,
	port uint16,
) (speedtest.SpeedResult, error) {
	return p.speed.MeasureUploadSpeed(ctx, speedtest.UploadConfig{
		Bytes:     p.uploadBytes,
		Timeout:   p.transferTimeout,
		MinSpeed:  p.minUpload,
		ProxyPort: port,
	})
}

// Close releases process-global transport state shared by all Xray
// instances (pooled dialers, stuck dials). Run's defer inst.Close() per IP
// is not enough: globals outlive every instance.
// Call once after a scan, when no concurrent Run is in flight.
func (p *XrayProbe) Close() error {
	if p.xray != nil {
		return p.xray.CloseShared()
	}
	return nil
}
