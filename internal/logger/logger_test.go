package logger

import (
	"strings"
	"testing"
	"time"
)

func newTestLogger(t *testing.T) *Logger {
	t.Helper()
	l, err := newLogger("test_" + t.Name())
	if err != nil {
		t.Fatalf("newLogger: %v", err)
	}
	t.Cleanup(func() {
		l.Close()
	})
	return l
}

func TestPubSub(t *testing.T) {
	l := newTestLogger(t)

	ch := l.Subscribe(10, 0)
	defer l.Unsubscribe(ch)

	l.write(LevelInfo, "hello %s", "world")

	select {
	case msg := <-ch:
		if !strings.Contains(msg, "hello world") {
			t.Fatalf("expected 'hello world' in message, got: %s", msg)
		}
		if !strings.Contains(msg, "[INFO]") {
			t.Fatalf("expected [INFO] level, got: %s", msg)
		}
		return
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestTail(t *testing.T) {
	l := newTestLogger(t)

	l.write(LevelInfo, "line one")
	l.write(LevelInfo, "line two")
	l.write(LevelInfo, "line three")

	ch := l.Subscribe(10, 2)
	defer l.Unsubscribe(ch)

	timeout := time.After(time.Second)
	count := 0
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				goto done
			}
			count++
			if count >= 2 {
				goto done
			}
		case <-timeout:
			t.Fatal("timed out waiting for tail messages")
		}
	}

done:
	if count < 1 {
		t.Fatalf("expected at least 1 tail message, got %d", count)
	}
}

func TestLifecycle(t *testing.T) {
	l := newTestLogger(t)

	ch := l.Subscribe(10, 0)

	l.write(LevelInfo, "before close")

	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("timed out")
	}

	l.Close()

	// Close() writes "session ended" before closing channels
	<-ch

	_, ok := <-ch
	if ok {
		t.Fatal("expected channel to be closed after Close()")
	}
}

func TestUnsubscribe(t *testing.T) {
	l := newTestLogger(t)

	ch := l.Subscribe(10, 0)
	l.Unsubscribe(ch)

	_, ok := <-ch
	if ok {
		t.Fatal("expected channel to be closed after Unsubscribe()")
	}
}

func TestEnableDisable(t *testing.T) {
	l := newTestLogger(t)

	ch := l.Subscribe(10, 0)
	defer l.Unsubscribe(ch)

	l.Disable()
	l.write(LevelInfo, "should not appear")

	select {
	case <-ch:
		t.Fatal("received message while disabled")
	case <-time.After(50 * time.Millisecond):
		// expected
	}

	l.Enable()
	l.write(LevelInfo, "should appear")

	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message after enable")
	}
}

func TestLogLevelString(t *testing.T) {
	cases := []struct {
		level LogLevel
		want  string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarning, "WARN"},
		{LevelError, "ERROR"},
	}

	for _, tc := range cases {
		if got := tc.level.String(); got != tc.want {
			t.Errorf("LogLevel(%d).String() = %q, want %q", tc.level, got, tc.want)
		}
	}
}

func TestMultipleSubscribers(t *testing.T) {
	l := newTestLogger(t)

	ch1 := l.Subscribe(10, 0)
	ch2 := l.Subscribe(10, 0)
	defer l.Unsubscribe(ch1)
	defer l.Unsubscribe(ch2)

	l.write(LevelInfo, "broadcast")

	for _, ch := range []chan string{ch1, ch2} {
		select {
		case msg := <-ch:
			if !strings.Contains(msg, "broadcast") {
				t.Fatalf("expected 'broadcast' in message, got: %s", msg)
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for broadcast message")
		}
	}
}
