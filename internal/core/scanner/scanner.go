// Package scanner provides the high-level scan orchestration layer.
package scanner

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"bgscan/internal/core/config"
	"bgscan/internal/core/config/validate"
	"bgscan/internal/core/dns"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/engine"
	"bgscan/internal/core/scanner/portmgr"
	"bgscan/internal/core/scanner/probe"
	"bgscan/internal/core/scanner/probe/dnsttprobe"
	"bgscan/internal/core/scanner/probe/httpprobe"
	"bgscan/internal/core/scanner/probe/icmpprobe"
	"bgscan/internal/core/scanner/probe/resolveprobe"
	"bgscan/internal/core/scanner/probe/slipstreamprobe"
	"bgscan/internal/core/scanner/probe/tcpprobe"
	"bgscan/internal/core/scanner/probe/vaydnsprobe"
	"bgscan/internal/core/scanner/probe/xrayprobe"
	"bgscan/internal/logger"

	"golang.org/x/time/rate"
)

var allowedPreScanTypes = []string{
	"tcp",
	"icmp",
	"none",
	"http",
}

// StageConfig describes one stage in a scan pipeline.
type StageConfig struct {
	Workers int
	Probe   probe.Probe
	Writer  result.Writer
	Hooks   engine.ScanHooks
}

// AddHooks sets the hooks invoked while this stage runs.
func (s *StageConfig) AddHooks(hooks engine.ScanHooks) *StageConfig {
	s.Hooks = hooks
	return s
}

// Scanner manages a scan pipeline and its lifecycle.
type Scanner interface {
	Run() error
	Close() error

	GetStages() []StageConfig
	AddStage(StageConfig)
	UpdateStageHooks(index int, hooks engine.ScanHooks) error

	Pause()
	Resume()
	IsPaused() bool
	PausedDuration() time.Duration

	BuildICMPStage(context.Context, ...engine.ScanHooks) (StageConfig, error)
	BuildTCPStage(context.Context, ...engine.ScanHooks) (StageConfig, error)
	BuildHTTPStage(context.Context, ...engine.ScanHooks) (StageConfig, error)
	BuildXrayStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
	BuildResolveStage(context.Context, ...engine.ScanHooks) (StageConfig, error)
	BuildDNSTTStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
	BuildSlipStreamStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
	BuildVayDNSStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
}

// WriterFactory creates a result writer for a stage.
type WriterFactory func(context.Context, result.WriterOptions) (result.Writer, error)

// scanRunner abstracts engine execution for testing.
type scanRunner interface {
	RunSingle(context.Context, string, engine.ScanConfig)
	RunChain(context.Context, string, engine.ChainConfig)
}

type engineRunner struct{}

func (engineRunner) RunSingle(
	ctx context.Context,
	input string,
	cfg engine.ScanConfig,
) {
	engine.RunScan(ctx, input, cfg)
}

func (engineRunner) RunChain(
	ctx context.Context,
	input string,
	cfg engine.ChainConfig,
) {
	engine.RunScanWithChain(ctx, input, cfg)
}

type scanner struct {
	ctx    context.Context
	cancel context.CancelFunc
	config *config.ScannerConfig

	mu      sync.Mutex
	closed  bool
	started bool
	wg      sync.WaitGroup

	pause  engine.PauseController
	input  string
	pm     portmgr.Manager
	stages []StageConfig

	writerFactory WriterFactory
	runner        scanRunner

	dnsttService      dns.DNSTTService
	slipstreamService dns.SlipstreamService
	vaydnsService     dns.VayDNSService
}

// ScannerOption configures a Scanner.
type ScannerOption func(*scanner)

// WithConfig uses cfg instead of loading scanner configuration from disk.
func WithConfig(cfg config.ScannerConfig) ScannerOption {
	return func(s *scanner) {
		s.config = &cfg
	}
}

// WithPauseController uses pause to coordinate pausing and resuming work.
func WithPauseController(pause engine.PauseController) ScannerOption {
	return func(s *scanner) {
		if pause != nil {
			s.pause = pause
		}
	}
}

