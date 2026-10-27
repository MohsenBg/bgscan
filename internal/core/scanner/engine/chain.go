package engine

import (
	"context"
	"errors"
	"net/netip"
	"sync"

	"github.com/MohsenBg/bgscan/internal/core/iplist"
)

const (
	defaultBatchSize    = 1000
	defaultStageChanBuf = 10_000
)

// RunScanWithChain executes a scan pipeline based on the configured chain mode.
func RunScanWithChain(ctx context.Context, input string, cfg ChainConfig) error {
	if len(cfg.Stages) == 0 {
		return errNoStages
	}

	// scan-level stop reaching MaxSuccessfulIPs cancels this context,
	ctx, stop := context.WithCancelCause(ctx)
	defer stop(nil)

	p := &pipeline{
		ctx:    ctx,
		cfg:    cfg,
		cancel: stop,
	}

	var err error
	switch cfg.Mode {
	case ModeSequential:
		err = p.runSequential(input)
	case ModeStreaming:
		err = p.runStreaming(input)
	case ModeBatch:
		err = p.runBatch(input)
	default:
		cfg.Log.Error("unknown chain mode: %v", cfg.Mode)
	}

	if errors.Is(context.Cause(p.ctx), ErrResultLimitReached) {
		cfg.Log.Info("scan finished early: result limit of %d reached", cfg.MaxSuccessfulIPs)
		return nil
	}

	if err != nil {
		return err
	}

	return p.stopCause()
}

// pipeline carries the shared state of a single chain run so stage helpers
// don't need long parameter lists.
type pipeline struct {
	ctx    context.Context
	cfg    ChainConfig
	cancel context.CancelCauseFunc
}

// stageErr converts executor errors into pipeline control flow. Reaching the
// result limit means the scan is done: stop the whole pipeline.
func (p *pipeline) stageErr(err error) error {
	switch {
	case errors.Is(err, ErrResultLimitReached):
		p.cancel(nil)
		return nil
	case errors.Is(err, ErrWriterFatal):
		p.cancel(ErrWriterFatal)
		return err
	default:
		return err
	}
}

// stopCause returns the pipeline's cancellation cause, treating an ordinary
// user cancellation as a clean stop rather than an error.
func (p *pipeline) stopCause() error {
	cause := context.Cause(p.ctx)
	if cause == nil || errors.Is(cause, context.Canceled) {
		return nil
	}
	return cause
}

// runSequential runs stages one after another using file-based outputs.
func (p *pipeline) runSequential(input string) error {
	current := input

	for i, stage := range p.cfg.Stages {
		if current == "" {
			p.cfg.Log.Info("stage %d skipped (no input)", i+1)
			return nil
		}
		if ctxDone(p.ctx) {
			return p.stopCause()
		}

		p.cfg.Log.Info("stage %d/%d starting", i+1, len(p.cfg.Stages))

		exec, err := newStageExecutor(p.ctx, stage, p.cfg, p.countInput(current), p.resultLimitFor(i))
		if err != nil {
			return err
		}

		runErr := p.streamStageFromFile(current, exec, nil, nil)
		runErr = errors.Join(runErr, exec.close())

		if runErr != nil {
			return runErr
		}

		current = stage.Writer.GetResultPath()
		p.cfg.Log.Info("stage %d/%d completed", i+1, len(p.cfg.Stages))
	}

	return nil
}

// runStreaming runs all stages concurrently as a streaming pipeline.
func (p *pipeline) runStreaming(input string) (err error) {
	total := p.countInput(input)
	p.cfg.Log.Info("stream pipeline started: stages=%d ips=%d", len(p.cfg.Stages), total)

	execs, err := p.buildExecutors(total)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, closeExecutors(execs)) }()

	channels := p.createStageChannels()

	var (
		wg       sync.WaitGroup
		errMu    sync.Mutex
		firstErr error
	)

	record := func(err error) {
		if err == nil {
			return
		}
		errMu.Lock()
		if firstErr == nil {
			firstErr = err
		}
		errMu.Unlock()
	}

	for i := range p.cfg.Stages {
		wg.Add(1)

		in := getInputChannel(i, channels)
		out := getOutputChannel(i, len(p.cfg.Stages), channels)

		go func() {
			defer wg.Done()
			defer closeOutputChannel(out)

			if in == nil {
				record(p.streamStageFromFile(input, execs[i], nextExec(execs, i), out))
			} else {
				record(p.streamStage(execs[i], nextExec(execs, i), in, out))
			}
		}()
	}

	wg.Wait()

	if firstErr != nil {
		return firstErr
	}
	return p.stopCause()
}

