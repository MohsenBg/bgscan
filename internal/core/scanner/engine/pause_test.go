package engine

import (
	"context"
	"testing"
	"time"
)

func TestPauseController_InitialState(t *testing.T) {
	pc := NewPauseController()
	if pc.IsPaused() {
		t.Fatal("expected not paused initially")
	}
	if d := pc.PausedDuration(); d != 0 {
		t.Fatalf("expected zero paused duration, got %v", d)
	}
}

func TestPauseController_PauseResume(t *testing.T) {
	pc := NewPauseController()
	pc.Pause()

	if !pc.IsPaused() {
		t.Fatal("expected paused after Pause()")
	}

	time.Sleep(10 * time.Millisecond)
	pc.Resume()

	if pc.IsPaused() {
		t.Fatal("expected not paused after Resume()")
	}
	if pc.PausedDuration() < 10*time.Millisecond {
		t.Fatalf("expected paused duration >= 10ms, got %v", pc.PausedDuration())
	}
}

func TestPauseController_DoublePauseIsNoOp(t *testing.T) {
	pc := NewPauseController()
	pc.Pause()
	pc.Pause()

	if !pc.IsPaused() {
		t.Fatal("expected still paused")
	}
	pc.Stop()
}

func TestPauseController_DoubleResumeIsNoOp(t *testing.T) {
	pc := NewPauseController()
	pc.Pause()
	pc.Resume()
	pc.Resume()
}

func TestPauseController_Wait_ReturnsFalseOnContextCancel(t *testing.T) {
	pc := NewPauseController()
	pc.Pause()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if pc.Wait(ctx) {
		t.Fatal("Wait should return false when context is already cancelled")
	}
}

func TestPauseController_Wait_ReturnsFalseAfterStop(t *testing.T) {
	pc := NewPauseController()
	pc.Pause()

	done := make(chan bool, 1)
	go func() {
		done <- pc.Wait(context.Background())
	}()

	time.Sleep(5 * time.Millisecond)
	pc.Stop()

	select {
	case v := <-done:
		if v {
			t.Fatal("Wait should return false after Stop")
		}
	case <-time.After(time.Second):
		t.Fatal("Wait did not unblock after Stop")
	}
}

func TestPauseController_Wait_ReturnsTrueWhenNotPaused(t *testing.T) {
	pc := NewPauseController()
	defer pc.Stop()

	if !pc.Wait(context.Background()) {
		t.Fatal("Wait should return true when not paused")
	}
}

func TestPauseController_Wait_BlocksThenUnblocksOnResume(t *testing.T) {
	pc := NewPauseController()
	pc.Pause()

	done := make(chan bool, 1)
	go func() {
		done <- pc.Wait(context.Background())
	}()

	time.Sleep(5 * time.Millisecond)
	pc.Resume()

	select {
	case v := <-done:
		if !v {
			t.Fatal("Wait should return true after Resume")
		}
	case <-time.After(time.Second):
		t.Fatal("Wait did not unblock after Resume")
	}
	pc.Stop()
}

func TestPauseController_PausedDuration_AccumulatesAcrossCycles(t *testing.T) {
	pc := NewPauseController()

	for range 3 {
		pc.Pause()
		time.Sleep(5 * time.Millisecond)
		pc.Resume()
	}

	if d := pc.PausedDuration(); d < 15*time.Millisecond {
		t.Fatalf("expected at least 15ms accumulated pause, got %v", d)
	}
	pc.Stop()
}

func TestPauseController_PausedDuration_IncludesCurrentPause(t *testing.T) {
	pc := NewPauseController()
	pc.Pause()
	time.Sleep(10 * time.Millisecond)

	if pc.PausedDuration() < 5*time.Millisecond {
		t.Fatal("expected PausedDuration to include current pause interval")
	}
	pc.Stop()
}