// WithPortManager uses pm to allocate local ports for tunnel-based probes.
func WithPortManager(pm portmgr.Manager) ScannerOption {
	return func(s *scanner) {
		if pm != nil {
			s.pm = pm
		}
	}
}

// WithWriterFactory replaces result writer creation.
func WithWriterFactory(factory WriterFactory) ScannerOption {
	return func(s *scanner) {
		if factory != nil {
			s.writerFactory = factory
		}
	}
}

// WithDNSTTService uses service to load DNSTT configurations.
func WithDNSTTService(service dns.DNSTTService) ScannerOption {
	return func(s *scanner) {
		if service != nil {
			s.dnsttService = service
		}
	}
}

// WithSlipstreamService uses service to load Slipstream configurations.
func WithSlipstreamService(service dns.SlipstreamService) ScannerOption {
	return func(s *scanner) {
		if service != nil {
			s.slipstreamService = service
		}
	}
}

// WithVayDNSService uses service to load VayDNS configurations.
func WithVayDNSService(service dns.VayDNSService) ScannerOption {
	return func(s *scanner) {
		if service != nil {
			s.vaydnsService = service
		}
	}
}

// withScanRunner replaces engine execution in scanner tests.
func withScanRunner(runner scanRunner) ScannerOption {
	return func(s *scanner) {
		if runner != nil {
			s.runner = runner
		}
	}
}

// NewScanner creates a scanner for input.
func NewScanner(
	ctx context.Context,
	input string,
	opts ...ScannerOption,
) (Scanner, error) {
	if ctx == nil {
		return nil, errors.New("scanner context is nil")
	}

	scanCtx, cancel := context.WithCancel(ctx)

	s := &scanner{
		ctx:           scanCtx,
		cancel:        cancel,
		input:         input,
		pause:         engine.NewPauseController(),
		writerFactory: result.NewWriter,
		runner:        engineRunner{},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}

	if err := s.loadConfig(); err != nil {
		cancel()
		return nil, err
	}

	if err := s.initPortManager(); err != nil {
		cancel()
		return nil, err
	}

	if s.slipstreamService == nil {
		srv, err := dns.NewSlipstreamService()
		if err != nil {
			return nil, err
		}
		s.slipstreamService = srv
	}

	if s.dnsttService == nil {
		s.dnsttService = dns.NewDNSTTService()
	}

	if s.vaydnsService == nil {
		s.vaydnsService = dns.NewVayDNSService()
	}

	return s, nil
}

func (s *scanner) loadConfig() error {
	if s.config != nil {
		_ = validate.NormalizeAll(s.config)
		return nil
	}

	cfg, err := config.NewStore().Load()
	if err != nil {
		return fmt.Errorf("load scanner config: %w", err)
	}

	_ = validate.NormalizeAll(&cfg)
	s.config = &cfg

	return nil
}

func (s *scanner) initPortManager() error {
	if s.pm != nil {
		return nil
	}

	const portPoolSize uint16 = 3000

	pm, err := portmgr.New(
		portmgr.RandomBase(portPoolSize),
		portPoolSize,
	)
	if err != nil {
		return fmt.Errorf("create port manager: %w", err)
	}

	s.pm = pm
	return nil
}

// GetStages returns a snapshot of the configured scan stages.
func (s *scanner) GetStages() []StageConfig {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]StageConfig(nil), s.stages...)
}

// AddStage appends a stage to the scan pipeline.
func (s *scanner) AddStage(stage StageConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed || s.started {
		return
	}

	s.stages = append(s.stages, stage)
}

// UpdateStageHooks updates hooks on an existing stage.
func (s *scanner) UpdateStageHooks(
	index int,
	hooks engine.ScanHooks,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started || s.closed {
		return errors.New(
			"cannot update hooks: scanner already started or closed",
		)
	}

	if index < 0 || index >= len(s.stages) {
		return fmt.Errorf("stage index %d out of range", index)
	}

	s.stages[index].Hooks = hooks
	return nil
}

