package result

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadCSV_ValidRecords(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\n5.6.7.8,0.5\n")
	schema := validSchema(t)

	var results []Result
	err := ReadCSV(path, schema, func(r Result) error {
		results = append(results, r)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadCSV() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Key() != "1.2.3.4" {
		t.Errorf("results[0].Key() = %q, want %q", results[0].Key(), "1.2.3.4")
	}
	if results[1].Key() != "5.6.7.8" {
		t.Errorf("results[1].Key() = %q, want %q", results[1].Key(), "5.6.7.8")
	}
}

func TestReadCSV_SkipsParseErrors(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\nbad_line\n5.6.7.8,0.5\n")
	schema := validSchema(t)

	var count int
	err := ReadCSV(path, schema, func(r Result) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("ReadCSV() error = %v", err)
	}
	if count != 2 {
		t.Errorf("got %d valid results, want 2", count)
	}
}

func TestReadCSV_CallbackError_StopsIteration(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\n5.6.7.8,0.5\n10.0.0.1,0.3\n")
	schema := validSchema(t)

	var count int
	err := ReadCSV(path, schema, func(r Result) error {
		count++
		if count == 2 {
			return os.ErrClosed
		}
		return nil
	})
	if err == nil {
		t.Fatal("expected error from callback, got nil")
	}
	if count != 2 {
		t.Errorf("expected callback to be called 2 times before error, got %d", count)
	}
}

func TestReadCSV_EmptyFile(t *testing.T) {
	path := writeTempCSV(t, "")
	schema := validSchema(t)

	var count int
	err := ReadCSV(path, schema, func(r Result) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("ReadCSV() error = %v", err)
	}
	if count != 0 {
		t.Errorf("got %d results from empty file, want 0", count)
	}
}

func TestReadCSV_NonexistentFile(t *testing.T) {
	schema := validSchema(t)
	err := ReadCSV("/nonexistent/path.csv", schema, func(r Result) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestReadCSV_SingleFieldLine(t *testing.T) {
	// Parser expects 2 fields; single field should be skipped.
	path := writeTempCSV(t, "1.2.3.4\n")
	schema := validSchema(t)

	var count int
	err := ReadCSV(path, schema, func(r Result) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("ReadCSV() error = %v", err)
	}
	if count != 0 {
		t.Errorf("got %d results, want 0 (parser should reject 1-field record)", count)
	}
}

func TestReadCSV_EmptyLines(t *testing.T) {
	path := writeTempCSV(t, "\n\n\n")
	schema := validSchema(t)

	var count int
	err := ReadCSV(path, schema, func(r Result) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("ReadCSV() error = %v", err)
	}
	if count != 0 {
		t.Errorf("got %d results from empty lines, want 0", count)
	}
}

// ---------- StreamWriteResults ----------

func TestStreamWriteResults_WritesCSV(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "out.csv")

	results := []*mockResult{
		newMockResult("1.2.3.4", 0.9, "1.2.3.4", "0.9"),
		newMockResult("5.6.7.8", 0.5, "5.6.7.8", "0.5"),
	}

	err := StreamWriteResults(path, func(emit func(Result) error) error {
		for _, r := range results {
			if err := emit(r); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("StreamWriteResults() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	text := strings.TrimSpace(string(data))
	lines := strings.Split(text, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), text)
	}
	if !strings.Contains(lines[0], "1.2.3.4") {
		t.Errorf("line 0 missing '1.2.3.4': %q", lines[0])
	}
	if !strings.Contains(lines[1], "5.6.7.8") {
		t.Errorf("line 1 missing '5.6.7.8': %q", lines[1])
	}
}

func TestStreamWriteResults_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "out.csv")

	err := StreamWriteResults(path, func(emit func(Result) error) error {
		// emit nothing
		return nil
	})
	if err != nil {
		t.Fatalf("StreamWriteResults() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("expected empty file, got %q", string(data))
	}
}

func TestStreamWriteResults_EmitError(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "out.csv")

	err := StreamWriteResults(path, func(emit func(Result) error) error {
		return os.ErrClosed
	})
	if err != os.ErrClosed {
		t.Errorf("expected os.ErrClosed, got %v", err)
	}
}

func TestStreamWriteResults_NonexistentDir(t *testing.T) {
	path := "/nonexistent/deeply/nested/dir/out.csv"

	err := StreamWriteResults(path, func(emit func(Result) error) error {
		return emit(newMockResult("1.2.3.4", 0.5, "1.2.3.4", "0.5"))
	})
	if err == nil {
		t.Fatal("expected error for nonexistent directory, got nil")
	}
}

func TestStreamWriteResults_PreservesRecordOrder(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "out.csv")

	results := []*mockResult{
		newMockResult("a", 0.1, "a", "0.1"),
		newMockResult("b", 0.2, "b", "0.2"),
		newMockResult("c", 0.3, "c", "0.3"),
	}

	_ = StreamWriteResults(path, func(emit func(Result) error) error {
		for _, r := range results {
			_ = emit(r)
		}
		return nil
	})

	data, _ := os.ReadFile(path)
	text := strings.TrimSpace(string(data))
	lines := strings.Split(text, "\n")

	expected := []string{"a,0.1", "b,0.2", "c,0.3"}
	for i, want := range expected {
		if i >= len(lines) {
			t.Errorf("missing line %d", i)
			continue
		}
		if strings.TrimSpace(lines[i]) != want {
			t.Errorf("line %d = %q, want %q", i, strings.TrimSpace(lines[i]), want)
		}
	}
}
