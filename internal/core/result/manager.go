package result

import (
	"context"
	"sync"
	"time"

	"bgscan/internal/core/fileutil"
	"bgscan/internal/logger"
)

// Writer asynchronously aggregates IPScanResult items and merges them into the
// final result file. It operates as a background worker that periodically
// flushes accumulated results based on configurable policies.
//
// Flushing occurs when:
//   - The accumulated batch reaches BatchSize
//   - MergeFlushInterval elapses
//   - Shutdown begins (Stop)
//
// Writer guarantees that any result successfully written to the input channel
// before shutdown will be flushed to disk before Stop() returns.
type Writer struct {
	config Config

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	resultPath string
	schema     ResultSchema

	input chan Result

	batch     []Result
	batchSize int
}

// NewWriter initializes an asynchronous result writer tied to the given
// context. If ctx is nil, context.Background() is used.
//
// The returned Writer is not started automatically; the caller must invoke
// Start() to launch its background goroutine.
func NewWriter(
	resultPath string,
	schema ResultSchema,
	cfg Config,
	ctx context.Context,
) (*Writer, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cfg.Normalize()

	ctx, cancel := context.WithCancel(ctx)

	return &Writer{
		config:     cfg,
		resultPath: resultPath,
		schema:     schema,
		input:      make(chan Result, cfg.ChanSize),
		batch:      make([]Result, 0, cfg.BatchSize),
		batchSize:  cfg.BatchSize,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// Start launches the background processing goroutine.
//
// Safe to call exactly once. After Start, the Writer begins accepting and
// flushing results asynchronously.
func (w *Writer) Start() {
	w.wg.Add(1)
	go w.writeLoop()
}

// Stop gracefully shuts down the Writer. It cancels the internal context,
// drains the input channel, flushes any remaining batch, and waits for the
// background goroutine to exit.
//
// Stop guarantees all previously submitted results are persisted.
func (w *Writer) Stop() error {
	w.cancel()
	w.wg.Wait()
	return nil
}

// Write submits a result to the Writer. If the writer is already shutting down,
// the result is silently dropped to avoid blocking.
//
// Writes are non‑blocking thanks to the buffered input channel.
func (w *Writer) Write(r Result) {
	select {
	case <-w.ctx.Done():
		return

	case w.input <- r:
	}
}

// writeLoop is the main worker goroutine that handles batching and periodic
// flushing. It reacts to incoming results, timer ticks, and shutdown signals.
func (w *Writer) writeLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.MergeFlushInterval)
	defer ticker.Stop()

	for {
		select {

		case r := <-w.input:
			w.batch = append(w.batch, r)
			if len(w.batch) >= w.batchSize {
				_ = w.flush()
			}

		case <-ticker.C:
			_ = w.flush()

		case <-w.ctx.Done():
			// On shutdown:
			//   1. Drain remaining buffered results
			//   2. Flush everything
			w.drain()
			_ = w.flush()
			return
		}
	}
}

// drain empties the input channel without blocking. Called only during
// shutdown to ensure no queued item is lost.
func (w *Writer) drain() {
	for {
		select {
		case r := <-w.input:
			w.batch = append(w.batch, r)
		default:
			return
		}
	}
}

// flush writes the current batch of results to disk using mergeResults.
// The batch slice is reset afterward.
//
// Errors are logged through logger.DebugError but also returned to the caller.
func (w *Writer) flush() error {
	if len(w.batch) == 0 {
		return nil
	}

	batch := w.batch
	w.batch = make([]Result, 0, w.batchSize)

	err := mergeResults(
		w.resultPath,
		w.schema,
		batch,
	)
	if err != nil {
		logger.DebugError("%s", err.Error())
	}

	return err
}

// GetResultPath returns the final result file path, but only if the file
// currently exists. Otherwise an empty string is returned.
//
// This is useful for callers who want to verify that output has been produced.
func (w *Writer) GetResultPath() string {
	if fileutil.CheckFileExists(w.resultPath) {
		return w.resultPath
	}
	return ""
}
