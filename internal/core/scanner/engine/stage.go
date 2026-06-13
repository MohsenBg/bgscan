package engine

import (
	"bgscan/internal/core/config"
	"bgscan/internal/logger"
	"context"
	"sync/atomic"
	"time"
)

// stageExecutor manages the execution state, metrics, and lifecycle of a scan stage.
type stageExecutor struct {
	ctx    context.Context
	stage  ScanConfig
	pause  *PauseController
	rateCh <-chan time.Time

	start     time.Time
	total     atomic.Uint64
	processed atomic.Uint64
	succeed   atomic.Uint64

	progressDone chan struct{}
}

// newStageExecutor creates and initializes a stage executor.
func newStageExecutor(ctx context.Context, stage ScanConfig, pause *PauseController, total uint64) *stageExecutor {
	exec := &stageExecutor{
		ctx:    ctx,
		stage:  stage,
		pause:  pause,
		rateCh: makeRateCh(stage.Rate),
		start:  time.Now(),
	}

	exec.total.Store(total)
	exec.stage.Writer.Start()
	exec.stage.Probe.Init(ctx)
	exec.startProgressReporter()

	return exec
}

// startProgressReporter periodically reports stage progress.
func (e *stageExecutor) startProgressReporter() {
	if e.stage.Hooks.OnProgress == nil {
		return
	}

	e.progressDone = make(chan struct{})

	go func() {
		interval := config.Get().General.StatusInterval.Duration()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-e.progressDone:
				return
			case <-e.ctx.Done():
				return
			case <-ticker.C:
				if e.pause != nil && e.pause.IsPaused() {
					continue
				}

				reportProgress(
					e.start,
					e.getPauseDuration(),
					e.total.Load(),
					e.processed.Load(),
					e.succeed.Load(),
					e.stage.Hooks.OnProgress,
				)
			}
		}
	}()
}

// cleanup releases stage resources and emits a final progress update.
func (e *stageExecutor) cleanup() {
	if e.progressDone != nil {
		select {
		case <-e.progressDone:
		default:
			close(e.progressDone)
		}
	}

	e.stage.Writer.Stop()

	reportProgress(
		e.start,
		e.getPauseDuration(),
		e.total.Load(),
		e.processed.Load(),
		e.succeed.Load(),
		e.stage.Hooks.OnProgress,
	)

	if err := e.stage.Probe.Close(); err != nil {
		e.stage.Hooks.callOnError(err)
	}

	e.stage.Hooks.callOnScanEnd()
}

// processIP executes the stage probe against an IP and returns whether it matched.
func (e *stageExecutor) processIP(ip string) bool {
	select {
	case <-e.rateCh:
	case <-e.ctx.Done():
		return false
	}

	res, err := e.stage.Probe.Run(e.ctx, ip)
	e.processed.Add(1)

	if err != nil {
		logger.CoreError("probe failed for %s: %v", ip, err)
		return false
	}

	e.succeed.Add(1)
	e.stage.Hooks.callOnSuccess(*res)
	e.stage.Writer.Write(*res)

	return true
}

// getPauseDuration returns the accumulated paused duration.
func (e *stageExecutor) getPauseDuration() time.Duration {
	if e.pause == nil {
		return 0
	}

	return e.pause.PausedDuration()
}

