package engine

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
)

func TestRunScan_EmptyInput(t *testing.T) {
	path := ipFile(t)
	prb := &mockProbe{}
	w := &mockWriter{}
	endCh := make(chan struct{}, 1)

	_ = RunScanWithChain(context.Background(), path, ChainConfig{
		MaxIPsToTest: 0,
		Mode:         ModeSequential,
		Stages: []StageConfig{
			{
				Workers: 2,
				Probe:   prb,
				Writer:  w,
				Hooks:   ScanHooks{OnScanEnd: func() { endCh <- struct{}{} }},
			},
		},
		Shuffled: false,
		Pause:    NewPauseController(),
	})

	select {
	case <-endCh:
	case <-time.After(time.Second):
		t.Fatal("OnScanEnd not called for empty input")
	}

	if prb.runCalled.Load() != 0 {
		t.Fatal("probe Run should not be called for empty input")
	}
}

func TestRunScan_ScansAllIPs(t *testing.T) {
	ips := []string{"1.2.3.4", "5.6.7.8", "9.10.11.12"}
	path := ipFile(t, ips...)
	prb := &mockProbe{}
	w := &mockWriter{}

	var successCount atomic.Int32
	endCh := make(chan struct{}, 1)

	_ = RunScanWithChain(context.Background(), path, ChainConfig{
		MaxIPsToTest: 0,
		Mode:         ModeSequential,
		Stages: []StageConfig{
			{
				Workers: 2,
				Probe:   prb,
				Writer:  w,
				Hooks: ScanHooks{
					OnSuccess: func(r result.Result) { successCount.Add(1) },
					OnScanEnd: func() { endCh <- struct{}{} },
				},
			},
		},
		Shuffled: false,
		Pause:    NewPauseController(),
	})

	<-endCh

	if got := int(successCount.Load()); got != len(ips) {
		t.Fatalf("expected %d successes, got %d", len(ips), got)
	}
	if got := len(w.results()); got != len(ips) {
		t.Fatalf("expected %d written results, got %d", len(ips), got)
	}
}

func TestRunScan_ProbeInitError(t *testing.T) {
	path := ipFile(t, "1.2.3.4")
	prb := &mockProbe{initErr: errors.New("init failed")}
	w := &mockWriter{}

	gotErr := RunScanWithChain(context.Background(), path, ChainConfig{
		MaxIPsToTest: 0,
		Mode:         ModeSequential,
		Stages: []StageConfig{
			{
				Workers: 1,
				Probe:   prb,
				Writer:  w,
			},
		},
		Shuffled: false,
		Pause:    NewPauseController(),
	})

	if gotErr == nil {
		t.Fatal("expected probe init error")
	}
	if prb.runCalled.Load() != 0 {
		t.Fatal("probe Run should not be called when Init fails")
	}
}

func TestRunScan_ProbeRunError(t *testing.T) {
	path := ipFile(t, "1.2.3.4", "2.3.4.5")
	prb := &mockProbe{runErr: errors.New("probe failure")}
	w := &mockWriter{}

	var successCount atomic.Int32
	endCh := make(chan struct{}, 1)

	_ = RunScanWithChain(context.Background(), path, ChainConfig{
		MaxIPsToTest: 0,
		Mode:         ModeSequential,
		Stages: []StageConfig{
			{
				Workers: 1,
				Probe:   prb,
				Writer:  w,
				Hooks: ScanHooks{
					OnSuccess: func(r result.Result) { successCount.Add(1) },
					OnScanEnd: func() { endCh <- struct{}{} },
				},
			},
		},
		Shuffled: false,
		Pause:    NewPauseController(),
	})

	<-endCh

	if successCount.Load() != 0 {
		t.Fatal("expected zero successes when all probes fail")
	}
	if len(w.results()) != 0 {
		t.Fatal("expected nothing written when all probes fail")
	}
}

func TestRunScan_WriterWriteErrorIsFatal(t *testing.T) {
	ips := make([]string, 200)
	for i := range ips {
		ips[i] = fmt.Sprintf("10.0.%d.%d", i/256, i%256)
	}
	path := ipFile(t, ips...)
	prb := &mockProbe{}
	w := &mockWriter{writeErr: errors.New("disk full")}

	err := RunScanWithChain(context.Background(), path, ChainConfig{
		MaxIPsToTest: 0,
		Mode:         ModeSequential,
		Stages: []StageConfig{
			{
				Workers: 4,
				Probe:   prb,
				Writer:  w,
			},
		},
		Shuffled: false,
		Pause:    NewPauseController(),
	})

	if !errors.Is(err, ErrWriterFatal) {
		t.Fatalf("expected ErrWriterFatal, got %v", err)
	}
	if got := len(w.results()); got != 0 {
		t.Fatalf("expected no persisted results, got %d", got)
	}
}

