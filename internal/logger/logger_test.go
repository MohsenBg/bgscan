package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestLogger(t *testing.T) *Logger {
	t.Helper()
	l, err := New("test.log", WithDir(t.TempDir()))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(l.Close)
	return l
}

func TestPubSub(t *testing.T) {
	l := newTestLogger(t)

	ch := l.Subscribe(10, 0)
	defer l.Unsubscribe(ch)

	l.Info("hello %s", "world")

	select {
	case msg := <-ch:
		if !strings.Contains(msg, "hello world") {
			t.Fatalf("expected 'hello world' in message, got: %s", msg)
		}
		if !strings.Contains(msg, "[INFO]") {
			t.Fatalf("expected [INFO] level, got: %s", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestTail(t *testing.T) {
	l := newTestLogger(t)

	l.Info("line one")
	l.Info("line two")
	l.Info("line three")

	ch := l.Subscribe(10, 2)
	defer l.Unsubscribe(ch)

	timeout := time.After(time.Second)
	count := 0
	for count < 2 {
		select {
		case _, ok := <-ch:
			if !ok {
				break
			}
			count++
		case <-timeout:
			t.Fatal("timed out waiting for tail messages")
		}
	}

	if count < 1 {
		t.Fatalf("expected at least 1 tail message, got %d", count)
	}
}

func TestLifecycle(t *testing.T) {
	l := newTestLogger(t)

	ch := l.Subscribe(10, 0)

	l.Info("before close")

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

	l.Close() // idempotent
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

func TestLevelDial(t *testing.T) {
	l := newTestLogger(t)

	ch := l.Subscribe(10, 0)
	defer l.Unsubscribe(ch)

	l.SetLevel(slog.LevelWarn)
	l.Info("filtered out")

	select {
	case <-ch:
		t.Fatal("received INFO while dial was at WARN")
	case <-time.After(50 * time.Millisecond):
	}

	l.Warn("passes warn dial")

	select {
	case msg := <-ch:
		if !strings.Contains(msg, "[WARN]") {
			t.Fatalf("expected [WARN], got: %s", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for WARN message")
	}

	l.SetLevel(slog.LevelDebug)
	l.Debug("debug now visible")

	select {
	case msg := <-ch:
		if !strings.Contains(msg, "[DEBUG]") {
			t.Fatalf("expected [DEBUG], got: %s", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for DEBUG message")
	}
}

func TestWithSharesFileAndDial(t *testing.T) {
	l := newTestLogger(t)

	child := l.With("probe", "slipstream")

	ch := l.Subscribe(10, 0)
	defer l.Unsubscribe(ch)

	child.Info("tagged line")

	select {
	case msg := <-ch:
		if !strings.Contains(msg, "probe=slipstream") {
			t.Fatalf("expected probe=slipstream tag, got: %s", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for child logger message")
	}

	// dial moved on the parent must filter the child too
	l.SetLevel(slog.LevelError)
	child.Info("filtered via shared dial")

	select {
	case <-ch:
		t.Fatal("child wrote INFO while shared dial was at ERROR")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestNilLoggerIsSafe(t *testing.T) {
	var l *Logger
	l.Info("no panic")
	l.Error("no panic")
	l.Debug("no panic")
	l.Warn("no panic")
	l.Dump("no panic", struct{}{})
	l.SetLevel(slog.LevelDebug)
	l.Close()

	ch := l.Subscribe(4, 2)
	if _, ok := <-ch; ok {
		t.Fatal("nil logger subscriber must be closed")
	}
	l.Unsubscribe(ch)
}

func TestFileFormat(t *testing.T) {
	dir := t.TempDir()
	l, err := New("fmt.log", WithDir(dir))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	l.Info("plain message")
	l.Close()

	data, err := readFileString(filepath.Join(dir, "fmt.log"))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}

	// session start + the message + session end
	if !strings.Contains(data, "[INFO] plain message") {
		t.Fatalf("expected formatted line, got:\n%s", data)
	}
	if !strings.Contains(data, sessionStart) {
		t.Fatalf("expected session start marker, got:\n%s", data)
	}
	// timestamp shape: 2006/01/02 15:04:05
	if !strings.Contains(data, " [INFO] plain message") {
		t.Fatalf("expected timestamp-prefixed line, got:\n%s", data)
	}
}

func TestMultipleSubscribers(t *testing.T) {
	l := newTestLogger(t)

	ch1 := l.Subscribe(10, 0)
	ch2 := l.Subscribe(10, 0)
	defer l.Unsubscribe(ch1)
	defer l.Unsubscribe(ch2)

	l.Info("broadcast")

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

func TestDiscardSet(t *testing.T) {
	s := DiscardSet()
	s.Core.Info("writes nowhere")
	s.UI.Error("writes nowhere")
	s.Debug.Dump("nope", map[string]int{"a": 1})

	ch := s.Core.Subscribe(4, 2)
	if _, ok := <-ch; ok {
		t.Fatal("discard subscriber must be closed")
	}
	s.Close()
}

func TestParseLevel(t *testing.T) {
	for name, want := range map[string]slog.Level{
		"debug":   slog.LevelDebug,
		"INFO":    slog.LevelInfo,
		" warn":   slog.LevelWarn,
		"warning": slog.LevelWarn,
		"error":   slog.LevelError,
	} {
		got, err := ParseLevel(name)
		if err != nil {
			t.Errorf("ParseLevel(%q): %v", name, err)
			continue
		}
		if got != want {
			t.Errorf("ParseLevel(%q) = %v, want %v", name, got, want)
		}
	}

	if _, err := ParseLevel("bogus"); err == nil {
		t.Error("ParseLevel(\"bogus\") = nil, want error")
	}
}

func TestSetLevelAppliesToEveryLogger(t *testing.T) {
	s := DiscardSet()

	s.SetLevel(slog.LevelDebug)
	for name, l := range map[string]*Logger{"core": s.Core, "ui": s.UI, "debug": s.Debug} {
		if got := l.Level(); got != slog.LevelDebug {
			t.Errorf("%s level = %v, want debug", name, got)
		}
	}

	s.SetLevel(slog.LevelError)
	for name, l := range map[string]*Logger{"core": s.Core, "ui": s.UI, "debug": s.Debug} {
		if got := l.Level(); got != slog.LevelError {
			t.Errorf("%s level = %v, want error", name, got)
		}
	}

	s.Close()
}

func readFileString(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}
