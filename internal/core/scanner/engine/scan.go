package engine

import (
	"context"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/iplist"
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/logger"
)

type scanProgress struct {
	start     time.Time
	total     uint64
	processed *atomic.Uint64
	success   *atomic.Uint64
	pause     PauseController
	callback  func(Progress)
}

func RunScan(ctx context.Context, input string, cfg ScanConfig) {
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

	workers := int(min(uint64(cfg.Workers), total))
	if workers < 1 {
		workers = 1
	}

	ips := make(chan netip.Addr, workers*2)
	results := make(chan result.Result, workers*4)

	var (
		processed atomic.Uint64
		success   atomic.Uint64
		start     = time.Now()
	)

	if err := cfg.Writer.Start(); err != nil {
		cfg.Hooks.callOnError(err)
		cfg.Hooks.callOnScanEnd()
		return
	}

	if err := cfg.Probe.Init(ctx); err != nil {
		_ = cfg.Writer.Stop()
		cfg.Hooks.callOnError(err)
		cfg.Hooks.callOnScanEnd()
		return
	}

	defer func() {
		if err := cfg.Writer.Stop(); err != nil {
			logger.CoreError("stopping writer: %v", err)
		}

		if err := cfg.Probe.Close(); err != nil {
			cfg.Hooks.callOnError(err)
		}

		reportProgress(
			start,
			pausedDuration(cfg.Pause),
			total,
			processed.Load(),
			success.Load(),
			cfg.Hooks.OnProgress,
		)

		cfg.Hooks.callOnScanEnd()
	}()

	var writerWG sync.WaitGroup
	writerWG.Go(func() {
		for res := range results {
			cfg.Writer.Write(res)
		}
	})

	progress := scanProgress{
		start:     start,
		total:     total,
		processed: &processed,
		success:   &success,
		pause:     cfg.Pause,
		callback:  cfg.Hooks.OnProgress,
	}

	progressDone := make(chan struct{})
	var progressWG sync.WaitGroup

	progressWG.Go(func() {
		runProgressReporter(ctx, cfg.ProgressInterval, progress, progressDone)
	})

	var workerWG sync.WaitGroup
	workerWG.Add(workers)

	for range workers {
		go func() {
			defer workerWG.Done()

			runWorker(ctx, cfg.Pause, ips, func(ip netip.Addr) {
				runProbe(ctx, ip, cfg, &processed, &success, results)
			})
		}()
	}

	if err := iplist.StreamActiveIPs(
		ctx,
		input,
		cfg.MaxIPsToTest,
		cfg.Shuffled,
		ips,
	); err != nil {
		cfg.Hooks.callOnError(err)
	}

	close(ips)

	workerWG.Wait()
	close(results)

	writerWG.Wait()

	close(progressDone)
	progressWG.Wait()
}

// runProgressReporter periodically publishes scan statistics.
func runProgressReporter(
	ctx context.Context,
	interval time.Duration,
	p scanProgress,
	progressDone <-chan struct{},
) {
	if p.callback == nil || interval <= 0 {
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-progressDone:
			return

		case <-ticker.C:
			if p.pause != nil && p.pause.IsPaused() {
				continue
			}

			reportProgress(
				p.start,
				pausedDuration(p.pause),
				p.total,
				p.processed.Load(),
				p.success.Load(),
				p.callback,
			)
		}
	}
}

// pausedDuration returns the cumulative paused duration, or zero if pause is disabled.
func pausedDuration(pause PauseController) time.Duration {
	if pause == nil {
		return 0
	}
	return pause.PausedDuration()
}

// runProbe acquires a rate-limit token, executes the probe, then enforces
// the minimum probe duration to prevent socket burst on slow-kernel devices.
func runProbe(
	ctx context.Context,
	ip netip.Addr,
	cfg ScanConfig,
	processed *atomic.Uint64,
	succeed *atomic.Uint64,
	results chan<- result.Result,
) {
	if cfg.RateLimiter != nil {
		if err := cfg.RateLimiter.Wait(ctx); err != nil {
			return
		}
	}
	probeStart := time.Now()
	res, err := cfg.Probe.Run(ctx, ip)
	processed.Add(1)

	if err != nil {
		logger.CoreError("probe failed for %s: %v", ip, err)
	} else {
		succeed.Add(1)
		cfg.Hooks.callOnSuccess(res)
		select {
		case results <- res:
		case <-ctx.Done():
		}
	}

	if cfg.MinProbeDuration > 0 {
		if remaining := cfg.MinProbeDuration - time.Since(probeStart); remaining > 0 {
			select {
			case <-time.After(remaining):
			case <-ctx.Done():
			}
		}
	}
}
