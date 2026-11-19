package engine

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"

	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/core/target"
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

func (e *stageExecutor) logIPError(entry target.Entry, err error) {
	if errors.Is(err, context.Canceled) || errors.Is(err, ErrResultLimitReached) {
		return
	}

	if e.errs.count(err) {
		e.log.Debug("%s: %v", entry.Value(), err)
		return
	}
	e.log.Error("%s: %v", entry.Value(), err)
}

type entryProcessor func(target.Entry) error

func (e *stageExecutor) workerCount() int {
	return getWorkerCount(e.stage.Workers)
}

func (e *stageExecutor) runPool(
	ctx context.Context,
	cancel context.CancelCauseFunc,
	input <-chan target.Entry,
	process entryProcessor,
) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	done := make(chan struct{})
	go func() {
		select {
		case err := <-errCh:
			cancel(err)
		case <-ctx.Done():
		case <-done:
		}
	}()

	wg.Add(e.workerCount())
	for range e.workerCount() {
		go func() {
			defer wg.Done()
			if err := e.runWorker(ctx, input, process); err != nil {
				select {
				case errCh <- err:
				default: // another worker reported first
				}
			}
		}()
	}

	wg.Wait()
	close(done)

	if cause := context.Cause(ctx); cause != nil && !errors.Is(cause, context.Canceled) {
		return cause
	}
	return nil
}

func (e *stageExecutor) runWorker(ctx context.Context, input <-chan target.Entry, process entryProcessor) error {
	for {
		if e.pause != nil && !e.pause.Wait(ctx) {
			return nil
		}

		select {
		case <-ctx.Done():
			return nil

		case entry, ok := <-input:
			if !ok {
				return nil
			}
			if err := process(entry); err != nil {
				if terminating(err) {
					return err
				}
				e.logIPError(entry, err)
			}
		}
	}
}

// terminating reports whether err ends the whole scan (vs. a single IP failing).
func terminating(err error) bool {
	return errors.Is(err, ErrResultLimitReached) ||
		errors.Is(err, ErrWriterFatal)
}

func getWorkerCount(workers int) int {
	if workers <= 0 {
		return 1
	}
	return workers
}
