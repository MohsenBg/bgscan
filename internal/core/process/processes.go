package process

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// Killable represents a resource or process that can be explicitly terminated.
type Killable interface {
	Kill() error
}

// ProcessTracker registers probe processes and terminates registered processes
// when its monitoring context is canceled.
type ProcessTracker interface {
	Start(context.Context)
	Register(context.Context, Killable) (string, error)
	Unregister(context.Context, string) error
}

type opType uint8

const (
	opAdd opType = iota
	opRemove
)

type action struct {
	id   string
	proc Killable
	op   opType
}

type processTracker struct {
	actionQueue chan action
	startOnce   sync.Once
}

// NewProcessTracker creates a process tracker.
//
// Start must be called before the tracker can process registrations.
func NewProcessTracker() ProcessTracker {
	return &processTracker{
		actionQueue: make(chan action, 100),
	}
}

// Start launches the background monitor goroutine.
// Subsequent calls are no-ops.
func (pr *processTracker) Start(ctx context.Context) {
	pr.startOnce.Do(func() {
		go pr.monitor(ctx)
	})
}

// Register adds a process to the tracker and returns an ID for later removal.
// It returns an error if the context is canceled before the action is queued.
func (pr *processTracker) Register(ctx context.Context, proc Killable) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	id := uuid.NewString()

	select {
	case pr.actionQueue <- action{
		id:   id,
		proc: proc,
		op:   opAdd,
	}:
		return id, nil

	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// Unregister removes a process from the tracker by its ID.
// It returns an error if the context is canceled before the action is queued.
func (pr *processTracker) Unregister(ctx context.Context, id string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	select {
	case pr.actionQueue <- action{
		id: id,
		op: opRemove,
	}:
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}

func (pr *processTracker) monitor(ctx context.Context) {
	processes := make(map[string]Killable)

	for {
		select {
		case <-ctx.Done():
			for _, p := range processes {
				_ = p.Kill()
			}
			return

		case act := <-pr.actionQueue:
			switch act.op {
			case opAdd:
				processes[act.id] = act.proc
			case opRemove:
				delete(processes, act.id)
			}
		}
	}
}
