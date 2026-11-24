package result

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/config"
	"github.com/MohsenBg/bgscan/internal/core/config/validate"
	"github.com/MohsenBg/bgscan/internal/core/fileutil"
	"github.com/MohsenBg/bgscan/internal/core/target"
	"github.com/MohsenBg/bgscan/internal/logger"
)

var (
	// ErrWriterNotStarted is returned by Write before Start succeeds.
	ErrWriterNotStarted = errors.New("result: writer not started")

	// ErrWriterClosed is returned by Write after Close (or parent ctx cancel).
	ErrWriterClosed = errors.New("result: writer closed")
)

// WriterStats is a point-in-time snapshot of writer counters.
type WriterStats struct {
	Written uint64 // results persisted successfully
	Dropped uint64 // results rejected (not started, closed, or ctx cancelled)
	Flushes uint64 // successful merge+write cycles
}

type Writer interface {
	// Start prepares the result directory and starts the background loop.
	// Safe to call multiple times; the first call's error is cached.
	Start() error

	// Write enqueues a result. Returns ErrWriterNotStarted before Start and
	// ErrWriterClosed after Close. Returns ctx.Err() if the parent context
	// fired before the result could be enqueued.
	Write(r Result) error

	// Close signals the loop to drain and flush, waits for it to exit, and
	// returns the first error encountered during flushing (if any).
	Close() error

	// GetResultPath returns the result file path if it exists, else "".
	GetResultPath() string

	// Stats returns current counters.
	Stats() WriterStats
}

type writer struct {
	log    *logger.Logger
	config config.WriterConfig
	schema ResultSchema
	prefix string

	// Result file path. Decided lazily at the first flush, from the first
	// result's KeyType: a run is all IPs or all domains, and the kind is
	// only known once a result arrives. Guarded by pathMu because the loop
	// goroutine writes it while GetResultPath may read it from another
	// goroutine.
	pathMu sync.RWMutex
	path   string

	// Lifecycle: parent ctx -> w.ctx, cancelled on shutdown.
	ctx    context.Context
	cancel context.CancelFunc

	// Inputs and loop termination.
	input  chan Result
	exited chan struct{}

	// Batch state (owned by loop goroutine only).
	batch []Result

	// Start / Close races.
	startOnce sync.Once
	startErr  error
	started   atomic.Bool

	closeOnce sync.Once
	closeErr  error

	// Write / Close mutual exclusion. Write holds RLock across the send so
	// Close cannot close the channel mid-send.
	stateMu sync.RWMutex
	closed  bool

	// First flush error, reported by Close.
	errMu    sync.Mutex
	firstErr error

	// Stats.
	statWritten atomic.Uint64
	statDropped atomic.Uint64
	statFlushes atomic.Uint64
}

// WriterOptions configures a Writer.
type WriterOptions struct {
	ResultPrefix string
	Schema       ResultSchema
	Config       config.WriterConfig
	Log          *logger.Logger
}

// NewWriter creates a Writer tied to the given context.
//
// The result file path is not built here: whether results land under the
// schema's ips or domains directory is only known when the first result
// arrives, so the path is decided at the first flush (see ensurePath).
func NewWriter(ctx context.Context, opts WriterOptions) (Writer, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if errs := validate.ValidateWriter(opts.Config); len(errs) > 0 {
		return nil, errors.Join(slices.Collect(maps.Values(errs))...)
	}
	if err := opts.Schema.Validate(); err != nil {
		return nil, err
	}
	if opts.ResultPrefix == "" {
		return nil, errors.New("result: prefix cannot be empty")
	}

	wctx, cancel := context.WithCancel(ctx)
	return &writer{
		log:    opts.Log,
		config: opts.Config,
		schema: opts.Schema,
		prefix: opts.ResultPrefix,
		ctx:    wctx,
		cancel: cancel,
		input:  make(chan Result, opts.Config.ChanSize),
		exited: make(chan struct{}),
		batch:  make([]Result, 0, opts.Config.BatchSize),
	}, nil
}

// --- Lifecycle ---

// Start starts the writer background loop.
func (w *writer) Start() error {
	w.startOnce.Do(func() {
		w.startErr = w.doStart()
		if w.startErr == nil {
			w.started.Store(true)
		}
	})
	return w.startErr
}

func (w *writer) doStart() error {
	// Nothing to prepare on disk yet: the result path (and its directory)
	// is only decided at the first flush, when the first result reveals
	// whether this run is IPs or domains.
	go w.loop()
	return nil
}

