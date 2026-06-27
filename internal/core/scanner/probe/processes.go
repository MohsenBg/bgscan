package probe

import (
	"context"
	"sync"

	"bgscan/internal/core/process"

	"github.com/google/uuid"
)

type opType uint8

const (
	opAdd opType = iota
	opRemove
	opShutdown
)

type action struct {
	id   string
	proc *process.Process
	op   opType
}

type ProcessRegistry struct {
	actionQueue chan action
	startOnce   sync.Once
}

func NewProcessRegistry() *ProcessRegistry {
	return &ProcessRegistry{
		actionQueue: make(chan action, 100),
	}
}

func (pr *ProcessRegistry) Start(ctx context.Context) {
	pr.startOnce.Do(func() {
		go pr.monitor(ctx)
	})
}

func (pr *ProcessRegistry) Register(ctx context.Context, proc *process.Process) (string, error) {
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

func (pr *ProcessRegistry) Unregister(ctx context.Context, id string) error {
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

func (pr *ProcessRegistry) monitor(ctx context.Context) {
	processes := make(map[string]*process.Process)

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