func TestRunScan_WriterWriteErrorIsIgnored(t *testing.T) {
	ips := make([]string, 200)
	for i := range ips {
		ips[i] = fmt.Sprintf("10.0.%d.%d", i/256, i%256)
	}
	path := ipFile(t, ips...)
	prb := &mockProbe{}
	w := &mockWriter{writeErr: errors.New("disk full")}

	var successCount atomic.Int32

	err := RunScanWithChain(context.Background(), path, ChainConfig{
		MaxIPsToTest:      0,
		Mode:              ModeSequential,
		IgnoreWriterError: true,
		Stages: []StageConfig{
			{
				Workers: 4,
				Probe:   prb,
				Writer:  w,
				Hooks: ScanHooks{
					OnSuccess: func(result.Result) { successCount.Add(1) },
				},
			},
		},
		Shuffled: false,
		Pause:    NewPauseController(),
	})
	if err != nil {
		t.Fatalf("expected ignored writer error, got %v", err)
	}
	if got := successCount.Load(); got != int32(len(ips)) {
		t.Fatalf("expected %d successes, got %d", len(ips), got)
	}
	if got := prb.runCalled.Load(); got != int32(len(ips)) {
		t.Fatalf("expected %d processed IPs, got %d", len(ips), got)
	}
	if got := len(w.results()); got != 0 {
		t.Fatalf("expected no persisted results, got %d", got)
	}
}

func TestRunScan_WriterCloseErrorIsIgnored(t *testing.T) {
	path := ipFile(t, "1.2.3.4")
	prb := &mockProbe{}
	w := &mockWriter{stopErr: errors.New("close failed")}

	err := RunScanWithChain(context.Background(), path, ChainConfig{
		MaxIPsToTest:      0,
		Mode:              ModeSequential,
		IgnoreWriterError: true,
		Stages: []StageConfig{
			{
				Workers: 1,
				Probe:   prb,
				Writer:  w,
			},
		},
		Shuffled: false,
		Pause:    NewPauseController(),
	})
	if err != nil {
		t.Fatalf("expected ignored writer close error, got %v", err)
	}
}

func TestRunScan_WriterStartError(t *testing.T) {
	path := ipFile(t, "1.2.3.4")
	prb := &mockProbe{}
	w := &mockWriter{startErr: errors.New("disk full")}

	gotErr := RunScanWithChain(context.Background(), path, ChainConfig{
		MaxIPsToTest: 0,
		Mode:         ModeSequential,
		Stages: []StageConfig{
			{
				Workers: 1,
				Probe:   prb,
				Writer:  w,
			},
		},
		Shuffled: false,
		Pause:    NewPauseController(),
	})

	if gotErr == nil {
		t.Fatal("expected writer Start error")
	}
}

func TestRunScan_WriterStartErrorIsIgnored(t *testing.T) {
	path := ipFile(t, "1.2.3.4")
	prb := &mockProbe{}
	w := &mockWriter{startErr: errors.New("disk full")}

	err := RunScanWithChain(context.Background(), path, ChainConfig{
		MaxIPsToTest:      0,
		Mode:              ModeSequential,
		IgnoreWriterError: true,
		Stages: []StageConfig{
			{
				Workers: 1,
				Probe:   prb,
				Writer:  w,
			},
		},
		Shuffled: false,
		Pause:    NewPauseController(),
	})
	if err != nil {
		t.Fatalf("expected ignored writer Start error, got %v", err)
	}
	if got := prb.runCalled.Load(); got != 1 {
		t.Fatalf("expected scan to continue after writer Start error, processed %d IPs", got)
	}
	if got := len(w.results()); got != 0 {
		t.Fatalf("expected no writes when the writer failed to start, got %d", got)
	}
}