// runBatch runs the batch-based pipeline chain.
func (p *pipeline) runBatch(input string) (err error) {
	total := p.countInput(input)
	batchSize := p.batchSize()
	p.cfg.Log.Info("batch pipeline started: batch=%d ips=%d", batchSize, total)

	execs, err := p.buildExecutors(total)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, closeExecutors(execs)) }()

	for batch := range p.streamBatches(input, batchSize) {
		if ctxDone(p.ctx) {
			return p.stopCause()
		}
		if err := p.processBatch(batch, execs); err != nil {
			return err
		}
	}

	return p.stopCause()
}

// resultLimitFor returns the success cap for stage i. Only the final stage
// counts, since only its successes are the scan's finished results.
func (p *pipeline) resultLimitFor(i int) uint64 {
	if i == len(p.cfg.Stages)-1 {
		return p.cfg.MaxSuccessfulIPs
	}
	return 0
}

// buildExecutors creates executors for all stages. Only the first stage gets
// the initial IP count later stages accumulate totals as IPs are forwarded.
func (p *pipeline) buildExecutors(firstTotal uint64) ([]*stageExecutor, error) {
	execs := make([]*stageExecutor, 0, len(p.cfg.Stages))

	for i, stage := range p.cfg.Stages {
		var total uint64
		if i == 0 {
			total = firstTotal
		}

		exec, err := newStageExecutor(p.ctx, stage, p.cfg, total, p.resultLimitFor(i))
		if err != nil {
			return nil, errors.Join(err, closeExecutors(execs))
		}
		execs = append(execs, exec)
	}

	return execs, nil
}

// countInput returns the number of active IPs, logging (not propagating) failures.
func (p *pipeline) countInput(input string) uint64 {
	total, err := iplist.CountActiveIPs(input)
	if err != nil {
		p.cfg.Log.Error("failed to count IPs: %v", err)
	}
	return total
}

// createStageChannels creates buffered channels between pipeline stages.
func (p *pipeline) createStageChannels() []chan netip.Addr {
	channels := make([]chan netip.Addr, len(p.cfg.Stages))

	for i := range channels {
		size := p.cfg.MaxBuffer
		if size <= 0 {
			size = defaultStageChanBuf
		}
		if i+1 < len(p.cfg.Stages) {
			size = max(size, getWorkerCount(p.cfg.Stages[i+1].Workers))
		}
		channels[i] = make(chan netip.Addr, size)
	}

	return channels
}

// getInputChannel returns the input channel for a stage (nil for the first).
func getInputChannel(stageIdx int, channels []chan netip.Addr) chan netip.Addr {
	if stageIdx == 0 {
		return nil
	}
	return channels[stageIdx-1]
}

// getOutputChannel returns the output channel for a stage (nil for the last).
func getOutputChannel(stageIdx, total int, channels []chan netip.Addr) chan netip.Addr {
	if stageIdx >= total-1 {
		return nil
	}
	return channels[stageIdx]
}

func closeOutputChannel(ch chan netip.Addr) {
	if ch != nil {
		close(ch)
	}
}

// batchSize determines the batch size for batch mode, sized so downstream
// worker pools stay busy.
func (p *pipeline) batchSize() int {
	if p.cfg.BatchSize <= 0 {
		return defaultBatchSize
	}
	if len(p.cfg.Stages) <= 1 {
		return p.cfg.BatchSize
	}

	maxWorkers := 0
	for _, s := range p.cfg.Stages[1:] {
		maxWorkers = max(maxWorkers, s.Workers)
	}
	return max(maxWorkers, p.cfg.BatchSize)
}

