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
	"bgscan/internal/core/scanner/probe/xrayprobe"
	"bgscan/internal/logger"
)

// StageConfig describes one stage in a scan pipeline.
type StageConfig struct {
	Workers int
	Probe   probe.Probe
	Writer  result.Writer
	Rate    int
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

	Pause()
	Resume()
	IsPaused() bool
	PausedDuration() time.Duration

	BuildICMPStage(context.Context) (StageConfig, error)
	BuildTCPStage(context.Context) (StageConfig, error)
	BuildHTTPStage(context.Context) (StageConfig, error)
	BuildXrayStage(context.Context, string) (StageConfig, error)
	BuildResolveStage(context.Context) (StageConfig, error)
	BuildDNSTTStage(context.Context) (StageConfig, error)
	BuildSlipStreamStage(context.Context) (StageConfig, error)
}

// WriterFactory creates a result writer for a stage.
type WriterFactory func(context.Context, result.WriterOptions) (result.Writer, error)

// scanRunner executes scanner stages through the scan engine.
//
// Keeping this dependency small lets scanner tests verify the assembled engine
// configuration without starting workers or performing network operations.
type scanRunner interface {
	RunSingle(context.Context, string, uint64, engine.ScanConfig, bool, engine.PauseController)
	RunChain(context.Context, string, uint64, *engine.ChainConfig)
}

type engineRunner struct{}

func (engineRunner) RunSingle(
	ctx context.Context,
	input string,
	maxIPs uint64,
	cfg engine.ScanConfig,
	shuffled bool,
	pause engine.PauseController,
) {
	engine.RunScan(ctx, input, maxIPs, cfg, shuffled, pause)
}

