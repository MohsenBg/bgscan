package dns

import (
	"context"
	"testing"
	"time"
)

func TestTestProxy_AlreadyCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if result := TestProxy(ctx, "127.0.0.1:1080", 0); result {
		t.Error("TestProxy with cancelled context should return false")
	}
}

func TestTestProxy_UnreachableProxy(t *testing.T) {
	ctx := context.Background()
	if result := TestProxy(ctx, "127.0.0.1:19999", 500*time.Millisecond); result {
		t.Error("TestProxy with unreachable proxy should return false")
	}
}

func TestTestProxy_InvalidAddress(t *testing.T) {
	ctx := context.Background()
	if result := TestProxy(ctx, "not-an-address", 500*time.Millisecond); result {
		t.Error("TestProxy with invalid address should return false")
	}
}

func TestTestProxy_TimeoutRespected(t *testing.T) {
	start := time.Now()
	timeout := 300 * time.Millisecond

	ctx := context.Background()
	result := TestProxy(ctx, "10.255.255.1:1080", timeout)

	elapsed := time.Since(start)
	if result {
		t.Error("TestProxy against unreachable host should return false")
	}
	if elapsed > 5*time.Second {
		t.Errorf("TestProxy did not respect timeout: elapsed %v, timeout %v", elapsed, timeout)
	}
}
