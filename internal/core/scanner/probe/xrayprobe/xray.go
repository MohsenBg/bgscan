package xrayprobe

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"os"
	"time"

	"bgscan/internal/core/config"
	"bgscan/internal/core/process"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/portmgr"
	"bgscan/internal/core/scanner/probe"
	"bgscan/internal/core/speedtest"
	"bgscan/internal/core/xray"
	"bgscan/internal/logger"
)

// XrayService creates, validates, and starts Xray configurations.
type XrayService interface {
	GetOutboundTemplateByName(string) (*xray.XrayOutboundsFile, error)
	GenerateConfig(outbound, ip string, port uint16) (string, error)
	ValidateConfig(context.Context, string) error
	Start(context.Context, string) (process.Process, error)
}

// SpeedTester measures latency and transfer speeds through a proxy.
type SpeedTester interface {
	MeasureLatency(context.Context, speedtest.LatencyConfig) (speedtest.LatencyResult, error)
	MeasureDownloadSpeed(context.Context, speedtest.DownloadConfig) (speedtest.SpeedResult, error)
	MeasureUploadSpeed(context.Context, speedtest.UploadConfig) (speedtest.SpeedResult, error)
}

type realXrayService struct{}

func (realXrayService) GetOutboundTemplateByName(name string) (*xray.XrayOutboundsFile, error) {
	return xray.GetOutboundTemplateByName(name)
}

func (realXrayService) GenerateConfig(outbound, ip string, port uint16) (string, error) {
	return xray.GenerateConfig(outbound, ip, port)
}

func (realXrayService) ValidateConfig(ctx context.Context, path string) error {
	return xray.ValidateConfig(ctx, path)
}

func (realXrayService) Start(ctx context.Context, path string) (process.Process, error) {
	return xray.StartXray(ctx, path)
}

type realSpeedTester struct{}

func (realSpeedTester) MeasureLatency(
	ctx context.Context,
	cfg speedtest.LatencyConfig,
) (speedtest.LatencyResult, error) {
	return speedtest.MeasureLatency(ctx, cfg)
}

func (realSpeedTester) MeasureDownloadSpeed(
	ctx context.Context,
	cfg speedtest.DownloadConfig,
) (speedtest.SpeedResult, error) {
	return speedtest.MeasureDownloadSpeed(ctx, cfg)
}

func (realSpeedTester) MeasureUploadSpeed(
	ctx context.Context,
	cfg speedtest.UploadConfig,
) (speedtest.SpeedResult, error) {
	return speedtest.MeasureUploadSpeed(ctx, cfg)
}

// XrayProbe validates connectivity and performance through a temporary local
// Xray SOCKS proxy configured for a target IP.
type XrayProbe struct {
	pm             portmgr.Manager
	processTracker probe.ProcessTracker
	xray           XrayService
	speed          SpeedTester

	outbound        string
	latencyTimeout  time.Duration
	transferTimeout time.Duration
	testMode        config.ConnectivityTest
	downloadBytes   int64
	uploadBytes     int64
	minDownload     speedtest.BitsPerSec
	minUpload       speedtest.BitsPerSec
	maxLatency      time.Duration

	remove   func(string) error
	waitOpen func(context.Context, string, time.Duration) error
}

// Option configures an XrayProbe.
type Option func(*XrayProbe)

// WithProcessTracker uses tracker to manage started Xray processes.
func WithProcessTracker(tracker probe.ProcessTracker) Option {
	return func(p *XrayProbe) {
		if tracker != nil {
			p.processTracker = tracker
		}
	}
}

// WithXrayService uses service for Xray configuration and process operations.
func WithXrayService(service XrayService) Option {
	return func(p *XrayProbe) {
		if service != nil {
			p.xray = service
		}
	}
}

// WithSpeedTester uses tester for proxy latency and speed measurements.
func WithSpeedTester(tester SpeedTester) Option {
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

	transferSeconds := int64(timeout.Seconds())

	p := &XrayProbe{
		pm:             pm,
		processTracker: probe.NewProcessTracker(),
		xray:           realXrayService{},
		speed:          realSpeedTester{},

		outbound:        outboundName,
		latencyTimeout:  timeout,
		transferTimeout: timeout,
		testMode:        cfg.ConnectivityTestType,
		downloadBytes:   int64(cfg.DownloadSpeed) * 1000 / 8 * transferSeconds,
		uploadBytes:     int64(cfg.UploadSpeed) * 1000 / 8 * transferSeconds,
		minDownload:     speedtest.BitsPerSec(cfg.DownloadSpeed) * speedtest.Kbps,
		minUpload:       speedtest.BitsPerSec(cfg.UploadSpeed) * speedtest.Kbps,

		remove:   os.Remove,
		waitOpen: portmgr.WaitOpen,
	}

	for _, opt := range opts {
		opt(p)
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

// Init starts process tracking for the probe.
func (p *XrayProbe) Init(ctx context.Context) error {
	p.processTracker.Start(ctx)

	return nil
}

// Run starts a temporary Xray proxy for ip and performs the configured
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

	configPath, err := p.xray.GenerateConfig(p.outbound, ip.String(), port)
	if err != nil {
		return nil, fmt.Errorf("generate Xray config: %w", err)
	}
	defer func() {
		if err := p.remove(configPath); err != nil {
			logger.CoreError("remove Xray config file: %v", err)
		}
	}()

	if err := p.xray.ValidateConfig(ctx, configPath); err != nil {
		return nil, fmt.Errorf("validate Xray config: %w", err)
	}

	proc, err := p.xray.Start(ctx, configPath)
	if err != nil {
		return nil, fmt.Errorf("start Xray: %w", err)
	}

	// Covers registration failures as well as all later failures.
	defer func() {
		if err := proc.Kill(); err != nil {
			logger.CoreError("terminate Xray: %v", err)
		}
	}()

	id, err := p.processTracker.Register(ctx, proc)
	if err != nil {
		return nil, fmt.Errorf("register Xray process: %w", err)
	}

	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if err := p.processTracker.Unregister(cleanupCtx, id); err != nil {
			logger.CoreError("unregister Xray process %s: %s", id, err)
		}
	}()

	addr := net.JoinHostPort("127.0.0.1", fmt.Sprint(port))

	if err := p.waitOpen(ctx, addr, time.Second); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}

		return nil, fmt.Errorf("wait for Xray proxy: %w", err)
	}

	latency, err := p.speed.MeasureLatency(ctx, speedtest.LatencyConfig{
		Timeout:    p.latencyTimeout,
		MaxLatency: p.maxLatency,
		ProxyPort:  port,
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

// Close releases no shared resources. Each Run cleans up its own Xray process.
func (p *XrayProbe) Close() error {
	return nil
}
