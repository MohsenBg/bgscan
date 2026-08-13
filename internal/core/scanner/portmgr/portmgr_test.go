package portmgr

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"
	"time"
)

func newTestManager(t *testing.T, startPort, count uint16) *manager {
	t.Helper()

	got, err := New(startPort, count)
	if err != nil {
		t.Fatal(err)
	}

	m, ok := got.(*manager)
	if !ok {
		t.Fatalf("New returned %T, want *manager", got)
	}

	return m
}

func TestGenerateInstanceRange(t *testing.T) {
	t.Parallel()

	base := uint16(20000)
	size := uint16(64)

	start, gotSize := GenerateInstanceRange(base, size)

	if gotSize != size {
		t.Fatalf("size = %d, want %d", gotSize, size)
	}

	min := base + uint16(os.Getpid()%20000)
	max := min + 1499

	if start < min || start > max {
		t.Fatalf("start = %d, want range [%d, %d]", start, min, max)
	}
}

func TestNewRejectsZeroCount(t *testing.T) {
	t.Parallel()

	m, err := New(30000, 0)
	if err == nil {
		t.Fatalf("expected error, got manager=%v", m)
	}
}

func TestGetReturnsSequentialPorts(t *testing.T) {
	t.Parallel()

	m := newTestManager(t, 30000, 3)
	defer m.Close()

	m.isFree = func(uint16) bool {
		return true
	}

	ctx := context.Background()

	for i := range uint16(3) {
		got, err := m.Get(ctx)
		if err != nil {
			t.Fatal(err)
		}

		want := uint16(30000) + i
		if got != want {
			t.Fatalf("got %d, want %d", got, want)
		}
	}
}

func TestGetContextCanceled(t *testing.T) {
	t.Parallel()

	m := newTestManager(t, 30000, 1)
	defer m.Close()

	m.isFree = func(uint16) bool {
		return true
	}

	_, err := m.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = m.Get(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context canceled", err)
	}
}

func TestGetClosed(t *testing.T) {
	t.Parallel()

	m := newTestManager(t, 30000, 1)
	m.Close()

	_, err := m.Get(context.Background())
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("got %v, want %v", err, ErrClosed)
	}
}

func TestGetSkipsUnavailablePorts(t *testing.T) {
	t.Parallel()

	m := newTestManager(t, 30000, 2)
	defer m.Close()

	m.isFree = func(port uint16) bool {
		return port != 30000
	}

	got, err := m.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if got != 30001 {
		t.Fatalf("got %d, want %d", got, 30001)
	}
}

func TestReleaseMakesPortAvailable(t *testing.T) {
	t.Parallel()

	m := newTestManager(t, 30000, 1)
	defer m.Close()

	m.isFree = func(uint16) bool {
		return true
	}

	ctx := context.Background()

	port, err := m.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}

	m.Release(port)

	got, err := m.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if got != port {
		t.Fatalf("got %d, want %d", got, port)
	}
}

func TestReleaseAfterClose(t *testing.T) {
	t.Parallel()

	m := newTestManager(t, 30000, 1)
	m.Close()

	// Must not panic.
	m.Release(30000)
}

func TestCloseIsIdempotent(t *testing.T) {
	t.Parallel()

	m := newTestManager(t, 30000, 1)

	m.Close()
	m.Close()
}

func TestPortAvailable(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = ln.Close()
	}()

	port := uint16(ln.Addr().(*net.TCPAddr).Port)

	if portAvailable(port) {
		t.Fatal("expected occupied port to be unavailable")
	}

	if err := ln.Close(); err != nil {
		t.Fatal(err)
	}

	if !portAvailable(port) {
		t.Fatal("expected free port")
	}
}

func TestWaitOpenSuccess(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = ln.Close()
	}()

	m := newTestManager(t, 30000, 2)
	defer m.Close()
	err = m.WaitOpen(context.Background(), ln.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWaitOpenTimeout(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	addr := ln.Addr().String()

	if err := ln.Close(); err != nil {
		t.Fatal(err)
	}

	m := newTestManager(t, 30000, 2)
	defer m.Close()

	err = m.WaitOpen(context.Background(), addr, 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout")
	}
}

func TestRandomBase(t *testing.T) {
	t.Parallel()

	size := uint16(100)
	base := RandomBase(size)

	if base < 20000 {
		t.Fatalf("base %d below minimum", base)
	}

	if base+size >= 49152 {
		t.Fatalf("base %d overlaps ephemeral ports", base)
	}
}
