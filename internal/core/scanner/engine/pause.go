package engine

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// PauseController controls pause, resume, and shutdown state for scanning work.
type PauseController interface {
	Pause()
	Resume()
	Stop()
	IsPaused() bool
	PausedDuration() time.Duration
	Wait(context.Context) bool
}

type pauseController struct {
	isPaused atomic.Bool

	mu       sync.RWMutex
	resumeCh chan struct{}
	pausedAt time.Time

	totalPauseNs atomic.Int64
	stopOnce     sync.Once
	doneCh       chan struct{}
}

// NewPauseController creates a controller in the running state.
func NewPauseController() PauseController {
	return &pauseController{
		resumeCh: make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Pause blocks workers at their next Wait call until Resume or Stop is called.
func (p *pauseController) Pause() {
	if !p.isPaused.CompareAndSwap(false, true) {
		return
	}

	p.mu.Lock()
	p.pausedAt = time.Now()
	p.resumeCh = make(chan struct{})
	p.mu.Unlock()
}

// Resume releases workers blocked by Wait.
func (p *pauseController) Resume() {
	if !p.isPaused.CompareAndSwap(true, false) {
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

// Stop permanently stops the controller and releases waiting workers.
func (p *pauseController) Stop() {
	p.stopOnce.Do(func() {
		p.Resume()
		close(p.doneCh)
	})
}

// IsPaused reports whether workers are currently paused.
func (p *pauseController) IsPaused() bool {
	return p.isPaused.Load()
}

// PausedDuration reports the total time spent paused, including the current
// pause interval when the controller is paused.
func (p *pauseController) PausedDuration() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()

	total := p.totalPauseNs.Load()

	if !p.pausedAt.IsZero() {
		total += time.Since(p.pausedAt).Nanoseconds()
	}

	return time.Duration(total)
}

// Wait blocks while paused.
//
// It returns false when ctx is canceled or the controller has stopped.
func (p *pauseController) Wait(ctx context.Context) bool {
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

	p.mu.RLock()
	resume := p.resumeCh
	p.mu.RUnlock()

	select {
	case <-ctx.Done():
		return false

	case <-p.doneCh:
		return false

	case <-resume:
		// Stop calls Resume before closing doneCh. Check done again so a worker
		// does not continue after shutdown if both channels became ready.
		select {
		case <-p.doneCh:
			return false
		default:
			return true
		}
	}
}
