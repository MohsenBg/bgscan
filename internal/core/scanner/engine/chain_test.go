package engine

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunScanWithChain_NilConfig(t *testing.T) {
	RunScanWithChain(context.Background(), "", 0, nil)
}

func TestRunScanWithChain_EmptyStages(t *testing.T) {
	RunScanWithChain(context.Background(), "", 0, &ChainConfig{})
}

func TestRunScanWithChain_Sequential_TwoStages(t *testing.T) {
	ips := []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"}
	stage1Path := ipFile(t, ips...)
	stage2Path := ipFile(t, ips...)

	w1 := &mockWriter{resultPath: stage2Path}
	w2 := &mockWriter{}
	prb1 := &mockProbe{}
	prb2 := &mockProbe{}

	endCh1 := make(chan struct{}, 1)
	endCh2 := make(chan struct{}, 1)

	pc := NewPauseController()
	defer pc.Stop()

	RunScanWithChain(context.Background(), stage1Path, 0, &ChainConfig{
		Mode:  ModeSequential,
		Pause: pc,
		Stages: []ScanConfig{
			{Workers: 2, Probe: prb1, Writer: w1, Hooks: ScanHooks{OnScanEnd: func() { endCh1 <- struct{}{} }}},
			{Workers: 2, Probe: prb2, Writer: w2, Hooks: ScanHooks{OnScanEnd: func() { endCh2 <- struct{}{} }}},
		},
	})

	for _, ch := range []chan struct{}{endCh1, endCh2} {
		select {
		case <-ch:
		case <-time.After(5 * time.Second):
			t.Fatal("stage did not complete within timeout")
		}
	}

	if prb1.runCalled.Load() == 0 {
		t.Fatal("stage 1 probe was never called")
	}
	if prb2.runCalled.Load() == 0 {
		t.Fatal("stage 2 probe was never called")
	}
}

func TestRunScanWithChain_Sequential_SkipsOnEmptyIntermediate(t *testing.T) {
	emptyPath := ipFile(t)
	stage1Input := ipFile(t, "1.2.3.4")

	w1 := &mockWriter{resultPath: emptyPath}
	prb2 := &mockProbe{}
	endCh1 := make(chan struct{}, 1)

	RunScanWithChain(context.Background(), stage1Input, 0, &ChainConfig{
		Mode:  ModeSequential,
		Pause: NewPauseController(),
		Stages: []ScanConfig{
			{Workers: 1, Probe: &mockProbe{}, Writer: w1, Hooks: ScanHooks{OnScanEnd: func() { endCh1 <- struct{}{} }}},
			{Workers: 1, Probe: prb2, Writer: &mockWriter{}},
		},
	})

	<-endCh1

	if prb2.runCalled.Load() != 0 {
		t.Fatal("stage 2 should be skipped when stage 1 output is empty")
	}
}

func TestRunScanWithChain_Streaming_TwoStages(t *testing.T) {
	ips := make([]string, 10)
	for i := range ips {
		ips[i] = fmt.Sprintf("192.168.1.%d", i+1)
	}
	path := ipFile(t, ips...)

	prb1 := &mockProbe{}
	prb2 := &mockProbe{}

	var end1, end2 atomic.Int32

	pc := NewPauseController()
	defer pc.Stop()

	RunScanWithChain(context.Background(), path, 0, &ChainConfig{
		Mode:      ModeStreaming,
		MaxBuffer: 100,
		Pause:     pc,
		Stages: []ScanConfig{
			{Workers: 2, Probe: prb1, Writer: &mockWriter{}, Hooks: ScanHooks{OnScanEnd: func() { end1.Add(1) }}},
			{Workers: 2, Probe: prb2, Writer: &mockWriter{}, Hooks: ScanHooks{OnScanEnd: func() { end2.Add(1) }}},
		},
	})

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if end1.Load() > 0 && end2.Load() > 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if end1.Load() == 0 {
		t.Fatal("stage 1 OnScanEnd was not called")
	}
	if end2.Load() == 0 {
		t.Fatal("stage 2 OnScanEnd was not called")
	}
	if prb1.runCalled.Load() == 0 {
		t.Fatal("stage 1 probe was never called")
	}
	if prb2.runCalled.Load() == 0 {
		t.Fatal("stage 2 probe was never called")
	}
}

func TestRunScanWithChain_Streaming_ContextCancellation(t *testing.T) {
	ips := make([]string, 200)
	for i := range ips {
		ips[i] = fmt.Sprintf("10.1.%d.%d", i/256, i%256)
	}
	path := ipFile(t, ips...)

	prb := &mockProbe{runDelay: 5 * time.Millisecond}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		RunScanWithChain(ctx, path, 0, &ChainConfig{
			Mode:  ModeStreaming,
			Pause: NewPauseController(),
			Stages: []ScanConfig{
				{Workers: 2, Probe: prb, Writer: &mockWriter{}},
			},
		})
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("streaming chain did not return after context cancellation")
	}
}

func TestRunScanWithChain_Batch_TwoStages(t *testing.T) {
	ips := make([]string, 15)
	for i := range ips {
		ips[i] = fmt.Sprintf("172.16.0.%d", i+1)
	}
	path := ipFile(t, ips...)

	prb1 := &mockProbe{}
	prb2 := &mockProbe{}

	var end1, end2 atomic.Int32

	pc := NewPauseController()
	defer pc.Stop()

	RunScanWithChain(context.Background(), path, 0, &ChainConfig{
		Mode:      ModeBatch,
		BatchSize: 5,
		Pause:     pc,
		Stages: []ScanConfig{
			{Workers: 2, Probe: prb1, Writer: &mockWriter{}, Hooks: ScanHooks{OnScanEnd: func() { end1.Add(1) }}},
			{Workers: 2, Probe: prb2, Writer: &mockWriter{}, Hooks: ScanHooks{OnScanEnd: func() { end2.Add(1) }}},
		},
	})

	if end1.Load() == 0 {
		t.Fatal("batch stage 1 OnScanEnd not called")
	}
	if end2.Load() == 0 {
		t.Fatal("batch stage 2 OnScanEnd not called")
	}
	if int(prb1.runCalled.Load()) != len(ips) {
		t.Fatalf("stage 1: expected %d runs, got %d", len(ips), prb1.runCalled.Load())
	}
	if int(prb2.runCalled.Load()) != len(ips) {
		t.Fatalf("stage 2: expected %d runs, got %d", len(ips), prb2.runCalled.Load())
	}
}

func TestRunScanWithChain_Batch_FiltersProperly(t *testing.T) {
	ips := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4"}
	path := ipFile(t, ips...)

	prb2 := &mockProbe{}

	RunScanWithChain(context.Background(), path, 0, &ChainConfig{
		Mode:      ModeBatch,
		BatchSize: 10,
		Pause:     NewPauseController(),
		Stages: []ScanConfig{
			{
				Workers: 1,
				Probe:   &filteringProbe{failIPs: map[string]bool{"10.0.0.2": true, "10.0.0.4": true}},
				Writer:  &mockWriter{},
			},
			{Workers: 1, Probe: prb2, Writer: &mockWriter{}},
		},
	})

	if got := prb2.runCalled.Load(); got != 2 {
		t.Fatalf("expected stage 2 to be called 2 times, got %d", got)
	}
}