// Run executes the configured scan stages.
func (s *scanner) Run() error {
	s.mu.Lock()

	switch {
	case s.closed:
		s.mu.Unlock()
		return errors.New("scanner is closed")

	case s.started:
		s.mu.Unlock()
		return errors.New("scanner has already run")

	case s.ctx.Err() != nil:
		err := s.ctx.Err()
		s.mu.Unlock()
		return err

	case len(s.stages) == 0:
		s.mu.Unlock()
		return errors.New("scanner has no stages")
	}

	s.started = true
	stages := append([]StageConfig(nil), s.stages...)
	s.wg.Add(1)

	s.mu.Unlock()

	defer s.wg.Done()
	defer s.pause.Stop()
	defer s.pm.Close()

	if len(stages) == 1 {
		s.runSingle(stages[0])
		return nil
	}

	s.runChain(stages)
	return nil
}

func (s *scanner) runSingle(stage StageConfig) {
	general := s.config.General

	s.runner.RunSingle(s.ctx, s.input, engine.ScanConfig{
		Workers:          stage.Workers,
		MaxIPsToTest:     uint64(max(general.MaxIPsToTest, 0)),
		Probe:            stage.Probe,
		Writer:           stage.Writer,
		MinProbeDuration: general.MinProbeDuration.Duration(),
		ProgressInterval: general.StatusInterval.Duration(),
		Hooks:            stage.Hooks,
		Shuffled:         general.Shuffled,
		Pause:            s.pause,
		RateLimiter:      s.newRateLimiter(),
	})
}

func (s *scanner) runChain(stages []StageConfig) {
	general := s.config.General

	engineStages := make([]engine.StageConfig, len(stages))

	for i, stage := range stages {
		engineStages[i] = engine.StageConfig{
			Workers:          stage.Workers,
			Probe:            stage.Probe,
			Writer:           stage.Writer,
			ProgressInterval: general.StatusInterval.Duration(),
			Hooks:            stage.Hooks,
		}
	}

	s.runner.RunChain(s.ctx, s.input, engine.ChainConfig{
		MaxIPsToTest:     uint64(max(general.MaxIPsToTest, 0)),
		Mode:             engine.ParsePipelineMode(general.PipelineMode),
		Stages:           engineStages,
		Pause:            s.pause,
		Shuffled:         general.Shuffled,
		MaxBuffer:        general.MaxIPsPerStage,
		BatchSize:        general.BatchSize,
		MinProbeDuration: general.MinProbeDuration.Duration(),
		RateLimiter:      s.newRateLimiter(),
	})
}

func (s *scanner) newRateLimiter() *rate.Limiter {
	general := s.config.General

	return rate.NewLimiter(
		rate.Limit(general.ProbePerSec),
		general.ProbeBurst,
	)
}

// Pause pauses work at the next pause checkpoint.
func (s *scanner) Pause() {
	s.pause.Pause()
}

// Resume releases workers paused at a pause checkpoint.
func (s *scanner) Resume() {
	s.pause.Resume()
}

// IsPaused reports whether work is currently paused.
func (s *scanner) IsPaused() bool {
	return s.pause.IsPaused()
}

// PausedDuration reports the total time spent paused.
func (s *scanner) PausedDuration() time.Duration {
	return s.pause.PausedDuration()
}

// Close cancels the scan and waits for Run to return.
func (s *scanner) Close() error {
	s.mu.Lock()

	if s.closed {
		s.mu.Unlock()
		return nil
	}

	s.closed = true
	s.mu.Unlock()

	s.cancel()
	s.pause.Stop()

	done := make(chan struct{})

	go func() {
		s.wg.Wait()
		close(done)
	}()

	const shutdownTimeout = 5 * time.Second

	timer := time.NewTimer(shutdownTimeout)
	defer timer.Stop()

	select {
	case <-done:
		s.pm.Close()
		return nil

	case <-timer.C:
		logger.CoreError(
			"scanner shutdown timed out after %s",
			shutdownTimeout,
		)
		return errors.New("timed out waiting for scanner shutdown")
	}
}

