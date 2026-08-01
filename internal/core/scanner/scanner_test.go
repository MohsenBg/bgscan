package scanner

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"bgscan/internal/core/config"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/engine"
	"bgscan/internal/core/scanner/portmgr"
)

// NewScanner

func TestNewScanner_NilContext(t *testing.T) {
	s, err := NewScanner(nil, "127.0.0.1") //nolint:staticcheck
	if err == nil {
		t.Fatal("expected error for nil context")
	}
	if s != nil {
		t.Fatal("expected nil scanner for nil context")
	}
}

func TestNewScanner_WithConfig(t *testing.T) {
	cfg := config.ScannerConfig{}
	s, err := NewScanner(context.Background(), "127.0.0.1", WithConfig(cfg))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = s.Close()
	}()

	sc := s.(*scanner)
	if sc.config == nil {
		t.Fatal("config not set")
	}
	if sc.pm == nil {
		t.Fatal("port manager should be created by default")
	}
	if sc.pause == nil {
		t.Fatal("pause controller should be created by default")
	}
	if sc.input != "127.0.0.1" {
		t.Fatalf("input = %q, want %q", sc.input, "127.0.0.1")
	}
	if sc.writerFactory == nil {
		t.Fatal("writer factory should default to result.NewWriter")
	}
	if sc.cancel == nil {
		t.Fatal("cancel func should be set")
	}
}

func TestNewScanner_WithPauseController(t *testing.T) {
	custom := engine.NewPauseController()
	s, err := NewScanner(
		context.Background(), "",
		WithConfig(config.ScannerConfig{}),
		WithPauseController(custom),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = s.Close()
	}()

	sc := s.(*scanner)
	if sc.pause != custom {
		t.Fatal("custom pause controller not set")
	}
}