func TestRunScan_ContextCancellation(t *testing.T) {
	ips := make([]string, 500)
	for i := range ips {
		ips[i] = fmt.Sprintf("10.0.%d.%d", i/256, i%256)
	}
	path := ipFile(t, ips...)

	prb := &mockProbe{runDelay: 5 * time.Millisecond}
	w := &mockWriter{}

	ctx, cancel := context.WithCancel(context.Background())
	endCh := make(chan struct{}, 1)

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	_ = RunScanWithChain(ctx, path, ChainConfig{
		MaxIPsToTest: 0,
		Mode:         ModeSequential,
		Stages: []StageConfig{
			{
				Workers: 4,
				Probe:   prb,
				Writer:  w,
				Hooks:   ScanHooks{OnScanEnd: func() { endCh <- struct{}{} }},
			},
		},
		Shuffled: false,
		Pause:    NewPauseController(),
	})

	select {
	case <-endCh:
	case <-time.After(3 * time.Second):
		t.Fatal("RunScan did not return after context cancellation")
	}

	if prb.runCalled.Load() >= 500 {
		t.Fatal("expected scan to stop early on cancellation")
	}
}

func TestRunScan_MaxIPLimit(t *testing.T) {
	ips := []string{"1.1.1.1", "2.2.2.2", "3.3.3.3", "4.4.4.4", "5.5.5.5"}
	path := ipFile(t, ips...)
	prb := &mockProbe{}
	w := &mockWriter{}

	const maxIPs uint64 = 2
	endCh := make(chan struct{}, 1)

	_ = RunScanWithChain(context.Background(), path, ChainConfig{
		MaxIPsToTest: maxIPs,
		Mode:         ModeSequential,
		Stages: []StageConfig{
			{
				Workers: 1,
				Probe:   prb,
				Writer:  w,
				Hooks:   ScanHooks{OnScanEnd: func() { endCh <- struct{}{} }},
			},
		},
		Shuffled: false,
		Pause:    NewPauseController(),
	})

	<-endCh

	if got := prb.runCalled.Load(); got > int32(maxIPs) {
		t.Fatalf("expected at most %d IPs processed, got %d", maxIPs, got)
	}
}

func TestRunScan_PauseAndResumeDuringRun(t *testing.T) {
	ips := make([]string, 20)
	for i := range ips {
		ips[i] = fmt.Sprintf("10.0.0.%d", i+1)
	}
	path := ipFile(t, ips...)
	prb := &mockProbe{runDelay: 2 * time.Millisecond}
	w := &mockWriter{}

	pc := NewPauseController()
	endCh := make(chan struct{}, 1)

	go func() {
		time.Sleep(10 * time.Millisecond)
		pc.Pause()
		time.Sleep(20 * time.Millisecond)
		pc.Resume()
	}()

	_ = RunScanWithChain(context.Background(), path, ChainConfig{
		MaxIPsToTest: 0,
		Mode:         ModeSequential,
		Stages: []StageConfig{
			{
				Workers: 2,
				Probe:   prb,
				Writer:  w,
				Hooks:   ScanHooks{OnScanEnd: func() { endCh <- struct{}{} }},
			},
		},
		Shuffled: false,
		Pause:    pc,
	})

	select {
	case <-endCh:
	case <-time.After(5 * time.Second):
		t.Fatal("RunScan did not complete after pause/resume cycle")
	}

	if prb.runCalled.Load() == 0 {
		t.Fatal("expected at least some IPs processed")
	}
}

func TestRunScan_OnProgressCallback(t *testing.T) {
	ips := make([]string, 10)
	for i := range ips {
		ips[i] = fmt.Sprintf("10.0.0.%d", i+1)
	}
	path := ipFile(t, ips...)
	prb := &mockProbe{runDelay: 5 * time.Millisecond}
	w := &mockWriter{}

	var progressCalls atomic.Int32
	endCh := make(chan struct{}, 1)

	_ = RunScanWithChain(context.Background(), path, ChainConfig{
		MaxIPsToTest: 0,
		Mode:         ModeSequential,
		Stages: []StageConfig{
			{
				Workers:          2,
				Probe:            prb,
				Writer:           w,
				ProgressInterval: 10 * time.Millisecond,
				Hooks: ScanHooks{
					OnProgress: func(p Progress) {
						progressCalls.Add(1)
						if p.RatePerSec < 0 {
							panic("negative rate")
						}
						if p.Percent < 0 || p.Percent > 100 {
							panic("percent out of range")
						}
					},
					OnScanEnd: func() { endCh <- struct{}{} },
				},
			},
		},
		Shuffled: false,
		Pause:    NewPauseController(),
	})

	<-endCh

	if progressCalls.Load() == 0 {
		t.Fatal("expected at least one OnProgress call")
	}
}