// BuildICMPStage creates an ICMP scan stage.
func (s *scanner) BuildICMPStage(
	ctx context.Context,
	hooks ...engine.ScanHooks,
) (StageConfig, error) {
	cfg := s.config.ICMP

	prb, err := icmpprobe.NewICMPProbe(icmpprobe.Options{
		Timeout: cfg.Timeout.Duration(),
		Tries:   cfg.Tries,
	})
	if err != nil {
		return StageConfig{}, fmt.Errorf("create ICMP probe: %w", err)
	}

	writer, err := s.newWriter(
		ctx,
		cfg.OutputPrefix,
		icmpprobe.Schema,
	)
	if err != nil {
		return StageConfig{}, err
	}

	return s.newStage(cfg.Workers, prb, writer, hooks...), nil
}

// BuildTCPStage creates a TCP scan stage.
func (s *scanner) BuildTCPStage(
	ctx context.Context,
	hooks ...engine.ScanHooks,
) (StageConfig, error) {
	cfg := s.config.TCP

	prb := tcpprobe.NewTCPProbe(
		fmt.Sprint(cfg.Port),
		cfg.Timeout.Duration(),
		cfg.Tries,
	)

	writer, err := s.newWriter(
		ctx,
		cfg.OutputPrefix,
		tcpprobe.Schema,
	)
	if err != nil {
		return StageConfig{}, err
	}

	return s.newStage(cfg.Workers, prb, writer, hooks...), nil
}

// BuildHTTPStage creates an HTTP scan stage.
func (s *scanner) BuildHTTPStage(
	ctx context.Context,
	hooks ...engine.ScanHooks,
) (StageConfig, error) {
	cfg := s.config.HTTP

	req, err := httpprobe.NewHTTPRequestFromConfig(cfg)
	if err != nil {
		return StageConfig{}, fmt.Errorf("create HTTP request: %w", err)
	}

	prb, err := s.newHTTPProbe(cfg, *req)
	if err != nil {
		return StageConfig{}, err
	}

	writer, err := s.newWriter(
		ctx,
		cfg.OutputPrefix,
		httpprobe.Schema,
	)
	if err != nil {
		return StageConfig{}, err
	}

	return s.newStage(cfg.Workers, prb, writer, hooks...), nil
}

func (s *scanner) newHTTPProbe(
	cfg config.HTTPConfig,
	req httpprobe.HTTPRequest,
) (probe.Probe, error) {
	if isHTTP3(cfg.Version) {
		prb, err := httpprobe.NewHTTP3Probe(
			req,
			cfg.AcceptedStatusCodes,
		)
		if err != nil {
			return nil, fmt.Errorf("create HTTP/3 probe: %w", err)
		}
		return prb, nil
	}

	return httpprobe.NewHTTPProbe(
		req,
		cfg.AcceptedStatusCodes,
	), nil
}

func isHTTP3(version string) bool {
	return version == "h3" || version == "http3"
}

// BuildXrayStage creates the stages required for an Xray scan against
// template. If the Xray config specifies a pre-scan type (tcp, icmp, or
// http), a stage for that pre-scan is prepended so it runs ahead of the
// Xray probe in the chain. A pre-scan type of "none" (or empty) skips this
// and returns only the Xray stage.
func (s *scanner) BuildXrayStage(
	ctx context.Context,
	template string,
	hooks ...engine.ScanHooks,
) ([]StageConfig, error) {
	cfg := s.config.Xray

	stages := make([]StageConfig, 0, 2)

	preScan, err := s.buildXrayPreScanStage(ctx, cfg.PreScanType)
	if err != nil {
		return nil, err
	}

	if preScan != nil {
		stages = append(stages, *preScan)
	}

	prb, err := xrayprobe.NewXrayProbe(
		&cfg,
		template,
		s.pm,
	)
	if err != nil {
		return nil, fmt.Errorf("create Xray probe: %w", err)
	}

	writer, err := s.newWriter(
		ctx,
		cfg.OutputPrefix,
		xrayprobe.Schema,
	)
	if err != nil {
		return nil, err
	}

	stages = append(stages, s.newStage(cfg.Workers, prb, writer, hooks...))

	return stages, nil
}