func TestNewScanner_WithPauseController_NilIgnored(t *testing.T) {
	s, err := NewScanner(
		context.Background(), "",
		WithConfig(config.ScannerConfig{}),
		WithPauseController(nil),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() {
		_ = s.Close()
	}()

	sc := s.(*scanner)
	if sc.pause == nil {
		t.Fatal("nil pause controller should be ignored")
	}
}

func TestNewScanner_WithPortManager(t *testing.T) {
	pm, err := portmgr.New(portmgr.RandomBase(64), 64)
	if err != nil {
		t.Fatalf("create port manager: %v", err)
	}
	defer pm.Close()

	s, err := NewScanner(
		context.Background(), "",
		WithConfig(config.ScannerConfig{}),
		WithPortManager(pm),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() {
		_ = s.Close()
	}()

	sc := s.(*scanner)
	if sc.pm != pm {
		t.Fatal("custom port manager not set")
	}
}

func TestNewScanner_WithPortManager_NilIgnored(t *testing.T) {
	s, err := NewScanner(
		context.Background(), "",
		WithConfig(config.ScannerConfig{}),
		WithPortManager(nil),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() {
		_ = s.Close()
	}()

	sc := s.(*scanner)
	if sc.pm == nil {
		t.Fatal("nil port manager should be ignored; default should be installed")
	}
}

func TestNewScanner_WithWriterFactory(t *testing.T) {
	factory := func(ctx context.Context, opts result.WriterOptions) (result.Writer, error) {
		return nil, nil
	}
	s, err := NewScanner(
		context.Background(), "",
		WithConfig(config.ScannerConfig{}),
		WithWriterFactory(factory),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() {
		_ = s.Close()
	}()

	sc := s.(*scanner)
	if sc.writerFactory == nil {
		t.Fatal("writer factory not set")
	}
}

func TestNewScanner_WithWriterFactory_NilIgnored(t *testing.T) {
	s, err := NewScanner(
		context.Background(), "",
		WithConfig(config.ScannerConfig{}),
		WithWriterFactory(nil),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = s.Close() }()

	sc := s.(*scanner)
	if sc.writerFactory == nil {
		t.Fatal("nil writer factory should be ignored")
	}
}

func TestNewScanner_NilOptionSkipped(t *testing.T) {
	var nilOpt ScannerOption
	s, err := NewScanner(
		context.Background(), "",
		WithConfig(config.ScannerConfig{}),
		nilOpt,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = s.Close() }()
}

// AddStage / GetStages

func TestGetStages_Empty(t *testing.T) {
	s, err := NewScanner(context.Background(), "", WithConfig(config.ScannerConfig{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() {
		_ = s.Close()
	}()

	stages := s.GetStages()
	if len(stages) != 0 {
		t.Fatalf("expected 0 stages, got %d", len(stages))
	}
}

func TestAddStage_GetStages(t *testing.T) {
	s, err := NewScanner(context.Background(), "", WithConfig(config.ScannerConfig{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defer func() {
		_ = s.Close()
	}()

	s.AddStage(StageConfig{Workers: 1})
	s.AddStage(StageConfig{Workers: 2})

	stages := s.GetStages()
	if len(stages) != 2 {
		t.Fatalf("expected 2 stages, got %d", len(stages))
	}
	if stages[0].Workers != 1 || stages[1].Workers != 2 {
		t.Fatalf("stages out of order: %+v", stages)
	}

	// Returned slice must be a copy.
	stages[0].Workers = 999
	again := s.GetStages()
	if again[0].Workers != 1 {
		t.Fatal("GetStages did not return a copy")
	}
}

func TestAddStage_AfterClose(t *testing.T) {
	s, err := NewScanner(context.Background(), "", WithConfig(config.ScannerConfig{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = s.Close()
	s.AddStage(StageConfig{Workers: 1})
	if got := s.GetStages(); len(got) != 0 {
		t.Fatalf("stage added after close; got %d stages", len(got))
	}
}

func TestAddStage_AfterStarted(t *testing.T) {
	s, err := NewScanner(context.Background(), "", WithConfig(config.ScannerConfig{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = s.Close() }()

	sc := s.(*scanner)
	sc.started = true

	s.AddStage(StageConfig{Workers: 1})
	if got := s.GetStages(); len(got) != 0 {
		t.Fatalf("stage added after start; got %d stages", len(got))
	}
}

// StageConfig.AddHooks

func TestStageConfig_AddHooks(t *testing.T) {
	s := StageConfig{Workers: 7, Rate: 100}
	hooks := engine.ScanHooks{}
	returned := s.AddHooks(hooks)

	if returned != &s {
		t.Fatal("AddHooks should return pointer to receiver")
	}
	if s.Workers != 7 || s.Rate != 100 {
		t.Fatal("AddHooks should not modify other fields")
	}
}

// Run error paths

func TestRun_Closed(t *testing.T) {
	s, err := NewScanner(context.Background(), "", WithConfig(config.ScannerConfig{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = s.Close()

	if err := s.Run(); err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("expected closed error, got %v", err)
	}
}

func TestRun_AlreadyStarted(t *testing.T) {
	s, err := NewScanner(context.Background(), "", WithConfig(config.ScannerConfig{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = s.Close() }()

	sc := s.(*scanner)
	sc.started = true

	if err := s.Run(); err == nil || !strings.Contains(err.Error(), "already run") {
		t.Fatalf("expected already-run error, got %v", err)
	}
}

func TestRun_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s, err := NewScanner(ctx, "", WithConfig(config.ScannerConfig{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = s.Close() }()

	s.AddStage(StageConfig{Workers: 1})
	cancel()

	if err := s.Run(); err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestRun_NoStages(t *testing.T) {
	s, err := NewScanner(context.Background(), "", WithConfig(config.ScannerConfig{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = s.Close() }()

	if err := s.Run(); err == nil || !strings.Contains(err.Error(), "no stages") {
		t.Fatalf("expected no-stages error, got %v", err)
	}
}

// Pause / Resume / IsPaused / PausedDuration

func TestPause_Resume(t *testing.T) {
	s, err := NewScanner(context.Background(), "", WithConfig(config.ScannerConfig{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = s.Close() }()

	if s.IsPaused() {
		t.Fatal("should not be paused initially")
	}
	if d := s.PausedDuration(); d != 0 {
		t.Fatalf("initial paused duration = %v, want 0", d)
	}

	s.Pause()
	if !s.IsPaused() {
		t.Fatal("should be paused after Pause")
	}

	time.Sleep(30 * time.Millisecond)
	elapsed := s.PausedDuration()
	if elapsed <= 0 {
		t.Fatalf("paused duration should be positive, got %v", elapsed)
	}

	s.Resume()
	if s.IsPaused() {
		t.Fatal("should not be paused after Resume")
	}

	// After resume, paused duration should not keep growing.
	time.Sleep(20 * time.Millisecond)
	later := s.PausedDuration()
	if later < elapsed {
		t.Fatalf("paused duration shrank: %v < %v", later, elapsed)
	}
	if later-elapsed > 15*time.Millisecond {
		t.Fatalf("paused duration grew while resumed: %v", later-elapsed)
	}
}

func TestPause_Resume_Idempotent(t *testing.T) {
	s, err := NewScanner(context.Background(), "", WithConfig(config.ScannerConfig{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = s.Close() }()

	s.Pause()
	s.Pause()
	if !s.IsPaused() {
		t.Fatal("should remain paused after double pause")
	}

	s.Resume()
	s.Resume()
	if s.IsPaused() {
		t.Fatal("should remain resumed after double resume")
	}
}

func TestPause_AccumulatesAcrossCycles(t *testing.T) {
	s, err := NewScanner(context.Background(), "", WithConfig(config.ScannerConfig{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = s.Close() }()

	s.Pause()
	time.Sleep(20 * time.Millisecond)
	s.Resume()

	first := s.PausedDuration()
	if first <= 0 {
		t.Fatalf("expected positive paused duration, got %v", first)
	}

	s.Pause()
	time.Sleep(20 * time.Millisecond)
	s.Resume()

	second := s.PausedDuration()
	if second <= first {
		t.Fatalf("paused duration should accumulate: first=%v second=%v", first, second)
	}
}

// Close

func TestClose_Idempotent(t *testing.T) {
	s, err := NewScanner(context.Background(), "", WithConfig(config.ScannerConfig{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("first close failed: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("second close failed: %v", err)
	}
}

func TestClose_CancelsContext(t *testing.T) {
	s, err := NewScanner(context.Background(), "", WithConfig(config.ScannerConfig{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sc := s.(*scanner)
	if err := s.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if sc.ctx.Err() == nil {
		t.Fatal("context should be cancelled after Close")
	}
}

func TestClose_MarksClosed(t *testing.T) {
	s, err := NewScanner(context.Background(), "", WithConfig(config.ScannerConfig{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	sc := s.(*scanner)
	if !sc.closed {
		t.Fatal("closed flag should be true after Close")
	}
}

// Fake writer + writer factory

type fakeWriter struct {
	started int32
	stopped int32
	writes  int32
	path    string
}

type fakeRunner struct {
	mu sync.Mutex

	singleCalls int
	chainCalls  int
	singleInput string
	chainInput  string
	singleCfg   engine.ScanConfig
	chainCfg    engine.ChainConfig

	started chan struct{}
	wait    bool
}

func (r *fakeRunner) RunSingle(
	ctx context.Context,
	input string,
	_ uint64,
	cfg engine.ScanConfig,
	_ bool,
	_ engine.PauseController,
) {
	r.mu.Lock()
	r.singleCalls++
	r.singleInput = input
	r.singleCfg = cfg
	started := r.started
	wait := r.wait
	r.mu.Unlock()

	if started != nil {
		close(started)
	}
	if wait {
		<-ctx.Done()
	}
}

func (r *fakeRunner) RunChain(ctx context.Context, input string, _ uint64, cfg *engine.ChainConfig) {
	r.mu.Lock()
	r.chainCalls++
	r.chainInput = input
	r.chainCfg = *cfg
	started := r.started
	wait := r.wait
	r.mu.Unlock()

	if started != nil {
		close(started)
	}
	if wait {
		<-ctx.Done()
	}
}

func TestRun_OneStageUsesSingleRunner(t *testing.T) {
	runner := &fakeRunner{}
	s, err := NewScanner(
		context.Background(),
		"127.0.0.1",
		WithConfig(config.ScannerConfig{}),
		withScanRunner(runner),
	)
	if err != nil {
		t.Fatalf("NewScanner() error = %v", err)
	}
	defer func() { _ = s.Close() }()

	s.AddStage(StageConfig{Workers: 3, Rate: 25})
	if err := s.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	runner.mu.Lock()
	defer runner.mu.Unlock()
	if runner.singleCalls != 1 || runner.chainCalls != 0 {
		t.Fatalf("runner calls = single %d, chain %d; want single 1, chain 0", runner.singleCalls, runner.chainCalls)
	}
	if runner.singleInput != "127.0.0.1" {
		t.Fatalf("single input = %q", runner.singleInput)
	}
	if runner.singleCfg.Workers != 3 || runner.singleCfg.Rate != 25 {
		t.Fatalf("single config = %+v", runner.singleCfg)
	}
}

func TestRun_MultipleStagesUsesChainRunner(t *testing.T) {
	runner := &fakeRunner{}
	s, err := NewScanner(
		context.Background(),
		"targets.txt",
		WithConfig(config.ScannerConfig{}),
		withScanRunner(runner),
	)
	if err != nil {
		t.Fatalf("NewScanner() error = %v", err)
	}
	defer func() { _ = s.Close() }()

	s.AddStage(StageConfig{Workers: 1, Rate: 10})
	s.AddStage(StageConfig{Workers: 2, Rate: 20})
	if err := s.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	runner.mu.Lock()
	defer runner.mu.Unlock()
	if runner.singleCalls != 0 || runner.chainCalls != 1 {
		t.Fatalf("runner calls = single %d, chain %d; want single 0, chain 1", runner.singleCalls, runner.chainCalls)
	}
	if runner.chainInput != "targets.txt" {
		t.Fatalf("chain input = %q", runner.chainInput)
	}
	if len(runner.chainCfg.Stages) != 2 {
		t.Fatalf("chain stages = %d, want 2", len(runner.chainCfg.Stages))
	}
	if runner.chainCfg.Stages[1].Workers != 2 || runner.chainCfg.Stages[1].Rate != 20 {
		t.Fatalf("second chain stage = %+v", runner.chainCfg.Stages[1])
	}
}

func TestClose_WaitsForActiveRun(t *testing.T) {
	runner := &fakeRunner{started: make(chan struct{}), wait: true}
	s, err := NewScanner(
		context.Background(),
		"127.0.0.1",
		WithConfig(config.ScannerConfig{}),
		withScanRunner(runner),
	)
	if err != nil {
		t.Fatalf("NewScanner() error = %v", err)
	}
	s.AddStage(StageConfig{Workers: 1})

	runDone := make(chan error, 1)
	go func() { runDone <- s.Run() }()

	select {
	case <-runner.started:
	case <-time.After(time.Second):
		t.Fatal("Run() did not reach the runner")
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := <-runDone; err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func (w *fakeWriter) Start() error          { atomic.AddInt32(&w.started, 1); return nil }
func (w *fakeWriter) Stop() error           { atomic.AddInt32(&w.stopped, 1); return nil }
func (w *fakeWriter) Write(r result.Result) { atomic.AddInt32(&w.writes, 1) }
func (w *fakeWriter) GetResultPath() string { return w.path }

func TestNewWriter_FactoryError(t *testing.T) {
	factory := func(ctx context.Context, opts result.WriterOptions) (result.Writer, error) {
		return nil, errors.New("boom")
	}
	s, err := NewScanner(
		context.Background(), "",
		WithConfig(config.ScannerConfig{}),
		WithWriterFactory(factory),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = s.Close() }()

	sc := s.(*scanner)
	_, err = sc.newWriter(context.Background(), "prefix", result.ResultSchema{Name: "test"})
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected wrapped factory error, got %v", err)
	}
	if !strings.Contains(err.Error(), "test") {
		t.Fatalf("expected schema name in wrapped error, got %v", err)
	}
}

func TestNewWriter_Success(t *testing.T) {
	fw := &fakeWriter{path: "/tmp/x"}
	factory := func(ctx context.Context, opts result.WriterOptions) (result.Writer, error) {
		if opts.ResultPrefix != "prefix" {
			t.Errorf("ResultPrefix = %q, want %q", opts.ResultPrefix, "prefix")
		}
		return fw, nil
	}
	s, err := NewScanner(
		context.Background(), "",
		WithConfig(config.ScannerConfig{}),
		WithWriterFactory(factory),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = s.Close() }()

	sc := s.(*scanner)
	w, err := sc.newWriter(context.Background(), "prefix", result.ResultSchema{Name: "icmp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w != fw {
		t.Fatal("writer not returned")
	}
}

// newStage

func TestNewStage(t *testing.T) {
	s, err := NewScanner(context.Background(), "", WithConfig(config.ScannerConfig{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = s.Close() }()

	sc := s.(*scanner)
	fw := &fakeWriter{path: "p"}

	stage := sc.newStage(10, nil, fw, 100*time.Millisecond)
	if stage.Workers != 10 {
		t.Fatalf("Workers = %d, want 10", stage.Workers)
	}
	// 1s / 100ms = 10 probes/worker/s; 10 workers -> 100
	if stage.Rate != 100 {
		t.Fatalf("Rate = %d, want 100", stage.Rate)
	}
	if stage.Writer != fw {
		t.Fatal("Writer not set")
	}
	if stage.Probe != nil {
		t.Fatal("Probe should be nil (passed as nil)")
	}
}

// calcRate

func TestCalcRate(t *testing.T) {
	cases := []struct {
		workers int
		min     time.Duration
		want    int
	}{
		{0, time.Millisecond, 0},
		{5, 0, 0},
		{-1, time.Millisecond, 0},
		{5, -time.Second, 0},
		{1, time.Second, 1},
		{2, 500 * time.Millisecond, 4},
		{4, 250 * time.Millisecond, 16},
		{10, 100 * time.Millisecond, 100},
		{1000, time.Microsecond, 1_000_000_000},
	}
	for _, c := range cases {
		got := calcRate(c.workers, c.min)
		if got != c.want {
			t.Errorf("calcRate(%d, %v) = %d, want %d", c.workers, c.min, got, c.want)
		}
	}
}

// isHTTP3

func TestIsHTTP3(t *testing.T) {
	cases := map[string]bool{
		"h3":    true,
		"http3": true,
		"H3":    false,
		"HTTP3": false,
		"http2": false,
		"":      false,
		"https": false,
		"h3c":   false,
		"h33":   false,
	}
	for in, want := range cases {
		if got := isHTTP3(in); got != want {
			t.Errorf("isHTTP3(%q) = %v, want %v", in, got, want)
		}
	}
}
