package engine

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// PauseController manages pause, resume, and shutdown state for background tasks.
// Implementations must be safe for concurrent use. Workers call Wait at
// a checkpoint and proceed only when it returns true.
type PauseController interface {
	// Pause marks the controller paused. Workers block in Wait until
	// Resume or Stop is called.
	Pause()

	// Resume clears the paused state and releases workers blocked in
	// Wait.
	Resume()

	// Stop releases workers blocked in Wait for good. Wait returns
	// false after Stop, even if Resume is called later.
	Stop()

	// IsPaused reports whether the controller is currently paused.
	IsPaused() bool

	// PausedDuration returns the total time spent paused, including
	// a pause that is still in progress.
	PausedDuration() time.Duration

	// Wait blocks while the controller is paused and reports whether
	// the caller may proceed. It returns false when ctx is cancelled
	// or the controller is stopped.
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

// NewPauseController returns an initialized PauseController in the running state.
func NewPauseController() PauseController {
	return &pauseController{
		resumeCh: make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

func (p *pauseController) Pause() {
	if !p.isPaused.CompareAndSwap(false, true) {
		return
	}

	p.mu.Lock()
	p.pausedAt = time.Now()
	p.resumeCh = make(chan struct{})
	p.mu.Unlock()
}

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

func (p *pauseController) Stop() {
	p.stopOnce.Do(func() {
		p.Resume()
		close(p.doneCh)
	})
}

func (p *pauseController) IsPaused() bool {
	return p.isPaused.Load()
}

func (p *pauseController) PausedDuration() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()

	total := p.totalPauseNs.Load()
	if !p.pausedAt.IsZero() {
		total += time.Since(p.pausedAt).Nanoseconds()
	}

	return time.Duration(total)
}

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
		// Guard against race if Stop closed doneCh immediately after Resume.
		select {
		case <-p.doneCh:
			return false
		default:
			return true
		}
	}
}