// buildXrayPreScanStage builds the optional connectivity pre-scan stage
// that runs ahead of an Xray stage, based on preScanType. It returns a nil
// stage (no error) when preScanType is "none" or empty.
func (s *scanner) buildXrayPreScanStage(
	ctx context.Context,
	preScanType string,
) (*StageConfig, error) {
	switch preScanType {
	case "", "none":
		return nil, nil

	case "tcp":
		stage, err := s.BuildTCPStage(ctx)
		if err != nil {
			return nil, fmt.Errorf("create Xray TCP pre-scan: %w", err)
		}
		return &stage, nil

	case "icmp":
		stage, err := s.BuildICMPStage(ctx)
		if err != nil {
			return nil, fmt.Errorf("create Xray ICMP pre-scan: %w", err)
		}
		return &stage, nil

	case "http":
		stage, err := s.BuildHTTPStage(ctx)
		if err != nil {
			return nil, fmt.Errorf("create Xray HTTP pre-scan: %w", err)
		}
		return &stage, nil

	default:
		return nil, fmt.Errorf(
			"unsupported Xray pre_scan_type %q (allowed: %v)",
			preScanType,
			allowedPreScanTypes,
		)
	}
}

// BuildResolveStage creates a DNS resolver scan stage.
func (s *scanner) BuildResolveStage(
	ctx context.Context,
	hooks ...engine.ScanHooks,
) (StageConfig, error) {
	cfg := s.config.DNS.Resolver
	return s.buildResolverStage(ctx, cfg, hooks...)
}

// BuildDNSTTStage creates the stages required for a DNSTT scan.
func (s *scanner) BuildDNSTTStage(
	ctx context.Context,
	configName string,
	hooks ...engine.ScanHooks,
) ([]StageConfig, error) {
	tunCfg := s.config.DNS.DNSTunneling

	dnsttCfg, err := s.dnsttService.LoadConfig(configName)
	if err != nil {
		return nil, fmt.Errorf("load DNSTT config: %w", err)
	}

	stages := make([]StageConfig, 0, 2)

	if tunCfg.CheckDNSResolver {
		resolverCfg := s.config.DNS.Resolver
		resolverCfg.Port = dnsttCfg.ResolverPort
		resolverCfg.Transport = string(dnsttCfg.ResolverType)

		stage, err := s.buildResolverStage(
			ctx,
			resolverCfg,
		)
		if err != nil {
			return nil, err
		}

		stages = append(stages, stage)
	}

	prb, err := dnsttprobe.NewDNSTTProbe(
		dnsttCfg,
		tunCfg.Timeout.Duration(),
	)
	if err != nil {
		return nil, fmt.Errorf("create DNSTT probe: %w", err)
	}

	writer, err := s.newWriter(
		ctx,
		tunCfg.OutputPrefix,
		dnsttprobe.Schema,
	)
	if err != nil {
		return nil, err
	}

	stages = append(
		stages,
		s.newStage(tunCfg.Workers, prb, writer, hooks...),
	)

	return stages, nil
}

func (s *scanner) buildResolverStage(
	ctx context.Context,
	cfg config.ResolverConfig,
	hooks ...engine.ScanHooks,
) (StageConfig, error) {
	rcodes := make([]uint16, 0, len(cfg.AcceptedRCodes))
	for _, code := range cfg.AcceptedRCodes {
		rcodes = append(
			rcodes,
			uint16(dns.ParseDNSRcode(code)),
		)
	}

	prb := resolveprobe.NewResolverProbe(&resolveprobe.DNSRequest{
		Domain:          cfg.Domain,
		Port:            cfg.Port,
		RandomSubdomain: cfg.RandomSubdomain,
		DpiCheck:        cfg.DPI.Enabled,
		DpiTimeout:      cfg.DPI.Timeout.Duration(),
		DpiTries:        cfg.DPI.Tries,
		Edns0Size:       cfg.EDNSBufSize,
		CheckTypes:      cfg.CheckTypes,
		AcceptedRcodes:  rcodes,
		Timeout:         cfg.Timeout.Duration(),
		Transport:       dns.ParseResolverType(cfg.Transport),
		Tries:           cfg.Tries,
	})

	writer, err := s.newWriter(
		ctx,
		cfg.OutputPrefix,
		resolveprobe.Schema,
	)
	if err != nil {
		return StageConfig{}, err
	}

	return s.newStage(cfg.Workers, prb, writer, hooks...), nil
}

