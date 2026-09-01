package result

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/core/config/validate"
	"github.com/MohsenBg/bgscan/internal/core/fileutil"
	"github.com/MohsenBg/bgscan/internal/logger"
)

// writer asynchronously batches and flushes Result items to a result file.
type writer struct {
	config     config.WriterConfig
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

// Writer batches Result items and persists them to a result file.
type Writer interface {
	Start() error
	Stop() error
	Write(r Result)
	GetResultPath() string
}

// WriterOptions configures a Writer.
type WriterOptions struct {
	ResultPrefix string
	Schema       ResultSchema
	Config       config.WriterConfig
}

// NewWriter creates a Writer tied to the given context (defaults to Background if nil).
func NewWriter(ctx context.Context, opts WriterOptions) (Writer, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	errs := validate.ValidateWriter(opts.Config)
	for _, err := range errs {
		return nil, err
	}

	path, err := prepareResultFilePath(opts.Config, opts.Schema, opts.ResultPrefix)
	if err != nil {
		return nil, err
	}

	err = opts.Schema.Validate()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	return &writer{
		config:     opts.Config,
		resultPath: path,
		schema:     opts.Schema,
		input:      make(chan Result, opts.Config.ChanSize),
		batch:      make([]Result, 0, opts.Config.BatchSize),
		batchSize:  opts.Config.BatchSize,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// Start ensures the result directory exists and removes any stale result file.
func (w *writer) Start() error {
	var err error
	w.startOnce.Do(func() {
		if err = fileutil.EnsureFileDir(w.resultPath); err != nil {
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
func (w *writer) Stop() error {
	w.stopOnce.Do(func() {
		w.cancel()
		w.wg.Wait()
	})
	return nil
}

// Write enqueues a result, dropping it if the context is already canceled.
func (w *writer) Write(r Result) {
	select {
	case <-w.ctx.Done():
		return
	case w.input <- r:
	}
}

func (w *writer) writeLoop() {
	defer w.wg.Done()
	ticker := time.NewTicker(w.config.MergeFlushInterval.Duration())
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

func (w *writer) drain() {
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

func (w *writer) flush() {
	if len(w.batch) == 0 {
		return
	}
	tmp := make([]Result, len(w.batch))
	copy(tmp, w.batch)
	w.batch = w.batch[:0]

	if err := mergeResults(w.resultPath, w.batchSize, w.schema, tmp); err != nil {
		logger.DebugError("failed to flush results: %v", err)
	}
}

// GetResultPath returns the result file path if the file exists, otherwise empty string.
func (w *writer) GetResultPath() string {
	if fileutil.CheckFileExists(w.resultPath) {
		return w.resultPath
	}
	return ""
}