func (engineRunner) RunChain(ctx context.Context, input string, maxIPs uint64, cfg *engine.ChainConfig) {
	engine.RunScanWithChain(ctx, input, maxIPs, cfg)
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
//
// It is primarily useful when constructing scanner stages in tests.
func WithWriterFactory(factory WriterFactory) ScannerOption {
	return func(s *scanner) {
		if factory != nil {
			s.writerFactory = factory
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
func NewScanner(ctx context.Context, input string, opts ...ScannerOption) (Scanner, error) {
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

	if s.config == nil {
		store := config.NewStore()

		cfg, err := store.Load()
		if err != nil {
			cancel()
			return nil, fmt.Errorf("load scanner config: %w", err)
		}

		_ = validate.NormalizeAll(&cfg)
		s.config = &cfg
	}

	if s.pm == nil {
		const portPoolSize uint16 = 3000

		pm, err := portmgr.New(portmgr.RandomBase(portPoolSize), portPoolSize)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("create port manager: %w", err)
		}

		s.pm = pm
	}

	return s, nil
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

	if !s.closed && !s.started {
		s.stages = append(s.stages, stage)
	}
}

// Run executes the configured scan stages.
func (s *scanner) Run() error {
	s.mu.Lock()

	if s.closed {
		s.mu.Unlock()
		return errors.New("scanner is closed")
	}

	if s.started {
		s.mu.Unlock()
		return errors.New("scanner has already run")
	}

	if s.ctx.Err() != nil {
		s.mu.Unlock()
		return s.ctx.Err()
	}

	if len(s.stages) == 0 {
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
	maxIPs := max(s.config.General.MaxIPsToTest, 0)

	s.runner.RunSingle(
		s.ctx,
		s.input,
		uint64(maxIPs),
		engine.ScanConfig{
			Workers:          stage.Workers,
			Probe:            stage.Probe,
			Writer:           stage.Writer,
			Rate:             stage.Rate,
			ProgressInterval: s.config.General.StatusInterval.Duration(),
			Hooks:            stage.Hooks,
		},
		s.config.General.Shuffled,
		s.pause,
	)
}

func (s *scanner) runChain(stages []StageConfig) {
	maxIPs := max(s.config.General.MaxIPsToTest, 0)

	engineStages := make([]engine.ScanConfig, len(stages))
	for i, stage := range stages {
		engineStages[i] = engine.ScanConfig{
			Workers:          stage.Workers,
			Probe:            stage.Probe,
			Writer:           stage.Writer,
			Rate:             stage.Rate,
			ProgressInterval: s.config.General.StatusInterval.Duration(),
			Hooks:            stage.Hooks,
		}
	}

	s.runner.RunChain(s.ctx, s.input, uint64(maxIPs), &engine.ChainConfig{
		Mode:      engine.ParsePipelineMode(s.config.General.PipelineMode),
		Stages:    engineStages,
		Pause:     s.pause,
		Shuffled:  s.config.General.Shuffled,
		MaxBuffer: s.config.General.MaxIPsPerStage,
		BatchSize: s.config.General.BatchSize,
	})
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

// PausedDuration reports the total time the scanner has spent paused.
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

	select {
	case <-done:
		s.pm.Close()
		return nil

	case <-time.After(shutdownTimeout):
		logger.CoreError("scanner shutdown timed out after %s", shutdownTimeout)
		return errors.New("timed out waiting for scanner shutdown")
	}
}

// BuildICMPStage creates an ICMP scan stage from the scanner configuration.
func (s *scanner) BuildICMPStage(ctx context.Context) (StageConfig, error) {
	cfg := s.config.ICMP

	prb, err := icmpprobe.NewICMPProbe(icmpprobe.Options{
		Timeout: cfg.Timeout.Duration(),
		Tries:   cfg.Tries,
	})
	if err != nil {
		return StageConfig{}, fmt.Errorf("create ICMP probe: %w", err)
	}

	writer, err := s.newWriter(ctx, cfg.OutputPrefix, icmpprobe.Schema)
	if err != nil {
		return StageConfig{}, err
	}

	return s.newStage(cfg.Workers, prb, writer, 25*time.Millisecond), nil
}

// BuildTCPStage creates a TCP scan stage from the scanner configuration.
func (s *scanner) BuildTCPStage(ctx context.Context) (StageConfig, error) {
	cfg := s.config.TCP
	prb := tcpprobe.NewTCPProbe(fmt.Sprint(cfg.Port), cfg.Timeout.Duration(), cfg.Tries)

	writer, err := s.newWriter(ctx, cfg.OutputPrefix, tcpprobe.Schema)
	if err != nil {
		return StageConfig{}, err
	}

	return s.newStage(cfg.Workers, prb, writer, 50*time.Millisecond), nil
}

// BuildHTTPStage creates an HTTP scan stage from the scanner configuration.
func (s *scanner) BuildHTTPStage(ctx context.Context) (StageConfig, error) {
	cfg := s.config.HTTP

	req, err := httpprobe.NewHTTPRequestFromConfig(cfg)
	if err != nil {
		return StageConfig{}, fmt.Errorf("create HTTP request: %w", err)
	}

	var prb probe.Probe
	if isHTTP3(cfg.Version) {
		prb, err = httpprobe.NewHTTP3Probe(*req, cfg.AcceptedStatusCodes)
		if err != nil {
			return StageConfig{}, fmt.Errorf("create HTTP/3 probe: %w", err)
		}
	} else {
		prb = httpprobe.NewHTTPProbe(*req, cfg.AcceptedStatusCodes)
	}

	writer, err := s.newWriter(ctx, cfg.OutputPrefix, httpprobe.Schema)
	if err != nil {
		return StageConfig{}, err
	}

	return s.newStage(cfg.Workers, prb, writer, 80*time.Millisecond), nil
}

func isHTTP3(version string) bool {
	return version == "h3" || version == "http3"
}

// BuildXrayStage creates an Xray scan stage for template.
func (s *scanner) BuildXrayStage(ctx context.Context, template string) (StageConfig, error) {
	cfg := s.config.Xray

	prb, err := xrayprobe.NewXrayProbe(&cfg, template, s.pm)
	if err != nil {
		return StageConfig{}, fmt.Errorf("create Xray probe: %w", err)
	}

	writer, err := s.newWriter(ctx, cfg.OutputPrefix, xrayprobe.Schema)
	if err != nil {
		return StageConfig{}, err
	}

	return s.newStage(cfg.Workers, prb, writer, 200*time.Millisecond), nil
}

// BuildResolveStage creates a DNS resolver scan stage.
func (s *scanner) BuildResolveStage(ctx context.Context) (StageConfig, error) {
	cfg := s.config.DNS.Resolver

	rcodes := make([]uint16, 0, len(cfg.AcceptedRCodes))
	for _, code := range cfg.AcceptedRCodes {
		rcodes = append(rcodes, uint16(dns.ParseDNSRcode(code)))
	}

	prb := resolveprobe.NewResolverProbe(&resolveprobe.DNSRequest{
		Domain:          cfg.Domain,
		Port:            cfg.Port,
		RandomSubdomain: cfg.RandomSubdomain,
		DpiCheck:        cfg.CheckDPI,
		DpiTimeout:      cfg.DPITimeout.Duration(),
		DpiTries:        cfg.DPITries,
		Edns0Size:       cfg.EDNSBufSize,
		CheckTypes:      cfg.CheckTypes,
		AcceptedRcodes:  rcodes,
		Timeout:         cfg.Timeout.Duration(),
		Transport:       dns.ParseTransport(cfg.Protocol),
		Tries:           cfg.Tries,
	})

	writer, err := s.newWriter(ctx, cfg.PrefixOutput, resolveprobe.Schema)
	if err != nil {
		return StageConfig{}, err
	}

	return s.newStage(cfg.Workers, prb, writer, 500*time.Millisecond), nil
}

// BuildDNSTTStage creates a DNSTT tunnel scan stage.
func (s *scanner) BuildDNSTTStage(ctx context.Context) (StageConfig, error) {
	cfg := s.config.DNS.DNSTT
	resolver := s.config.DNS.Resolver

	prb, err := dnsttprobe.NewDNSTTProbe(dnsttprobe.DNSTTConfig{
		Domain:    cfg.Domain,
		PubKey:    cfg.PublicKey,
		Transport: dns.ParseTransport(resolver.Protocol),
		DNSPort:   resolver.Port,
		Timeout:   cfg.Timeout.Duration(),
	}, s.pm)
	if err != nil {
		return StageConfig{}, fmt.Errorf("create DNSTT probe: %w", err)
	}

	writer, err := s.newWriter(ctx, cfg.OutputPrefix, dnsttprobe.Schema)
	if err != nil {
		return StageConfig{}, err
	}

	return s.newStage(cfg.Workers, prb, writer, time.Second), nil
}

// BuildSlipStreamStage creates a Slipstream tunnel scan stage.
func (s *scanner) BuildSlipStreamStage(ctx context.Context) (StageConfig, error) {
	cfg := s.config.DNS.SlipStream
	resolver := s.config.DNS.Resolver

	prb, err := slipstreamprobe.NewSlipstreamProbe(slipstreamprobe.SlipstreamConfig{
		Domain:   cfg.Domain,
		CertPath: cfg.CertPath,
		DNSPort:  resolver.Port,
		Timeout:  cfg.Timeout.Duration(),
	}, s.pm)
	if err != nil {
		return StageConfig{}, fmt.Errorf("create Slipstream probe: %w", err)
	}

	writer, err := s.newWriter(ctx, cfg.OutputPrefix, slipstreamprobe.Schema)
	if err != nil {
		return StageConfig{}, err
	}

	return s.newStage(cfg.Workers, prb, writer, time.Second), nil
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

func (s *scanner) newStage(workers int, prb probe.Probe, writer result.Writer, minimumProbeTime time.Duration) StageConfig {
	return StageConfig{
		Workers: workers,
		Probe:   prb,
		Writer:  writer,
		Rate:    calcRate(workers, minimumProbeTime),
	}
}

func calcRate(workers int, minimumProbeTime time.Duration) int {
	if workers <= 0 || minimumProbeTime <= 0 {
		return 0
	}

	return int(time.Second/minimumProbeTime) * workers
}
