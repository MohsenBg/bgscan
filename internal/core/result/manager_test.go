package result

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"bgscan/internal/core/config"
)

func newWriterForTest(t *testing.T, ctx context.Context, cfg config.WriterConfig) Writer {
	t.Helper()

	w, err := NewWriter(ctx, WriterOptions{
		ResultPrefix: "test_",
		Schema:       validSchema(t),
		Config:       cfg,
	})
	if err != nil {
		t.Fatalf("NewWriter() error: %v", err)
	}

	return w
}

func TestNewWriter_NilContextUsesBackground(t *testing.T) {
	w := newWriterForTest(t, nil, defaultTestWriterConfig(t))
	if err := w.Stop(); err != nil {
		t.Fatalf("Stop() error: %v", err)
	}
}

func TestNewWriter_RejectsInvalidConfig(t *testing.T) {
	_, err := NewWriter(context.Background(), WriterOptions{
		ResultPrefix: "test_",
		Schema:       validSchema(t),
		Config:       config.WriterConfig{},
	})
	if err == nil {
		t.Fatal("expected invalid writer config error")
	}
}

func TestNewWriter_RejectsInvalidSchema(t *testing.T) {
	_, err := NewWriter(context.Background(), WriterOptions{
		ResultPrefix: "test_",
		Config:       defaultTestWriterConfig(t),
	})
	if err == nil {
		t.Fatal("expected invalid schema error")
	}
}

func TestWriter_StartAndStopAreIdempotent(t *testing.T) {
	w := newWriterForTest(t, context.Background(), defaultTestWriterConfig(t))

	if err := w.Start(); err != nil {
		t.Fatalf("first Start() error: %v", err)
	}
	if err := w.Start(); err != nil {
		t.Fatalf("second Start() error: %v", err)
	}
	if err := w.Stop(); err != nil {
		t.Fatalf("first Stop() error: %v", err)
	}
	if err := w.Stop(); err != nil {
		t.Fatalf("second Stop() error: %v", err)
	}
}

func TestWriter_StopWithoutStart(t *testing.T) {
	w := newWriterForTest(t, context.Background(), defaultTestWriterConfig(t))

	if err := w.Stop(); err != nil {
		t.Fatalf("Stop() error: %v", err)
	}
}

func TestWriter_WriteFlushesOnStop(t *testing.T) {
	w := newWriterForTest(t, context.Background(), defaultTestWriterConfig(t))
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}

	w.Write(newMockResult("1.2.3.4", 0.9, "1.2.3.4", "0.9"))
	w.Write(newMockResult("5.6.7.8", 0.5, "5.6.7.8", "0.5"))
	if err := w.Stop(); err != nil {
		t.Fatal(err)
	}

	path := w.GetResultPath()
	if path == "" {
		t.Fatal("GetResultPath() returned an empty path")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read result file: %v", err)
	}
	if text := string(data); !strings.Contains(text, "1.2.3.4") || !strings.Contains(text, "5.6.7.8") {
		t.Fatalf("result file is missing records: %q", text)
	}
}

func TestWriter_SortsResultsByScore(t *testing.T) {
	w := newWriterForTest(t, context.Background(), defaultTestWriterConfig(t))
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}

	w.Write(newMockResult("c", 0.1, "c", "0.1"))
	w.Write(newMockResult("a", 0.9, "a", "0.9"))
	w.Write(newMockResult("b", 0.5, "b", "0.5"))
	if err := w.Stop(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(w.GetResultPath())
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	want := []string{"a,", "b,", "c,"}
	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d", len(lines), len(want))
	}
	for i, prefix := range want {
		if !strings.HasPrefix(lines[i], prefix) {
			t.Errorf("line %d = %q, want prefix %q", i, lines[i], prefix)
		}
	}
}

func TestWriter_WriteAfterStopIsDropped(t *testing.T) {
	w := newWriterForTest(t, context.Background(), defaultTestWriterConfig(t))
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}
	if err := w.Stop(); err != nil {
		t.Fatal(err)
	}

	w.Write(newMockResult("1.2.3.4", 0.9, "1.2.3.4", "0.9"))
	if path := w.GetResultPath(); path != "" {
		t.Errorf("GetResultPath() = %q, want empty path", path)
	}
}

func TestWriter_ConcurrentWrites(t *testing.T) {
	w := newWriterForTest(t, context.Background(), defaultTestWriterConfig(t))
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				w.Write(newMockResult("10.0.0.1", 0.5, "10.0.0.1", "0.5"))
			}
		}()
	}
	wg.Wait()

	if err := w.Stop(); err != nil {
		t.Fatal(err)
	}
	if path := w.GetResultPath(); path == "" {
		t.Fatal("expected result file after concurrent writes")
	}
}

func TestWriter_ContextCancelFlushesQueuedResults(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	w := newWriterForTest(t, ctx, defaultTestWriterConfig(t))
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}

	w.Write(newMockResult("1.2.3.4", 0.9, "1.2.3.4", "0.9"))
	cancel()

	if err := w.Stop(); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(time.Second)
	for w.GetResultPath() == "" && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if path := w.GetResultPath(); path == "" {
		t.Fatal("expected result file after context cancellation")
	}
}
