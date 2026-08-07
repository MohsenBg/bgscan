package engine

import (
	"context"
	"net/netip"
	"sync"

	"bgscan/internal/core/iplist"
	"bgscan/internal/logger"
)

const (
	defaultBatchSize    = 1000
	defaultStageChanBuf = 10_000
)

// RunScanWithChain executes a scan pipeline based on the configured chain mode.
func RunScanWithChain(ctx context.Context, input string, cfg ChainConfig) {
	if len(cfg.Stages) == 0 {
		return
	}

	switch cfg.Mode {
	case ModeSequential:
		executeSequentialChain(ctx, input, cfg)
	case ModeStreaming:
		executeStreamingPipeline(ctx, input, cfg)
	case ModeBatch:
		executeBatchPipeline(ctx, input, cfg)
	}
}

// executeSequentialChain runs stages one after another using file-based outputs.
func executeSequentialChain(ctx context.Context, input string, cfg ChainConfig) {
	currentInput := input

	for i, stage := range cfg.Stages {
		if currentInput == "" {
			logger.CoreInfo("stage %d skipped (no input)", i+1)
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		logger.CoreInfo("stage %d/%d starting", i+1, len(cfg.Stages))
		RunScan(ctx, currentInput, ScanConfig{
			Workers:          stage.Workers,
			MaxIPsToTest:     cfg.MaxIPsToTest,
			Probe:            stage.Probe,
			Writer:           stage.Writer,
			MinProbeDuration: cfg.MinProbeDuration,
			ProgressInterval: stage.ProgressInterval,
			Hooks:            stage.Hooks,
			Shuffled:         cfg.Shuffled,
			Pause:            cfg.Pause,
			RateLimiter:      cfg.RateLimiter,
		})
		currentInput = stage.Writer.GetResultPath()
		logger.CoreInfo("stage %d/%d completed", i+1, len(cfg.Stages))
	}
}

// executeStreamingPipeline runs all stages concurrently in a streaming pipeline.
func executeStreamingPipeline(ctx context.Context, input string, cfg ChainConfig) {
	totalIPs, err := iplist.CountActiveIPs(input)
	if err != nil {
		logger.CoreError("failed to count IPs: %v", err)
		totalIPs = 0
	}

	logger.CoreInfo("stream pipeline started: stages=%d ips=%d", len(cfg.Stages), totalIPs)

	channels := createStageChannels(cfg)
	executors := make([]*stageExecutor, 0, len(cfg.Stages))

	for i, stage := range cfg.Stages {
		var total uint64
		if i == 0 {
			total = totalIPs
		}

		exec, err := newStageExecutor(ctx, stage, cfg, total)
		if err != nil {
			stage.Hooks.callOnError(err)
			return
		}

		executors = append(executors, exec)
	}

	defer func() {
		for _, e := range executors {
			e.cleanup()
		}
	}()

	var wg sync.WaitGroup

	for i, stage := range cfg.Stages {
		wg.Add(1)

		in := getInputChannel(i, channels)
		out := getOutputChannel(i, len(cfg.Stages), channels)

		var next *stageExecutor
		if i+1 < len(executors) {
			next = executors[i+1]
		}

		go func(idx int, s StageConfig, in, out chan netip.Addr, exec, nextExec *stageExecutor) {
			defer wg.Done()
			defer closeOutputChannel(out)

			if in == nil {
				streamStageFromFile(ctx, input, cfg.MaxIPsToTest, s, cfg.Shuffled, out, exec, nextExec, cfg.Pause)
			} else {
				streamStageFromChannel(ctx, in, s, out, exec, nextExec, cfg.Pause)
			}
		}(i, stage, in, out, executors[i], next)
	}

	wg.Wait()
}

// createStageChannels creates buffered channels between pipeline stages.
func createStageChannels(cfg ChainConfig) []chan netip.Addr {
	channels := make([]chan netip.Addr, len(cfg.Stages))

	for i := range channels {
		size := cfg.MaxBuffer
		if size <= 0 {
			size = defaultStageChanBuf
		}
		if i+1 < len(cfg.Stages) {
			size = max(size, getWorkerCount(cfg.Stages[i+1].Workers))
		}
		channels[i] = make(chan netip.Addr, size)
	}

	return channels
}

// getInputChannel returns the input channel for a stage.
func getInputChannel(stageIdx int, channels []chan netip.Addr) chan netip.Addr {
	if stageIdx == 0 {
		return nil
	}
	return channels[stageIdx-1]
}

// getOutputChannel returns the output channel for a stage.
func getOutputChannel(stageIdx, total int, channels []chan netip.Addr) chan netip.Addr {
	if stageIdx >= total-1 {
		return nil
	}
	return channels[stageIdx]
}

// closeOutputChannel closes a channel safely.
func closeOutputChannel(ch chan netip.Addr) {
	if ch != nil {
		close(ch)
	}
}

// executeBatchPipeline runs the batch-based pipeline chain.
func executeBatchPipeline(ctx context.Context, input string, cfg ChainConfig) {
	totalIPs, err := iplist.CountActiveIPs(input)
	if err != nil {
		logger.CoreError("failed to count IPs: %v", err)
		totalIPs = 0
	}

	batchSize := calculateBatchSize(cfg)
	logger.CoreInfo("batch pipeline started: batch=%d ips=%d", batchSize, totalIPs)

	stream := streamIPsFromFile(ctx, input, cfg.Shuffled, cfg.MaxIPsToTest, batchSize)

	executors := make([]*stageExecutor, 0, len(cfg.Stages))

	defer func() {
		for _, e := range executors {
			e.cleanup()
		}
	}()

	for i, stage := range cfg.Stages {
		var total uint64
		if i == 0 {
			total = totalIPs
		}

		exec, err := newStageExecutor(ctx, stage, cfg, total)
		if err != nil {
			stage.Hooks.callOnError(err)
			return
		}

		executors = append(executors, exec)
	}

	for batch := range stream {
		select {
		case <-ctx.Done():
			return
		default:
		}

		processBatch(ctx, batch, executors, cfg.Pause)
	}
}

// processBatch runs a single batch through all stages.
func processBatch(ctx context.Context, batch []netip.Addr, execs []*stageExecutor, pause PauseController) {
	current := batch

	for i, exec := range execs {
		if len(current) == 0 {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		current = executeBatch(ctx, current, exec, pause)

		if i+1 < len(execs) {
			execs[i+1].total.Add(uint64(len(current)))
		}
	}
}

// executeBatch processes a batch in worker pool
func executeBatch(ctx context.Context, batch []netip.Addr, exec *stageExecutor, pause PauseController) []netip.Addr {
	workers := getWorkerCount(exec.stage.Workers)
	input := make(chan netip.Addr, workers*2)
	go func() {
		defer close(input)
		for _, ip := range batch {
			select {
			case input <- ip:
			case <-ctx.Done():
				return
			}
		}
	}()

	var (
		mu  sync.Mutex
		out = make([]netip.Addr, 0, len(batch))
	)

	runWorkerPool(ctx, workers, pause, input, func(ip netip.Addr) {
		if exec.processIP(ctx, ip) {
			mu.Lock()
			out = append(out, ip)
			mu.Unlock()
		}
	})

	return out
}

// calculateBatchSize determines optimal batch size for pipeline mode.
func calculateBatchSize(cfg ChainConfig) int {
	if cfg.BatchSize <= 0 {
		return defaultBatchSize
	}

	if len(cfg.Stages) <= 1 {
		return cfg.BatchSize
	}

	maxWorkers := 0
	for _, s := range cfg.Stages[1:] {
		if s.Workers > maxWorkers {
			maxWorkers = s.Workers
		}
	}

	return max(maxWorkers, cfg.BatchSize)
}

// streamIPsFromFile streams IPs in batches from file input.
func streamIPsFromFile(ctx context.Context, input string, shuffled bool, maxIP uint64, batchSize int) <-chan []netip.Addr {
	out := make(chan []netip.Addr, 2)

	go func() {
		defer close(out)

		ipCh := make(chan netip.Addr, batchSize*2)
		done := make(chan error, 1)

		go func() {
			defer close(ipCh)
			done <- iplist.StreamActiveIPs(ctx, input, maxIP, shuffled, ipCh)
		}()

		batch := make([]netip.Addr, 0, batchSize)

		for ip := range ipCh {
			batch = append(batch, ip)

			if len(batch) >= batchSize {
				select {
				case out <- batch:
					batch = make([]netip.Addr, 0, batchSize)
				case <-ctx.Done():
					return
				}
			}
		}

		if len(batch) > 0 {
			select {
			case out <- batch:
			case <-ctx.Done():
			}
		}

		if err := <-done; err != nil && err != context.Canceled {
			logger.CoreError("stream error: %v", err)
		}
	}()

	return out
}

func streamStageFromFile(
	ctx context.Context,
	input string,
	maxIP uint64,
	stage StageConfig,
	shuffled bool,
	output chan netip.Addr,
	exec *stageExecutor,
	next *stageExecutor,
	pause PauseController,
) {
	workers := getWorkerCount(stage.Workers)
	in := make(chan netip.Addr, workers*2)

	done := make(chan error, 1)

	go func() {
		defer close(in)
		done <- iplist.StreamActiveIPs(ctx, input, maxIP, shuffled, in)
	}()

	runWorkerPool(ctx, workers, pause, in, func(ip netip.Addr) {
		if exec.processIP(ctx, ip) && output != nil {
			select {
			case output <- ip:
				if next != nil {
					next.total.Add(1)
				}
			case <-ctx.Done():
			}
		}
	})

	if err := <-done; err != nil && err != context.Canceled {
		logger.CoreError("stream error: %v", err)
		stage.Hooks.callOnError(err)
	}
}

func streamStageFromChannel(
	ctx context.Context,
	input chan netip.Addr,
	stage StageConfig,
	output chan netip.Addr,
	exec *stageExecutor,
	next *stageExecutor,
	pause PauseController,
) {
	workers := getWorkerCount(stage.Workers)

	runWorkerPool(ctx, workers, pause, input, func(ip netip.Addr) {
		if exec.processIP(ctx, ip) && output != nil {
			select {
			case output <- ip:
				if next != nil {
					next.total.Add(1)
				}
			case <-ctx.Done():
			}
		}
	})
}
