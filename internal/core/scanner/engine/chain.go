package engine

import (
	"context"
	"sync"

	"bgscan/internal/core/config"
	"bgscan/internal/core/iplist"
	"bgscan/internal/logger"
)

const (
	defaultBatchSize    = 1000
	defaultStageChanBuf = 10_000
)

// RunScanWithChain executes a scan pipeline based on the configured chain mode.
func RunScanWithChain(ctx context.Context, input string, maxIP int, cfg *ChainConfig) {
	if len(cfg.Stages) == 0 {
		return
	}

	switch cfg.Mode {
	case ModeSequential:
		executeSequentialChain(ctx, input, maxIP, cfg)
	case ModeStreaming:
		executeStreamingPipeline(ctx, input, maxIP, cfg)
	case ModeBatch:
		executeBatchPipeline(ctx, input, maxIP, cfg)
	}
}

// executeSequentialChain runs stages one after another using file-based outputs.
func executeSequentialChain(ctx context.Context, input string, maxIP int, cfg *ChainConfig) {
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
		RunScan(ctx, currentInput, maxIP, stage, cfg.Shuffled, cfg.Pause)
		currentInput = stage.Writer.GetResultPath()
		logger.CoreInfo("stage %d/%d completed", i+1, len(cfg.Stages))
	}
}

// executeStreamingPipeline runs all stages concurrently in a streaming pipeline.
func executeStreamingPipeline(ctx context.Context, input string, maxIP int, cfg *ChainConfig) {
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

		exec, err := newStageExecutor(ctx, stage, cfg.Pause, total)
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

		go func(idx int, s ScanConfig, in, out chan string, exec, nextExec *stageExecutor) {
			defer wg.Done()
			defer closeOutputChannel(out)

			if in == nil {
				streamStageFromFile(ctx, input, maxIP, s, cfg.Shuffled, out, exec, nextExec, cfg.Pause)
			} else {
				streamStageFromChannel(ctx, in, s, out, exec, nextExec, cfg.Pause)
			}
		}(i, stage, in, out, executors[i], next)
	}

	wg.Wait()
}

// createStageChannels creates buffered channels between pipeline stages.
func createStageChannels(cfg *ChainConfig) []chan string {
	channels := make([]chan string, len(cfg.Stages))

	for i := range channels {
		size := defaultStageChanBuf
		if i+1 < len(cfg.Stages) {
			size = max(cfg.Stages[i+1].Workers, config.GetGeneral().MaxIPsPerStage)
		}
		channels[i] = make(chan string, size)
	}

	return channels
}

// getInputChannel returns the input channel for a stage.
func getInputChannel(stageIdx int, channels []chan string) chan string {
	if stageIdx == 0 {
		return nil
	}
	return channels[stageIdx-1]
}

// getOutputChannel returns the output channel for a stage.
func getOutputChannel(stageIdx, total int, channels []chan string) chan string {
	if stageIdx >= total-1 {
		return nil
	}
	return channels[stageIdx]
}

// closeOutputChannel closes a channel safely.
func closeOutputChannel(ch chan string) {
	if ch != nil {
		close(ch)
	}
}

// executeBatchPipeline runs the batch-based pipeline chain.
func executeBatchPipeline(ctx context.Context, input string, maxIP int, cfg *ChainConfig) {
	totalIPs, err := iplist.CountActiveIPs(input)
	if err != nil {
		logger.CoreError("failed to count IPs: %v", err)
		totalIPs = 0
	}

	batchSize := calculateBatchSize(cfg)
	logger.CoreInfo("batch pipeline started: batch=%d ips=%d", batchSize, totalIPs)

	stream := streamIPsFromFile(ctx, input, cfg.Shuffled, maxIP, batchSize)

	executors := make([]*stageExecutor, 0, len(cfg.Stages))

	defer func() {
		for _, e := range executors {
			e.cleanup()
		}

		for _, s := range cfg.Stages {
			s.Hooks.callOnScanEnd()
		}
	}()

	for i, stage := range cfg.Stages {
		var total uint64
		if i == 0 {
			total = totalIPs
		}

		exec, err := newStageExecutor(ctx, stage, cfg.Pause, total)
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
func processBatch(ctx context.Context, batch []string, execs []*stageExecutor, pause *PauseController) {
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

// executeBatch processes a batch in worker pool.
func executeBatch(ctx context.Context, batch []string, exec *stageExecutor, pause *PauseController) []string {
	workers := getWorkerCount(exec.stage.Workers)
	input := make(chan string, workers*2)

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
		out []string
	)

	runWorkerPool(ctx, workers, pause, input, func(ip string) {
		if exec.processIP(ctx, ip) {
			mu.Lock()
			out = append(out, ip)
			mu.Unlock()
		}
	})

	return out
}

// calculateBatchSize determines optimal batch size for pipeline mode.
func calculateBatchSize(cfg *ChainConfig) int {
	if len(cfg.Stages) <= 1 {
		return config.GetGeneral().BatchSize
	}

	maxWorkers := 0
	for _, s := range cfg.Stages[1:] {
		if s.Workers > maxWorkers {
			maxWorkers = s.Workers
		}
	}

	return max(maxWorkers, config.GetGeneral().BatchSize)
}

// streamIPsFromFile streams IPs in batches from file input.
func streamIPsFromFile(ctx context.Context, input string, shuffled bool, maxIP, batchSize int) <-chan []string {
	out := make(chan []string, 2)

	go func() {
		defer close(out)

		ipCh := make(chan string, batchSize*2)
		done := make(chan error, 1)

		go func() {
			defer close(ipCh)
			done <- iplist.StreamActiveIPs(ctx, input, maxIP, shuffled, ipCh)
		}()

		batch := make([]string, 0, batchSize)

		for ip := range ipCh {
			batch = append(batch, ip)

			if len(batch) >= batchSize {
				select {
				case out <- batch:
					batch = make([]string, 0, batchSize)
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
	maxIP int,
	stage ScanConfig,
	shuffled bool,
	output chan string,
	exec *stageExecutor,
	next *stageExecutor,
	pause *PauseController,
) {
	workers := getWorkerCount(stage.Workers)
	in := make(chan string, workers*2)

	done := make(chan error, 1)

	go func() {
		defer close(in)
		done <- iplist.StreamActiveIPs(ctx, input, maxIP, shuffled, in)
	}()

	runWorkerPool(ctx, workers, pause, in, func(ip string) {
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
	input chan string,
	stage ScanConfig,
	output chan string,
	exec *stageExecutor,
	next *stageExecutor,
	pause *PauseController,
) {
	workers := getWorkerCount(stage.Workers)

	runWorkerPool(ctx, workers, pause, input, func(ip string) {
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
