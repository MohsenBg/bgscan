// Package scanner provides the high-level scan orchestration layer.
package scanner

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/core/config/validate"
	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner/engine"
	"github.com/MohsenBg/bgscan/internal/core/scanner/portmgr"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/logger"

	"golang.org/x/time/rate"
)

const (
	closeTimeout = time.Second * 10
)

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
	AddStage(StageConfig) error
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
	BuildMasterDNSStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
	BuildStormDNSStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
	BuildTheFeedStage(context.Context, string, ...engine.ScanHooks) ([]StageConfig, error)
}

// WriterFactory creates a result writer for a stage.
type WriterFactory func(context.Context, result.WriterOptions) (result.Writer, error)

// scanRunner abstracts engine execution for testing.
type scanRunner interface {
	RunChain(context.Context, string, engine.ChainConfig)
}

type engineRunner struct{}

func (engineRunner) RunChain(ctx context.Context, input string, cfg engine.ChainConfig) {
	engine.RunScanWithChain(ctx, input, cfg)
}

type scanner struct {
	ctx    context.Context
	cancel context.CancelFunc
	config *config.ScannerConfig

	closeOnce sync.Once
	closed    atomic.Bool
	closeErr  error

	started atomic.Bool
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
	masterDNSService  dns.MasterDNSService
	stormDNSService   dns.StormDNSService
	thefeedService    dns.TheFeedService

	log *logger.Logger
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

// WithLog provides the core.log file logger used by the scanner and its probes.
func WithLog(l *logger.Logger) ScannerOption {
	return func(s *scanner) {
		if l != nil {
			s.log = l
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

// WithMasterDNSService uses service to load MasterDNS configurations.
func WithMasterDNSService(service dns.MasterDNSService) ScannerOption {
	return func(s *scanner) {
		if service != nil {
			s.masterDNSService = service
		}
	}
}

// WithStormDNSService uses service to load StormDNS configurations.
func WithStormDNSService(service dns.StormDNSService) ScannerOption {
	return func(s *scanner) {
		if service != nil {
			s.stormDNSService = service
		}
	}
}

// WithTheFeedService uses service to load TheFeed configurations.
func WithTheFeedService(service dns.TheFeedService) ScannerOption {
	return func(s *scanner) {
		if service != nil {
			s.thefeedService = service
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

// NewScanner creates a scanner for the target file at input.
func NewScanner(ctx context.Context, input string, opts ...ScannerOption) (Scanner, error) {
	if ctx == nil {
		return nil, errors.New("scanner context cannot be nil")
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

	if s.log == nil {
		s.log = logger.Discard()
	}

	if err := s.loadConfig(); err != nil {
		cancel()
		return nil, fmt.Errorf("load config: %w", err)
	}

	if err := s.initPortManager(); err != nil {
		cancel()
		return nil, fmt.Errorf("init port manager: %w", err)
	}

	if err := s.initDNSServices(); err != nil {
		cancel()
		return nil, err
	}

	return s, nil
}

func (s *scanner) loadConfig() error {
	if s.config != nil {
		_ = validate.NormalizeAll(s.config)
		return nil
	}

	cfg, err := config.NewStore(config.WithLogger(s.log)).Load()
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
	pm, err := portmgr.New(portmgr.RandomBase(portPoolSize), portPoolSize)
	if err != nil {
		return fmt.Errorf("create port manager: %w", err)
	}

	s.pm = pm
	return nil
}

func (s *scanner) initDNSServices() error {
	if s.slipstreamService == nil {
		srv, err := dns.NewSlipstreamService()
		if err != nil {
			return fmt.Errorf("init slipstream service: %w", err)
		}
		s.slipstreamService = srv
	}

	if s.dnsttService == nil {
		s.dnsttService = dns.NewDNSTTService()
	}
	if s.vaydnsService == nil {
		s.vaydnsService = dns.NewVayDNSService()
	}
	if s.masterDNSService == nil {
		s.masterDNSService = dns.NewMasterDNSService()
	}
	if s.stormDNSService == nil {
		s.stormDNSService = dns.NewStormDNSService()
	}
	if s.thefeedService == nil {
		s.thefeedService = dns.NewTheFeedService()
	}

	return nil
}

// GetStages returns a snapshot of the configured scan stages.
func (s *scanner) GetStages() []StageConfig {
	return slices.Clone(s.stages)
}

// AddStage appends a stage to the scan pipeline.
func (s *scanner) AddStage(stage StageConfig) error {
	if s.started.Load() {
		return errors.New("scanner already started")
	}
	if s.closed.Load() {
		return errors.New("scanner is closed")
	}

	s.stages = append(s.stages, stage)
	return nil
}

// UpdateStageHooks updates hooks on an existing stage.
func (s *scanner) UpdateStageHooks(index int, hooks engine.ScanHooks) error {
	if s.started.Load() || s.closed.Load() {
		return errors.New("cannot update hooks: scanner already started or closed")
	}
	if index < 0 || index >= len(s.stages) {
		return fmt.Errorf("stage index %d out of range", index)
	}

	s.stages[index].Hooks = hooks
	return nil
}

// Run executes the configured scan stages.
func (s *scanner) Run() error {
	switch {
	case s.closed.Load():
		return errors.New("scanner is closed")
	case s.started.Load():
		return errors.New("scanner has already run")
	case s.ctx.Err() != nil:
		return fmt.Errorf("scanner context cancelled: %w", s.ctx.Err())
	case len(s.stages) == 0:
		return errors.New("scanner has no stages")
	}

	if !s.started.CompareAndSwap(false, true) {
		return errors.New("scanner has already run")
	}

	if config.DetectPlatform() == config.Android {
		if err := exec.Command("termux-wake-lock").Run(); err != nil {
			s.log.Warn("termux-wake-lock failed: %v", err)
		}

		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := exec.CommandContext(ctx, "termux-wake-unlock").Run(); err != nil {
				s.log.Warn("termux-wake-unlock failed: %v", err)
			}
		}()
	}

	stages := slices.Clone(s.stages)

	s.wg.Add(1)
	defer s.wg.Done()
	defer s.pause.Stop()

	s.runPipeline(stages)
	return nil
}

// runPipeline executes the scan stages, handling both single and multi-stage chains uniformly.
func (s *scanner) runPipeline(stages []StageConfig) {
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

	mode := engine.ModeSequential
	if len(stages) > 1 {
		mode = engine.ParsePipelineMode(general.PipelineMode)
	}

	s.runner.RunChain(s.ctx, s.input, engine.ChainConfig{
		MaxIPsToTest:     uint64(max(general.MaxIPsToTest, 0)),
		MaxSuccessfulIPs: uint64(max(general.MaxSuccessfulIPs, 0)),
		Mode:             mode,
		Stages:           engineStages,
		Pause:            s.pause,
		Shuffled:         general.Shuffled,
		MaxBuffer:        general.MaxIPsPerStage,
		BatchSize:        general.BatchSize,
		MinProbeDuration: general.MinProbeDuration.Duration(),
		RateLimiter:      s.newRateLimiter(),
		Log:              s.log,
	})
}

func (s *scanner) newRateLimiter() *rate.Limiter {
	general := s.config.General
	return rate.NewLimiter(rate.Limit(general.ProbePerSec), general.ProbeBurst)
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

// Close cancels the scan and waits for Run to return, cleaning up resources.
func (s *scanner) Close() error {
	s.closeOnce.Do(func() {
		s.closed.Store(true)
		s.cancel()
		s.pause.Stop()

		done := make(chan struct{})
		go func() {
			s.wg.Wait()
			if s.pm != nil {
				s.pm.Close()
			}
			close(done)
		}()

		timer := time.NewTimer(closeTimeout)
		defer timer.Stop()

		select {
		case <-done:
		case <-timer.C:
			s.log.Error("scanner shutdown timed out after %s", closeTimeout)
			s.closeErr = errors.New("timed out waiting for scanner shutdown")
		}
	})
	return s.closeErr
}