// BuildSlipStreamStage creates the stages required for a Slipstream scan.
func (s *scanner) BuildSlipStreamStage(
	ctx context.Context,
	configName string,
	hooks ...engine.ScanHooks,
) ([]StageConfig, error) {
	tunCfg := s.config.DNS.DNSTunneling

	slipstreamCfg, err := s.slipstreamService.LoadConfig(configName)
	if err != nil {
		return nil, fmt.Errorf("load Slipstream config: %w", err)
	}

	stages := make([]StageConfig, 0, 2)

	if tunCfg.CheckDNSResolver {
		resolverCfg := s.config.DNS.Resolver
		resolverCfg.Port = slipstreamCfg.ResolverPort
		resolverCfg.Transport = string(dns.ResolverTypeUDP)

		stage, err := s.buildResolverStage(
			ctx,
			resolverCfg,
		)
		if err != nil {
			return nil, err
		}

		stages = append(stages, stage)
	}

	prb, err := slipstreamprobe.NewSlipstreamProbe(
		slipstreamCfg,
		tunCfg.Timeout.Duration(),
		s.pm,
		slipstreamprobe.WithSlipstreamService(s.slipstreamService),
	)
	if err != nil {
		return nil, fmt.Errorf("create Slipstream probe: %w", err)
	}

	writer, err := s.newWriter(
		ctx,
		tunCfg.OutputPrefix,
		slipstreamprobe.Schema,
	)
	if err != nil {
		return nil, err
	}

	stages = append(
		stages,
		s.newStage(tunCfg.Workers, prb, writer, hooks...),
	)

	return stages, nil
}

// BuildVayDNSStage creates the stages required for a VayDNS scan.
func (s *scanner) BuildVayDNSStage(
	ctx context.Context,
	configName string,
	hooks ...engine.ScanHooks,
) ([]StageConfig, error) {
	tunCfg := s.config.DNS.DNSTunneling

	vaydnsCfg, err := s.vaydnsService.LoadConfig(configName)
	if err != nil {
		return nil, fmt.Errorf("load VayDNS config: %w", err)
	}

	stages := make([]StageConfig, 0, 2)

	if tunCfg.CheckDNSResolver {
		resolverCfg := s.config.DNS.Resolver
		resolverCfg.Port = vaydnsCfg.ResolverPort
		resolverCfg.Transport = string(vaydnsCfg.ResolverType)

		stage, err := s.buildResolverStage(
			ctx,
			resolverCfg,
		)
		if err != nil {
			return nil, err
		}

		stages = append(stages, stage)
	}

	prb, err := vaydnsprobe.NewVayDNSProbe(
		vaydnsCfg,
		tunCfg.Timeout.Duration(),
		vaydnsprobe.WithVayDNSService(s.vaydnsService),
	)
	if err != nil {
		return nil, fmt.Errorf("create VayDNS probe: %w", err)
	}

	writer, err := s.newWriter(
		ctx,
		tunCfg.OutputPrefix,
		vaydnsprobe.Schema,
	)
	if err != nil {
		return nil, err
	}

	stages = append(
		stages,
		s.newStage(tunCfg.Workers, prb, writer, hooks...),
	)

	return stages, nil
}

func (s *scanner) newWriter(ctx context.Context, prefix string, schema result.ResultSchema) (result.Writer, error) {
	writer, err := s.writerFactory(ctx, result.WriterOptions{
		ResultPrefix: prefix,
		Schema:       schema,
		Config:       s.config.Writer,
	})
	if err != nil {
		return nil, fmt.Errorf("create %s result writer: %w", schema.Name, err)
	}

	return writer, nil
}

func (s *scanner) newStage(
	workers int,
	prb probe.Probe,
	writer result.Writer,
	hooks ...engine.ScanHooks,
) StageConfig {
	stage := StageConfig{
		Workers: workers,
		Probe:   prb,
		Writer:  writer,
	}

	if len(hooks) > 0 {
		stage.AddHooks(hooks[0])
	}

	return stage
}
