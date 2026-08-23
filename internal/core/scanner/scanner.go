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