// Write queues one result for the background loop.
func (w *writer) Write(r Result) error {
	if !w.started.Load() {
		w.statDropped.Add(1)
		return ErrWriterNotStarted
	}
	if err := w.ctx.Err(); err != nil {
		w.statDropped.Add(1)
		return err
	}

	w.stateMu.RLock()
	defer w.stateMu.RUnlock()

	if w.closed {
		w.statDropped.Add(1)
		return ErrWriterClosed
	}

	select {
	case w.input <- r:
		return nil
	case <-w.ctx.Done():
		w.statDropped.Add(1)
		return w.ctx.Err()
	}
}

// Close drains and flushes queued results, then stops the loop.
func (w *writer) Close() error {
	w.closeOnce.Do(func() {
		// Prevent new writes, then close the channel so the loop drains.
		// Write holds RLock across its send, so this cannot race a send.
		w.stateMu.Lock()
		w.closed = true
		close(w.input)
		w.stateMu.Unlock()

		if w.started.Load() {
			<-w.exited // loop flushed and exited
		} else {
			// Start never ran: no loop to wait for.
			w.cancel()
		}

		w.closeErr = w.flushError()
	})
	return w.closeErr
}

// --- Background loop ---

func (w *writer) loop() {
	defer close(w.exited)
	defer w.cancel()

	timer := time.NewTimer(w.config.MergeFlushInterval.Duration())
	defer timer.Stop()

	resetTimer := func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(w.config.MergeFlushInterval.Duration())
	}

	for {
		select {
		case r, ok := <-w.input:
			if !ok {
				// Close was called; final flush.
				w.flush()
				return
			}
			w.batch = append(w.batch, r)
			if len(w.batch) >= w.config.BatchSize {
				if w.flush() {
					// Fatal flush error: stop accepting input and drain.
					w.cancel()
				} else {
					resetTimer()
				}
			}

		case <-timer.C:
			if w.flush() {
				w.cancel()
			} else {
				resetTimer()
			}

		case <-w.ctx.Done():
			w.drainIntoBatch()
			w.flush()
			return
		}
	}
}

// drainIntoBatch pulls everything currently buffered in w.input into the
// batch. Called only from the loop goroutine on shutdown.
func (w *writer) drainIntoBatch() {
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

// --- Flushing ---

// flush writes the current batch to disk. Returns true on fatal error.
func (w *writer) flush() (fatal bool) {
	if len(w.batch) == 0 {
		return false
	}

	// Clone the batch: mergeResults sorts its argument in place.
	tmp := slices.Clone(w.batch)
	n := len(tmp)
	w.batch = w.batch[:0]

	// The first result of the first batch decides the file path for the
	// whole run: a run is all IPs or all domains, never mixed.
	if err := w.ensurePath(tmp[0].KeyType()); err != nil {
		w.log.Error("prepare result path: %v", err)
		if w.config.IgnoreErr {
			return false
		}
		w.recordErr(err)
		return true
	}

	if err := mergeResults(w.log, w.resultPath(), w.schema, tmp); err != nil {
		w.log.Error("flush results: %v", err)
		if w.config.IgnoreErr {
			return false
		}
		w.recordErr(err)
		return true
	}

	w.statFlushes.Add(1)
	w.statWritten.Add(uint64(n))
	return false
}

// ensurePath builds the result file path the first time it is needed,
// placing the file under the schema directory's ips or domains
// subdirectory according to kind.
func (w *writer) ensurePath(kind target.Kind) error {
	if w.resultPath() != "" {
		return nil
	}

	schema := w.schema
	schema.Directory = filepath.Join(schema.Directory, kind.Dir())

	path, err := prepareResultFilePath(w.config, schema, w.prefix)
	if err != nil {
		return err
	}
	if err := fileutil.RemoveFile(path); err != nil {
		return fmt.Errorf("remove stale result file: %w", err)
	}

	w.pathMu.Lock()
	w.path = path
	w.pathMu.Unlock()
	return nil
}

func (w *writer) resultPath() string {
	w.pathMu.RLock()
	defer w.pathMu.RUnlock()
	return w.path
}

func (w *writer) recordErr(err error) {
	w.errMu.Lock()
	defer w.errMu.Unlock()
	if w.firstErr == nil {
		w.firstErr = err
	}
}

// flushError returns the first flush error seen, or nil.
func (w *writer) flushError() error {
	w.errMu.Lock()
	defer w.errMu.Unlock()
	return w.firstErr
}

// --- Accessors ---

// GetResultPath returns the result file path, or an empty string.
func (w *writer) GetResultPath() string {
	if !w.started.Load() {
		return ""
	}
	path := w.resultPath()
	if path == "" {
		return ""
	}
	if fileutil.FileExists(path) {
		return path
	}
	return ""
}

// Stats returns a snapshot of the writer counters.
func (w *writer) Stats() WriterStats {
	return WriterStats{
		Written: w.statWritten.Load(),
		Dropped: w.statDropped.Load(),
		Flushes: w.statFlushes.Load(),
	}
}
