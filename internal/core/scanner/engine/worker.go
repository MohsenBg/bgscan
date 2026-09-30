package engine

import (
	"context"
	"errors"
	"io"
	"net/netip"
	"sync"
	"sync/atomic"

	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
)

type errStats struct {
	timeout     atomic.Uint64
	unreachable atomic.Uint64
	refused     atomic.Uint64
	badResponse atomic.Uint64
	closed      atomic.Uint64
	other       atomic.Uint64
}

// count categorizes err and increments the corresponding counter.
// It reports whether the error represents an expected probe failure.
func (s *errStats) count(err error) bool {
	switch {
	case errors.Is(err, probe.ErrTimeout):
		s.timeout.Add(1)
	case errors.Is(err, probe.ErrUnreachable):
		s.unreachable.Add(1)
	case errors.Is(err, probe.ErrRefused):
		s.refused.Add(1)
	case errors.Is(err, probe.ErrBadResponse):
		s.badResponse.Add(1)
	case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
		s.closed.Add(1)
	default:
		s.other.Add(1)
		return false
	}
	return true
}

func (s *errStats) snapshot() ErrorStats {
	return ErrorStats{
		Timeout:     s.timeout.Load(),
		Unreachable: s.unreachable.Load(),
		Refused:     s.refused.Load(),
		BadResponse: s.badResponse.Load(),
		Closed:      s.closed.Load(),
		Other:       s.other.Load(),
	}
}

func (e *stageExecutor) logIPError(ip netip.Addr, err error) {
	if errors.Is(err, context.Canceled) || errors.Is(err, ErrResultLimitReached) {
		return
	}

	if e.errs.count(err) {
		e.log.Debug("%s: %v", ip, err)
		return
	}
	e.log.Error("%s: %v", ip, err)
}

type ipProcessor func(netip.Addr) error

func (e *stageExecutor) workerCount() int {
	return getWorkerCount(e.stage.Workers)
}

func (e *stageExecutor) runPool(ctx context.Context, input <-chan netip.Addr, process ipProcessor) {
	workers := e.workerCount()

	var wg sync.WaitGroup
	wg.Add(workers)

	for range workers {
		go func() {
			defer wg.Done()
			e.runWorker(ctx, input, process)
		}()
	}

	wg.Wait()
}

func (e *stageExecutor) runWorker(ctx context.Context, input <-chan netip.Addr, process ipProcessor) {
	for {
		if e.pause != nil && !e.pause.Wait(ctx) {
			return
		}

		select {
		case <-ctx.Done():
			return

		case ip, ok := <-input:
			if !ok {
				return
			}
			if err := process(ip); err != nil {
				e.logIPError(ip, err)
			}
		}
	}
}

func getWorkerCount(workers int) int {
	if workers <= 0 {
		return 1
	}
	return workers
}
