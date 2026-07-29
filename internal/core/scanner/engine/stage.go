package engine

import (
	"context"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"

	"bgscan/internal/logger"
)

// stageExecutor owns the runtime state and lifecycle of one scan stage.
type stageExecutor struct {
	stage ScanConfig
	pause PauseController

	rateCh <-chan time.Time
	start  time.Time
	total  atomic.Uint64

	processed atomic.Uint64
	succeeded atomic.Uint64

	progressDone     chan struct{}
	progressStopOnce sync.Once
}

// newStageExecutor starts the writer, initializes the probe, and begins
// progress reporting when the stage has an OnProgress hook.
func newStageExecutor(
	ctx context.Context,
	stage ScanConfig,
	pause PauseController,
	total uint64,
) (*stageExecutor, error) {
	executor := &stageExecutor{
		stage:  stage,
		pause:  pause,
		rateCh: makeRateCh(stage.Rate),
		start:  time.Now(),
	}
	executor.total.Store(total)

	if err := executor.stage.Writer.Start(); err != nil {
		return nil, err
	}

	if err := executor.stage.Probe.Init(ctx); err != nil {
		if stopErr := executor.stage.Writer.Stop(); stopErr != nil {
			executor.stage.Hooks.callOnError(stopErr)
		}

		return nil, err
	}

	executor.startProgressReporter(ctx, stage.ProgressInterval)

	return executor, nil
}

// startProgressReporter starts periodic progress reporting when configured.
func (e *stageExecutor) startProgressReporter(
	ctx context.Context,
	interval time.Duration,
) {
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

				e.reportProgress()
			}
		}
	}()
}

// cleanup stops progress reporting, releases stage resources, and invokes the
// final lifecycle hook.
func (e *stageExecutor) cleanup() {
	e.stopProgressReporter()

	if err := e.stage.Writer.Stop(); err != nil {
		e.stage.Hooks.callOnError(err)
	}

	e.reportProgress()

	if err := e.stage.Probe.Close(); err != nil {
		e.stage.Hooks.callOnError(err)
	}

	e.stage.Hooks.callOnScanEnd()
}

func (e *stageExecutor) stopProgressReporter() {
	if e.progressDone == nil {
		return
	}

	e.progressStopOnce.Do(func() {
		close(e.progressDone)
	})
}

// processIP runs the stage probe for ip and reports successful results.
func (e *stageExecutor) processIP(ctx context.Context, ip netip.Addr) bool {
	select {
	case <-e.rateCh:

	case <-ctx.Done():
		return false
	}

	res, err := e.stage.Probe.Run(ctx, ip)
	e.processed.Add(1)

	if err != nil {
		// Cancellation is expected during scanner shutdown.
		if ctx.Err() == nil {
			logger.CoreError("probe failed for %s: %v", ip, err)
		}

		return false
	}

	e.succeeded.Add(1)

	e.stage.Hooks.callOnSuccess(res)
	e.stage.Writer.Write(res)

	return true
}

func (e *stageExecutor) reportProgress() {
	if e.stage.Hooks.OnProgress == nil {
		return
	}

	reportProgress(
		e.start,
		e.pausedDuration(),
		e.total.Load(),
		e.processed.Load(),
		e.succeeded.Load(),
		e.stage.Hooks.OnProgress,
	)
}

func (e *stageExecutor) pausedDuration() time.Duration {
	if e.pause == nil {
		return 0
	}

	return e.pause.PausedDuration()
}
