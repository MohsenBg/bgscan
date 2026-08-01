package probe

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockKillable struct {
	killed    atomic.Bool
	killErr   error
	killCalls atomic.Int32
}

func (m *mockKillable) Kill() error {
	m.killed.Store(true)
	m.killCalls.Add(1)

	return m.killErr
}

func (m *mockKillable) wasKilled() bool {
	return m.killed.Load()
}

func newTestProcessTracker(t *testing.T) *processTracker {
	t.Helper()

	got := NewProcessTracker()

	pr, ok := got.(*processTracker)
	if !ok {
		t.Fatalf("NewProcessTracker returned %T, want *processTracker", got)
	}

	return pr
}

func waitFor(t *testing.T, condition func() bool) {
	t.Helper()

	deadline := time.Now().Add(time.Second)

	for time.Now().Before(deadline) {
		if condition() {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatal("condition was not met before timeout")
}

func TestNewProcessTracker(t *testing.T) {
	pr := newTestProcessTracker(t)

	if pr.actionQueue == nil {
		t.Fatal("actionQueue is nil")
	}

	if cap(pr.actionQueue) != 100 {
		t.Errorf("cap(actionQueue) = %d, want 100", cap(pr.actionQueue))
	}
}

func TestStartIdempotent(t *testing.T) {
	pr := NewProcessTracker()
	ctx := t.Context()

	pr.Start(ctx)
	pr.Start(ctx)
	pr.Start(ctx)
}

func TestRegisterReturnsValidUUID(t *testing.T) {
	pr := NewProcessTracker()
	ctx := t.Context()
	pr.Start(ctx)

	id, err := pr.Register(ctx, &mockKillable{})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	if _, err := uuid.Parse(id); err != nil {
		t.Errorf("id %q is not a valid UUID: %v", id, err)
	}
}

func TestRegisterUniqueIDs(t *testing.T) {
	pr := NewProcessTracker()
	ctx := t.Context()
	pr.Start(ctx)

	seen := make(map[string]struct{}, 20)

	for i := range 20 {
		id, err := pr.Register(ctx, &mockKillable{})
		if err != nil {
			t.Fatalf("Register() #%d error: %v", i, err)
		}

		if _, duplicate := seen[id]; duplicate {
			t.Fatalf("duplicate ID: %s", id)
		}

		seen[id] = struct{}{}
	}
}

func TestRegisterCanceledContext(t *testing.T) {
	pr := NewProcessTracker()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := pr.Register(ctx, &mockKillable{})

	if err != context.Canceled {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestRegisterDeadlineExceeded(t *testing.T) {
	pr := NewProcessTracker()

	ctx, cancel := context.WithTimeout(t.Context(), time.Nanosecond)
	defer cancel()

	<-ctx.Done()

	_, err := pr.Register(ctx, &mockKillable{})

	if err != context.DeadlineExceeded {
		t.Fatalf("error = %v, want context.DeadlineExceeded", err)
	}
}

func TestRegisterBeforeStartBuffered(t *testing.T) {
	pr := NewProcessTracker()

	id, err := pr.Register(t.Context(), &mockKillable{})
	if err != nil {
		t.Fatalf("Register() before Start() error: %v", err)
	}

	if id == "" {
		t.Fatal("expected non-empty ID")
	}
}

func TestUnregisterSuccess(t *testing.T) {
	pr := NewProcessTracker()
	ctx := t.Context()
	pr.Start(ctx)

	id, err := pr.Register(ctx, &mockKillable{})
	if err != nil {
		t.Fatal(err)
	}

	if err := pr.Unregister(ctx, id); err != nil {
		t.Fatalf("Unregister() error: %v", err)
	}
}

func TestUnregisterNonexistentID(t *testing.T) {
	pr := NewProcessTracker()
	ctx := t.Context()
	pr.Start(ctx)

	if err := pr.Unregister(ctx, "missing"); err != nil {
		t.Fatalf("Unregister() error: %v", err)
	}
}

func TestUnregisterCanceledContext(t *testing.T) {
	pr := NewProcessTracker()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := pr.Unregister(ctx, "some-id")

	if err != context.Canceled {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestShutdownKillsAllRegistered(t *testing.T) {
	pr := NewProcessTracker()

	ctx, cancel := context.WithCancel(t.Context())
	pr.Start(ctx)

	procs := make([]*mockKillable, 5)

	for i := range procs {
		procs[i] = &mockKillable{}

		if _, err := pr.Register(ctx, procs[i]); err != nil {
			t.Fatalf("Register() #%d error: %v", i, err)
		}
	}

	time.Sleep(20 * time.Millisecond)
	cancel()

	for i, p := range procs {
		p := p

		waitFor(t, p.wasKilled)

		if !p.wasKilled() {
			t.Errorf("proc[%d] was not killed", i)
		}
	}
}

func TestShutdownSkipsUnregistered(t *testing.T) {
	pr := NewProcessTracker()

	ctx, cancel := context.WithCancel(t.Context())
	pr.Start(ctx)

	stay := &mockKillable{}
	die := &mockKillable{}

	stayID, err := pr.Register(ctx, stay)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := pr.Register(ctx, die); err != nil {
		t.Fatal(err)
	}

	if err := pr.Unregister(ctx, stayID); err != nil {
		t.Fatal(err)
	}

	time.Sleep(20 * time.Millisecond)
	cancel()

	waitFor(t, die.wasKilled)

	if stay.wasKilled() {
		t.Error("unregistered process was killed")
	}

	if !die.wasKilled() {
		t.Error("registered process was not killed")
	}
}

func TestShutdownEmptyRegistry(t *testing.T) {
	pr := NewProcessTracker()

	ctx, cancel := context.WithCancel(t.Context())
	pr.Start(ctx)

	cancel()
}

func TestShutdownIgnoresKillError(t *testing.T) {
	pr := NewProcessTracker()

	ctx, cancel := context.WithCancel(t.Context())
	pr.Start(ctx)

	p := &mockKillable{
		killErr: context.DeadlineExceeded,
	}

	if _, err := pr.Register(ctx, p); err != nil {
		t.Fatal(err)
	}

	time.Sleep(20 * time.Millisecond)
	cancel()

	waitFor(t, p.wasKilled)
}

func TestConcurrentRegisterUnregister(t *testing.T) {
	pr := NewProcessTracker()
	ctx := t.Context()
	pr.Start(ctx)

	const count = 100

	var wg sync.WaitGroup
	errCh := make(chan error, count*2)

	for range count {
		wg.Add(1)

		go func() {
			defer wg.Done()

			id, err := pr.Register(ctx, &mockKillable{})
			if err != nil {
				errCh <- err
				return
			}

			if err := pr.Unregister(ctx, id); err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent operation error: %v", err)
	}
}

func TestConcurrentRegisterDuringShutdown(t *testing.T) {
	pr := NewProcessTracker()

	ctx, cancel := context.WithCancel(t.Context())
	pr.Start(ctx)

	var wg sync.WaitGroup

	for range 20 {
		wg.Add(1)

		go func() {
			defer wg.Done()
			_, _ = pr.Register(ctx, &mockKillable{})
		}()
	}

	cancel()
	wg.Wait()
}

func TestRegisterUnregisterRegisterShutdown(t *testing.T) {
	pr := NewProcessTracker()

	ctx, cancel := context.WithCancel(t.Context())
	pr.Start(ctx)

	p1 := &mockKillable{}
	p2 := &mockKillable{}

	id1, err := pr.Register(ctx, p1)
	if err != nil {
		t.Fatal(err)
	}

	if err := pr.Unregister(ctx, id1); err != nil {
		t.Fatal(err)
	}

	if _, err := pr.Register(ctx, p2); err != nil {
		t.Fatal(err)
	}

	time.Sleep(20 * time.Millisecond)
	cancel()

	waitFor(t, p2.wasKilled)

	if p1.wasKilled() {
		t.Error("p1 was killed despite being unregistered")
	}

	if !p2.wasKilled() {
		t.Error("p2 was not killed despite being registered")
	}
}

func TestShutdownKillCalledOnce(t *testing.T) {
	pr := NewProcessTracker()

	ctx, cancel := context.WithCancel(t.Context())
	pr.Start(ctx)

	p := &mockKillable{}

	if _, err := pr.Register(ctx, p); err != nil {
		t.Fatal(err)
	}

	time.Sleep(20 * time.Millisecond)
	cancel()

	waitFor(t, p.wasKilled)

	if calls := p.killCalls.Load(); calls != 1 {
		t.Errorf("Kill() called %d times, want 1", calls)
	}
}
