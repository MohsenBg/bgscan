package result

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewWriter_NilContextUsesBackground(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "results.csv")
	schema := validSchema(t)
	cfg := defaultTestConfig()

	w, err := NewWriter(path, schema, cfg, nil)
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}
	if w == nil {
		t.Fatal("NewWriter() returned nil writer")
	}
	w.cancel()
}

func TestNewWriter_NormalizeConfig(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "results.csv")
	schema := validSchema(t)

	cfg := Config{MergeFlushInterval: 0, ChanSize: -1, BatchSize: 0}

	w, err := NewWriter(path, schema, cfg, context.Background())
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}
	if w.batchSize != DefaultBatchSize {
		t.Errorf("batchSize = %d, want %d after Normalize", w.batchSize, DefaultBatchSize)
	}
	w.cancel()
}

func TestWriter_StartStop_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "results.csv")
	schema := validSchema(t)

	w, _ := NewWriter(path, schema, defaultTestConfig(), context.Background())

	w.Start()

	err := w.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestWriter_StopWithoutStart(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "results.csv")
	schema := validSchema(t)

	w, _ := NewWriter(path, schema, defaultTestConfig(), context.Background())

	err := w.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestWriter_Write_FlushesOnStop(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "results.csv")
	schema := validSchema(t)

	cfg := defaultTestConfig()
	w, _ := NewWriter(path, schema, cfg, context.Background())
	w.Start()

	w.Write(newMockResult("1.2.3.4", 0.9, "1.2.3.4", "0.9"))
	w.Write(newMockResult("5.6.7.8", 0.5, "5.6.7.8", "0.5"))

	_ = w.Stop()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	text := strings.TrimSpace(string(data))
	if !strings.Contains(text, "1.2.3.4") {
		t.Errorf("output missing '1.2.3.4': %q", text)
	}
	if !strings.Contains(text, "5.6.7.8") {
		t.Errorf("output missing '5.6.7.8': %q", text)
	}
}

func TestWriter_Write_ScoreSorted(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "results.csv")
	schema := validSchema(t)

	w, _ := NewWriter(path, schema, defaultTestConfig(), context.Background())
	w.Start()

	w.Write(newMockResult("c", 0.1, "c", "0.1"))
	w.Write(newMockResult("a", 0.9, "a", "0.9"))
	w.Write(newMockResult("b", 0.5, "b", "0.5"))

	_ = w.Stop()

	data, _ := os.ReadFile(path)
	text := strings.TrimSpace(string(data))
	lines := strings.Split(text, "\n")

	if len(lines) < 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "a,") {
		t.Errorf("first line should be 'a' (score 0.9), got: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "b,") {
		t.Errorf("second line should be 'b' (score 0.5), got: %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "c,") {
		t.Errorf("third line should be 'c' (score 0.1), got: %q", lines[2])
	}
}

func TestWriter_Write_AfterStopIsDropped(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "results.csv")
	schema := validSchema(t)

	w, _ := NewWriter(path, schema, defaultTestConfig(), context.Background())
	w.Start()
	_ = w.Stop()

	w.Write(newMockResult("1.2.3.4", 0.9, "1.2.3.4", "0.9"))

	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	if strings.TrimSpace(string(data)) != "" {
		t.Errorf("file should be empty after write post-stop, got: %q", string(data))
	}
}

func TestWriter_ConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "results.csv")
	schema := validSchema(t)

	cfg := Config{
		MergeFlushInterval: 10 * time.Second,
		ChanSize:           1000,
		BatchSize:          500,
	}

	w, _ := NewWriter(path, schema, cfg, context.Background())
	w.Start()

	var wg sync.WaitGroup
	const goroutines = 10
	const writesPer = 50

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < writesPer; i++ {
				key := "10.0.0.1"
				w.Write(newMockResult(key, 0.5, key, "0.5"))
			}
			_ = gid
		}(g)
	}

	wg.Wait()
	_ = w.Stop()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		t.Error("expected some output after concurrent writes")
	}
}

func TestWriter_FlushOnBatchSize(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "results.csv")
	schema := validSchema(t)

	cfg := Config{
		MergeFlushInterval: 10 * time.Second,
		ChanSize:           100,
		BatchSize:          3,
	}

	w, _ := NewWriter(path, schema, cfg, context.Background())
	w.Start()

	w.Write(newMockResult("a", 0.1, "a", "0.1"))
	w.Write(newMockResult("b", 0.2, "b", "0.2"))
	w.Write(newMockResult("c", 0.3, "c", "0.3"))

	time.Sleep(200 * time.Millisecond)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile after batch flush: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected data after batch-size flush")
	}

	_ = w.Stop()
}

func TestWriter_GetResultPath_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "results.csv")
	schema := validSchema(t)

	w, _ := NewWriter(path, schema, defaultTestConfig(), context.Background())
	w.Start()
	w.Write(newMockResult("1.2.3.4", 0.9, "1.2.3.4", "0.9"))
	_ = w.Stop()

	got := w.GetResultPath()
	if got != path {
		t.Errorf("GetResultPath() = %q, want %q", got, path)
	}
}

func TestWriter_GetResultPath_FileNotYetCreated(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "results.csv")
	schema := validSchema(t)

	w, _ := NewWriter(path, schema, defaultTestConfig(), context.Background())

	got := w.GetResultPath()
	if got != "" {
		t.Errorf("GetResultPath() = %q, want empty string", got)
	}
}

func TestWriter_Write_EmptyResultsNoFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "results.csv")
	schema := validSchema(t)

	w, _ := NewWriter(path, schema, defaultTestConfig(), context.Background())
	w.Start()
	_ = w.Stop()

	got := w.GetResultPath()
	if got != "" {
		t.Errorf("GetResultPath() = %q, want empty (no results written)", got)
	}
}

func TestWriter_ContextCancel_StopsGracefully(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "results.csv")
	schema := validSchema(t)

	w, _ := NewWriter(path, schema, defaultTestConfig(), ctx)
	w.Start()

	w.Write(newMockResult("1.2.3.4", 0.9, "1.2.3.4", "0.9"))

	cancel()

	err := w.Stop()
	if err != nil {
		t.Errorf("Stop() after context cancel: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected file after ctx cancel + stop: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected data in file after drain")
	}
}
