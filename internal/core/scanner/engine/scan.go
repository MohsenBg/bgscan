package engine

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"bgscan/internal/core/config"
	"bgscan/internal/core/iplist"
	"bgscan/internal/core/result"
	"bgscan/internal/logger"
)

// RunScan orchestrates the full lifecycle of a standalone scan stage.
// It counts active targets, feeds them into a worker pool, tracks metrics,
// and flushes outputs to disk. It blocks until the scan completes or is cancelled.
func RunScan(
	ctx context.Context,
	input string,
	maxIP int,
	cfg ScanConfig,
	shuffled bool,
	pause *PauseController,
) {
	// Resolve and calculate workload size
	total, err := iplist.CountActiveIPs(input)
	if err != nil {
		cfg.Hooks.callOnError(err)
		cfg.Hooks.callOnScanEnd()
		return
	}

	if total == 0 {
		cfg.Hooks.callOnScanEnd()
		return
	}

	workers := min(cfg.Workers, int(total))
	if workers <= 0 {
		workers = 1
	}

	// Instantiate communication pipelines
	ips := make(chan string, workers*2)
	results := make(chan result.IPScanResult, workers*4)

	var (
		processed atomic.Uint64
		succeed   atomic.Uint64
		start     = time.Now()
	)

	// Set up synchronization boundaries
	var writerDone sync.WaitGroup
	progressDone := make(chan struct{})

	// Initialize required system dependencies
	cfg.Writer.Start()
	if err := cfg.Probe.Init(ctx); err != nil {
		_ = cfg.Writer.Stop()
		cfg.Hooks.callOnError(err)
		cfg.Hooks.callOnScanEnd()
		return
	}

	rateCh := makeRateCh(cfg.Rate)

	// Ensure structural cleanup and shutdown operations run on exit
	defer func() {
		if err := cfg.Writer.Stop(); err != nil {
			logger.CoreError("error stopping writer: %v", err)
		}

		if err := cfg.Probe.Close(); err != nil {
			cfg.Hooks.callOnError(err)
		}

		// Send one final, exact progress update before signaling absolute termination
		reportProgress(start, getPauseDuration(pause), total, processed.Load(), succeed.Load(), cfg.Hooks.OnProgress)
		cfg.Hooks.callOnScanEnd()
	}()

	// Spin up downstream consumer (File Writer)
	writerDone.Go(func() {
		for res := range results {
			cfg.Writer.Write(res)
		}
	})

	// Spin up background progress telemetry metrics
	go runProgressReporter(ctx, progressDone, pause, start, total, &processed, &succeed, cfg.Hooks.OnProgress)

	// Spin up downstream compute engine (Worker Pool)
	var workerWg sync.WaitGroup
	workerWg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer workerWg.Done()
			runWorker(ctx, pause, ips, func(ip string) {
				// Enforce rate limiting boundaries
				select {
				case <-rateCh:
				case <-ctx.Done():
					return
				}

				// Execute probe
				res, err := cfg.Probe.Run(ctx, ip)
				processed.Add(1)

				if err != nil {
					logger.CoreError("probe failed for %s: %v", ip, err)
					return
				}

				succeed.Add(1)
				cfg.Hooks.callOnSuccess(*res)

				select {
				case results <- *res:
				case <-ctx.Done():
				}
			})
		}()
	}

	// Execute upstream producer (File Reader/Streamer)
	streamErr := iplist.StreamActiveIPs(ctx, input, maxIP, shuffled, ips)
	if streamErr != nil {
		cfg.Hooks.callOnError(streamErr)
	}
	close(ips) // Signal worker pools that production has finished

	workerWg.Wait()     // Wait for workers to drain current in-flight channel elements
	close(results)      // Signal writer thread to flush and terminate
	writerDone.Wait()   // Wait for writer thread disk flush to resolve cleanly
	close(progressDone) // Halt telemetry tracking routines
}

// makeRateCh builds an unbuffered channel emitting ticks to enforce throughput caps.
func makeRateCh(rate int) <-chan time.Time {
	if rate > 0 {
		return time.NewTicker(time.Second / time.Duration(rate)).C
	}
	ch := make(chan time.Time)
	close(ch)
	return ch
}

// runProgressReporter monitors telemetry updates at configurable heartbeat intervals.
func runProgressReporter(
	ctx context.Context,
	progressDone <-chan struct{},
	pause *PauseController,
	start time.Time,
	total uint64,
	processed *atomic.Uint64,
	succeed *atomic.Uint64,
	onProgress func(Progress),
) {
	if onProgress == nil {
		return
	}

	interval := config.Get().General.StatusInterval.Duration()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-progressDone:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			if pause != nil && pause.IsPaused() {
				continue
			}
			reportProgress(
				start,
				getPauseDuration(pause),
				total,
				processed.Load(),
				succeed.Load(),
				onProgress,
			)
		}
	}
}

// getPauseDuration isolates nil-safe evaluations of pause timings.
func getPauseDuration(p *PauseController) time.Duration {
	if p == nil {
		return 0
	}
	return p.PausedDuration()
}
