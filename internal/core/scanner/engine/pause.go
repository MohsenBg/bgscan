package engine

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// PauseController manages the pause, resume, and shutdown state of active scanning loops.
// It is fully thread-safe and can be shared safely across multiple worker pools.
type PauseController struct {
	isPaused atomic.Bool

	mu       sync.RWMutex
	resumeCh chan struct{}
	pausedAt time.Time

	// Accumulates paused time in nanoseconds
	totalPauseNs atomic.Int64
	stopOnce     sync.Once
	doneCh       chan struct{}
}

// NewPauseController instantiates an initialized PauseController.
func NewPauseController() *PauseController {
	return &PauseController{
		resumeCh: make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Pause transitions the controller into a paused state.
// Workers encountering Wait() will block until Resume() or Stop() is called.
func (p *PauseController) Pause() {
	if !p.isPaused.CompareAndSwap(false, true) {
		// Already paused
		return
	}

	p.mu.Lock()
	p.pausedAt = time.Now()

	// Re-create the resume channel because the previous one might have been closed
	p.resumeCh = make(chan struct{})
	p.mu.Unlock()
}

// Resume transitions the controller out of a paused state, releasing any waiting workers.
func (p *PauseController) Resume() {
	if !p.isPaused.CompareAndSwap(true, false) {
		// Already running
		return
	}

	p.mu.Lock()
	if !p.pausedAt.IsZero() {
		p.totalPauseNs.Add(time.Since(p.pausedAt).Nanoseconds())
		p.pausedAt = time.Time{}
	}
	close(p.resumeCh)
	p.mu.Unlock()
}

// Stop signals to all waiting workers that the controller is permanently shutting down.
func (p *PauseController) Stop() {
	p.stopOnce.Do(func() {
		// Ensure we aren't leaving workers hung on a pause state during shutdown
		p.Resume()
		close(p.doneCh)
	})
}

// IsPaused returns true if the engine is currently in a paused state.
func (p *PauseController) IsPaused() bool {
	return p.isPaused.Load()
}

// PausedDuration returns the total accumulated time spent in a paused state.
func (p *PauseController) PausedDuration() time.Duration {
	p.mu.RLock()
	total := p.totalPauseNs.Load()
	if !p.pausedAt.IsZero() {
		total += time.Since(p.pausedAt).Nanoseconds()
	}
	p.mu.RUnlock()
	return time.Duration(total)
}

// Wait blocks the calling goroutine if the controller is paused.
// It returns true if execution should continue, or false if the engine has been stopped
// or the execution context was cancelled.
func (p *PauseController) Wait(ctx context.Context) bool {
	// Fast path: if not paused and not stopped, keep moving.
	if !p.isPaused.Load() {
		select {
		case <-ctx.Done():
			return false
		case <-p.doneCh:
			return false
		default:
			return true
		}
	}

	// Slow path: we are paused, look up the channel and wait for broad-casted events.
	p.mu.RLock()
	resume := p.resumeCh
	p.mu.RUnlock()

	select {
	case <-ctx.Done():
		return false
	case <-p.doneCh:
		return false
	case <-resume:
		return true
	}
}