func TestWorkerPool_ConcurrencyRespected(t *testing.T) {
	const workerCount = 4
	ips := make([]string, 40)
	for i := range ips {
		ips[i] = fmt.Sprintf("10.0.0.%d", i+1)
	}
	path := ipFile(t, ips...)

	var (
		mu         sync.Mutex
		concurrent int
		peakConc   int
	)

	prb := &mockProbe{}
	prb.onRun = func(_ netip.Addr) {
		mu.Lock()
		concurrent++
		if concurrent > peakConc {
			peakConc = concurrent
		}
		mu.Unlock()

		time.Sleep(5 * time.Millisecond)

		mu.Lock()
		concurrent--
		mu.Unlock()
	}

	w := &mockWriter{}
	endCh := make(chan struct{}, 1)

	_ = RunScanWithChain(context.Background(), path, ChainConfig{
		MaxIPsToTest: 0,
		Mode:         ModeSequential,
		Stages: []StageConfig{
			{
				Workers: workerCount,
				Probe:   prb,
				Writer:  w,
				Hooks:   ScanHooks{OnScanEnd: func() { endCh <- struct{}{} }},
			},
		},
		Shuffled: false,
		Pause:    NewPauseController(),
	})

	<-endCh

	mu.Lock()
	defer mu.Unlock()

	if peakConc > workerCount {
		t.Fatalf("peak concurrency %d exceeded worker count %d", peakConc, workerCount)
	}
	if peakConc == 0 {
		t.Fatal("expected some concurrent execution")
	}
}

func TestWorkerPool_ZeroWorkersDefaultsToOne(t *testing.T) {
	path := ipFile(t, "1.2.3.4")
	prb := &mockProbe{}
	w := &mockWriter{}
	endCh := make(chan struct{}, 1)

	_ = RunScanWithChain(context.Background(), path, ChainConfig{
		MaxIPsToTest: 0,
		Mode:         ModeSequential,
		Stages: []StageConfig{
			{
				Workers: 0,
				Probe:   prb,
				Writer:  w,
				Hooks:   ScanHooks{OnScanEnd: func() { endCh <- struct{}{} }},
			},
		},
		Shuffled: false,
		Pause:    NewPauseController(),
	})

	select {
	case <-endCh:
	case <-time.After(2 * time.Second):
		t.Fatal("scan did not complete with Workers=0")
	}

	if prb.runCalled.Load() != 1 {
		t.Fatalf("expected 1 run call, got %d", prb.runCalled.Load())
	}
}

func TestProgress_FieldsViaScan(t *testing.T) {
	ips := make([]string, 5)
	for i := range ips {
		ips[i] = fmt.Sprintf("10.0.0.%d", i+1)
	}
	path := ipFile(t, ips...)
	prb := &mockProbe{runDelay: 5 * time.Millisecond}
	w := &mockWriter{}

	var (
		mu        sync.Mutex
		snapshots []Progress
	)
	endCh := make(chan struct{}, 1)

	_ = RunScanWithChain(context.Background(), path, ChainConfig{
		MaxIPsToTest: 0,
		Mode:         ModeSequential,
		Stages: []StageConfig{
			{
				Workers:          1,
				Probe:            prb,
				Writer:           w,
				ProgressInterval: 5 * time.Millisecond,
				Hooks: ScanHooks{
					OnProgress: func(p Progress) {
						mu.Lock()
						snapshots = append(snapshots, p)
						mu.Unlock()
					},
					OnScanEnd: func() { endCh <- struct{}{} },
				},
			},
		},
		Shuffled: false,
		Pause:    NewPauseController(),
	})

	<-endCh

	mu.Lock()
	defer mu.Unlock()

	if len(snapshots) == 0 {
		t.Fatal("expected at least one progress snapshot")
	}

	last := snapshots[len(snapshots)-1]

	if last.Total != uint64(len(ips)) {
		t.Fatalf("last progress Total = %d, want %d", last.Total, len(ips))
	}
	if last.Processed != uint64(len(ips)) {
		t.Fatalf("last progress Processed = %d, want %d", last.Processed, len(ips))
	}
	if last.Succeed != uint64(len(ips)) {
		t.Fatalf("last progress Succeed = %d, want %d", last.Succeed, len(ips))
	}
	if last.Percent != 100.0 {
		t.Fatalf("last progress Percent = %v, want 100", last.Percent)
	}
	if last.RatePerSec <= 0 {
		t.Fatalf("last progress RatePerSec = %v, want > 0", last.RatePerSec)
	}
}
