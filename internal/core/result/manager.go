package result

import (
	"context"
	"os"
	"sync"
	"time"

	"bgscan/internal/core/fileutil"
	"bgscan/internal/logger"
)

// Writer asynchronously batches and flushes IPScanResult items to a result file.
type Writer struct {
	config     Config
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	resultPath string
	schema     ResultSchema
	input      chan Result
	batch      []Result
	batchSize  int
	startOnce  sync.Once
	stopOnce   sync.Once
}

// NewWriter creates a Writer tied to the given context (defaults to Background if nil).
func NewWriter(resultPath string, schema ResultSchema, cfg Config, ctx context.Context) (*Writer, error) {
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

// Start ensures the result directory exists and removes any stale result file.
func (w *Writer) Start() error {
	var err error
	w.startOnce.Do(func() {
		if err = fileutil.EnsureDir(w.resultPath); err != nil {
			logger.DebugError("failed to ensure directory: %v", err)
			return
		}
		if fileutil.CheckFileExists(w.resultPath) {
			if err = os.Remove(w.resultPath); err != nil {
				logger.CoreError("failed to remove stale result file: %v", err)
				return
			}
		}
		w.wg.Add(1)
		go w.writeLoop()
	})
	return err
}

// Stop cancels the writer, drains remaining items, and waits for the loop to exit.
func (w *Writer) Stop() error {
	w.stopOnce.Do(func() {
		w.cancel()
		w.wg.Wait()
	})
	return nil
}

// Write enqueues a result, dropping it if the context is already canceled.
func (w *Writer) Write(r Result) {
	select {
	case <-w.ctx.Done():
		return
	case w.input <- r:
	}
}

func (w *Writer) writeLoop() {
	defer w.wg.Done()
	ticker := time.NewTicker(w.config.MergeFlushInterval)
	defer ticker.Stop()

	for {
		select {
		case r, ok := <-w.input:
			if !ok {
				w.flush()
				return
			}
			w.batch = append(w.batch, r)
			if len(w.batch) >= w.batchSize {
				w.flush()
			}
		case <-ticker.C:
			w.flush()
		case <-w.ctx.Done():
			w.drain()
			w.flush()
			return
		}
	}
}

func (w *Writer) drain() {
	for {
		select {
		case r, ok := <-w.input:
			if !ok {
				return
			}
			w.batch = append(w.batch, r)
		default:
			return
		}
	}
}

func (w *Writer) flush() {
	if len(w.batch) == 0 {
		return
	}
	tmp := make([]Result, len(w.batch))
	copy(tmp, w.batch)
	w.batch = w.batch[:0]

	if err := mergeResults(w.resultPath, w.schema, tmp); err != nil {
		logger.DebugError("failed to flush results: %v", err)
	}
}

// GetResultPath returns the result file path if the file exists, otherwise empty string.
func (w *Writer) GetResultPath() string {
	if fileutil.CheckFileExists(w.resultPath) {
		return w.resultPath
	}
	return ""
}
