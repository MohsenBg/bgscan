package engine

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"sync/atomic"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/logger"

	"golang.org/x/time/rate"
)

// ErrResultLimitReached signals that the target success goal has been met.
var (
	ErrResultLimitReached = errors.New("result limit reached")
	ErrWriterFatal        = errors.New("writer fatal")
	errNoStages           = errors.New("scan chain has no stages")
)

type stageExecutor struct {
	log              *logger.Logger
	stage            StageConfig
	pause            PauseController
	rateLimiter      *rate.Limiter
	minProbeDuration time.Duration
	resultLimit      uint64

	start        time.Time
	total        atomic.Uint64
	processed    atomic.Uint64
	succeeded    atomic.Uint64
	stopProgress chan struct{}

	ignoreWriterErr bool
	writerDisabled  bool
	errs            errStats
}

func newStageExecutor(ctx context.Context, stage StageConfig, cfg ChainConfig, total, resultLimit uint64) (*stageExecutor, error) {
	exec := &stageExecutor{
		log:              cfg.Log,
		stage:            stage,
		pause:            cfg.Pause,
		rateLimiter:      cfg.RateLimiter,
		minProbeDuration: cfg.MinProbeDuration,
		resultLimit:      resultLimit,
		start:            time.Now(),
		ignoreWriterErr:  cfg.IgnoreWriterError,
	}
	exec.total.Store(total)

	if err := exec.stage.Writer.Start(); err != nil {
		if !exec.ignoreWriterErr {
			return nil, err
		}
		exec.log.Error("ignoring writer start error: %v", err)
		exec.writerDisabled = true
	}
	if err := exec.stage.Probe.Init(ctx); err != nil {
		if stopErr := exec.stage.Writer.Close(); stopErr != nil {
			exec.log.Error("stopping writer after probe init failure: %v", stopErr)
			return nil, err
		}
		return nil, err
	}

	exec.startProgressReporter(ctx, stage.ProgressInterval)
	return exec, nil
}

func (e *stageExecutor) startProgressReporter(ctx context.Context, interval time.Duration) {
	if e.stage.Hooks.OnProgress == nil || interval <= 0 {
		return
	}

	e.stopProgress = make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-e.stopProgress:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				if e.pause != nil && e.pause.IsPaused() {
					continue
				}
				e.emitProgress()
			}
		}
	}()
}

// close shuts the stage down and reports an error that may have left the
// persisted results incomplete. Probe cleanup failures are logged only; the
// scan is already over at that point.
func (e *stageExecutor) close() error {
	if e.stopProgress != nil {
		select {
		case <-e.stopProgress:
		default:
			close(e.stopProgress)
		}
	}

	var closeErr error
	if err := e.stage.Writer.Close(); err != nil {
		e.log.Error("failed to close writer: %v", err)
		if !e.ignoreWriterErr {
			closeErr = err
		}
	}

	e.emitProgress()

	if err := e.stage.Probe.Close(); err != nil {
		e.log.Error("failed to close probe: %v", err)
	}

	e.stage.Hooks.callOnScanEnd()

	return closeErr
}

func (e *stageExecutor) processIP(ctx context.Context, ip netip.Addr) error {
	if e.resultLimitReached() {
		return ErrResultLimitReached
	}

	if e.rateLimiter != nil {
		if err := e.rateLimiter.Wait(ctx); err != nil {
			return err
		}
	}

	probeStart := time.Now()

	res, err := e.stage.Probe.Run(ctx, ip)
	e.processed.Add(1)

	e.waitMinProbeDuration(ctx, probeStart)

	if err != nil {
		if ctx.Err() != nil {
			return err
		}
		return probe.NormalizeErr(err)
	}

	n, ok := e.claimResultSlot()
	if !ok {
		return ErrResultLimitReached
	}

	e.stage.Hooks.callOnSuccess(res)

	if !e.writerDisabled {
		if err := e.stage.Writer.Write(res); err != nil {
			if e.ignoreWriterErr {
				if ctx.Err() == nil {
					e.log.Warn("ignoring writer error: %v", err)
				}
			} else {
				if ctx.Err() != nil {
					return err
				}
				return fmt.Errorf("%w: %w", ErrWriterFatal, err)
			}
		}
	}

	if e.resultLimit > 0 && n >= e.resultLimit {
		return ErrResultLimitReached
	}

	return nil
}

func (e *stageExecutor) resultLimitReached() bool {
	return e.resultLimit > 0 && e.succeeded.Load() >= e.resultLimit
}

func (e *stageExecutor) claimResultSlot() (uint64, bool) {
	if e.resultLimit == 0 {
		return e.succeeded.Add(1), true
	}

	for {
		cur := e.succeeded.Load()
		if cur >= e.resultLimit {
			return cur, false
		}
		if e.succeeded.CompareAndSwap(cur, cur+1) {
			return cur + 1, true
		}
	}
}

func (e *stageExecutor) waitMinProbeDuration(ctx context.Context, probeStart time.Time) {
	if e.minProbeDuration <= 0 {
		return
	}

	if remaining := e.minProbeDuration - time.Since(probeStart); remaining > 0 {
		select {
		case <-time.After(remaining):
		case <-ctx.Done():
		}
	}
}

func (e *stageExecutor) pausedDuration() time.Duration {
	if e.pause == nil {
		return 0
	}
	return e.pause.PausedDuration()
}

func (e *stageExecutor) emitProgress() {
	reportProgress(
		e.start,
		e.pausedDuration(),
		e.total.Load(),
		e.processed.Load(),
		e.succeeded.Load(),
		e.errs.snapshot(),
		e.stage.Hooks.OnProgress,
	)
}
