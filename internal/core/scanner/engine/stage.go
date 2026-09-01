package engine

import (
	"context"
	"net/netip"
	"sync/atomic"
	"time"

	"github.com/MohsenBg/bgscan/internal/logger"

	"golang.org/x/time/rate"
)

// stageExecutor manages execution state, metrics, and lifecycle of a single scan stage.
type stageExecutor struct {
	stage            StageConfig
	pause            PauseController
	rateLimiter      *rate.Limiter
	minProbeDuration time.Duration

	start        time.Time
	total        atomic.Uint64
	processed    atomic.Uint64
	succeed      atomic.Uint64
	progressDone chan struct{}
}

// newStageExecutor initialises a stage executor and starts background progress reporting.
func newStageExecutor(ctx context.Context, stage StageConfig, cfg ChainConfig, total uint64) (*stageExecutor, error) {
	exec := &stageExecutor{
		stage:            stage,
		pause:            cfg.Pause,
		rateLimiter:      cfg.RateLimiter,
		minProbeDuration: cfg.MinProbeDuration,
		start:            time.Now(),
	}
	exec.total.Store(total)

	if err := exec.stage.Writer.Start(); err != nil {
		return nil, err
	}
	if err := exec.stage.Probe.Init(ctx); err != nil {
		_ = exec.stage.Writer.Stop()
		return nil, err
	}

	exec.startProgressReporter(ctx, stage.ProgressInterval)
	return exec, nil
}

// startProgressReporter emits progress at regular intervals until the stage ends or ctx is cancelled.
func (e *stageExecutor) startProgressReporter(ctx context.Context, interval time.Duration) {
	if e.stage.Hooks.OnProgress == nil || interval <= 0 {
		return
	}

	e.progressDone = make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-e.progressDone:
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

// cleanup stops the writer, emits a final progress snapshot, and closes the probe.
// Always call this via defer after newStageExecutor succeeds.
func (e *stageExecutor) cleanup() {
	if e.progressDone != nil {
		select {
		case <-e.progressDone:
		default:
			close(e.progressDone)
		}
	}

	if err := e.stage.Writer.Stop(); err != nil {
		e.stage.Hooks.callOnError(err)
	}

	e.emitProgress()

	if err := e.stage.Probe.Close(); err != nil {
		e.stage.Hooks.callOnError(err)
	}

	e.stage.Hooks.callOnScanEnd()
}

// processIP acquires a rate-limit token, runs the probe, then enforces the minimum
// probe duration to prevent socket bursts on limited devices such as Android/Termux.
// Returns true if the probe matched.
func (e *stageExecutor) processIP(ctx context.Context, ip netip.Addr) bool {
	if e.rateLimiter != nil {
		if err := e.rateLimiter.Wait(ctx); err != nil {
			return false
		}
	}

	probeStart := time.Now()

	res, err := e.stage.Probe.Run(ctx, ip)
	e.processed.Add(1)

	if err != nil {
		if ctx.Err() == nil {
			logger.CoreError("probe failed for %s: %v", ip, err)
		}
		e.enforceMinProbeDuration(ctx, probeStart)
		return false
	}

	e.succeed.Add(1)
	e.stage.Hooks.callOnSuccess(res)
	e.stage.Writer.Write(res)

	e.enforceMinProbeDuration(ctx, probeStart)
	return true
}

// enforceMinProbeDuration sleeps for the remainder of MinProbeDuration if the
// probe finished early. Gives the kernel time to recycle TIME_WAIT sockets.
func (e *stageExecutor) enforceMinProbeDuration(ctx context.Context, probeStart time.Time) {
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

// emitProgress reports the current scan metrics to the progress hook.
func (e *stageExecutor) emitProgress() {
	reportProgress(
		e.start,
		e.pausedDuration(),
		e.total.Load(),
		e.processed.Load(),
		e.succeed.Load(),
		e.stage.Hooks.OnProgress,
	)
}

// pausedDuration returns the cumulative paused duration, or zero if pause is disabled.
func (e *stageExecutor) pausedDuration() time.Duration {
	if e.pause == nil {
		return 0
	}
	return e.pause.PausedDuration()
}