// streamBatches streams IPs from the input file in fixed-size batches.
func (p *pipeline) streamBatches(input string, batchSize int) <-chan []netip.Addr {
	out := make(chan []netip.Addr, 2)

	go func() {
		defer close(out)

		ipCh := make(chan netip.Addr, batchSize*2)
		done := make(chan error, 1)

		go func() {
			defer close(ipCh)
			done <- iplist.StreamActiveIPs(p.cfg.Log, p.ctx, input, p.cfg.MaxIPsToTest, p.cfg.Shuffled, ipCh)
		}()

		batch := make([]netip.Addr, 0, batchSize)

		for ip := range ipCh {
			batch = append(batch, ip)

			if len(batch) == batchSize {
				select {
				case out <- batch:
					batch = make([]netip.Addr, 0, batchSize)
				case <-p.ctx.Done():
					return
				}
			}
		}

		if len(batch) > 0 {
			select {
			case out <- batch:
			case <-p.ctx.Done():
			}
		}

		if err := <-done; err != nil && !errors.Is(err, context.Canceled) {
			p.cfg.Log.Error("stream error: %v", err)
		}
	}()

	return out
}

// processBatch runs a single batch through all stages.
func (p *pipeline) processBatch(batch []netip.Addr, execs []*stageExecutor) error {
	current := batch

	for i, exec := range execs {
		if len(current) == 0 || ctxDone(p.ctx) {
			return p.stopCause()
		}

		var err error
		current, err = p.runStageBatch(current, exec)
		if err != nil {
			return err
		}

		if next := nextExec(execs, i); next != nil {
			next.total.Add(uint64(len(current)))
		}
	}

	return nil
}

// runStageBatch processes one batch through a single stage's worker pool and
// returns the IPs that passed.
func (p *pipeline) runStageBatch(batch []netip.Addr, exec *stageExecutor) ([]netip.Addr, error) {
	input := make(chan netip.Addr, exec.workerCount()*2)
	go func() {
		defer close(input)
		for _, ip := range batch {
			select {
			case input <- ip:
			case <-p.ctx.Done():
				return
			}
		}
	}()

	var (
		mu  sync.Mutex
		out = make([]netip.Addr, 0, len(batch))
	)

	err := exec.runPool(p.ctx, p.cancel, input, func(ip netip.Addr) error {
		if err := exec.processIP(p.ctx, ip); err != nil {
			return err // no classifier here
		}
		mu.Lock()
		out = append(out, ip)
		mu.Unlock()
		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

// streamStageFromFile runs a single stage reading IPs from a file. If out is
// non-nil, passing IPs are forwarded to the next stage.
func (p *pipeline) streamStageFromFile(input string, exec, next *stageExecutor, out chan netip.Addr) error {
	in := make(chan netip.Addr, exec.workerCount()*2)
	done := make(chan error, 1)

	go func() {
		defer close(in)
		done <- iplist.StreamActiveIPs(p.cfg.Log, p.ctx, input, p.cfg.MaxIPsToTest, p.cfg.Shuffled, in)
	}()

	runErr := p.streamStage(exec, next, in, out)

	// A read failure stops the stage; a cancellation caused by an earlier
	// pipeline stop is not an error of its own.
	readErr := <-done
	if errors.Is(readErr, context.Canceled) {
		readErr = nil
	}
	if readErr != nil {
		p.cfg.Log.Error("stream error: %v", readErr)
	}

	return errors.Join(runErr, readErr)
}

// streamStage runs a single stage over an input channel of IPs.
func (p *pipeline) streamStage(exec, next *stageExecutor, in, out chan netip.Addr) error {
	return exec.runPool(p.ctx, p.cancel, in, func(ip netip.Addr) error {
		if err := exec.processIP(p.ctx, ip); err != nil {
			return p.stageErr(err)
		}
		forward(p.ctx, out, next, ip)
		return nil
	})
}

// forward passes an IP to the next stage, counting it toward its total.
func forward(ctx context.Context, out chan netip.Addr, next *stageExecutor, ip netip.Addr) {
	if out == nil {
		return
	}
	select {
	case out <- ip:
		if next != nil {
			next.total.Add(1)
		}
	case <-ctx.Done():
	}
}

// nextExec returns the executor following index i, or nil for the last stage.
func nextExec(execs []*stageExecutor, i int) *stageExecutor {
	if i+1 < len(execs) {
		return execs[i+1]
	}
	return nil
}

func closeExecutors(execs []*stageExecutor) error {
	errs := make([]error, 0, len(execs))
	for _, e := range execs {
		errs = append(errs, e.close())
	}
	return errors.Join(errs...)
}

// ctxDone reports whether the context has been cancelled.
func ctxDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
