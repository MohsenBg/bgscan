package fileutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeLines(t *testing.T, path string, lines []string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create dir: %v", err)
	}
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func readLines(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	return strings.Split(strings.TrimRight(string(data), "\n"), "\n")
}

func isSorted(lines []string) bool {
	for i := 1; i < len(lines); i++ {
		if lines[i] < lines[i-1] {
			return false
		}
	}
	return true
}

func tmpPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(t.TempDir(), name)
}

func TestSortFile_SmallFile(t *testing.T) {
	input := tmpPath(t, "input.txt")
	output := tmpPath(t, "output.txt")

	lines := []string{"banana", "apple", "cherry", "date", "elderberry"}
	writeLines(t, input, lines)

	if err := SortFile(context.Background(), input, output); err != nil {
		t.Fatalf("SortFile: %v", err)
	}

	got := readLines(t, output)
	if !isSorted(got) {
		t.Errorf("output is not sorted: %v", got)
	}
	if len(got) != len(lines) {
		t.Errorf("line count mismatch: got %d, want %d", len(got), len(lines))
	}
}

func TestSortFile_AlreadySorted(t *testing.T) {
	input := tmpPath(t, "input.txt")
	output := tmpPath(t, "output.txt")

	lines := []string{"apple", "banana", "cherry"}
	writeLines(t, input, lines)

	if err := SortFile(context.Background(), input, output); err != nil {
		t.Fatalf("SortFile: %v", err)
	}

	got := readLines(t, output)
	for i, want := range lines {
		if got[i] != want {
			t.Errorf("line %d: got %q, want %q", i, got[i], want)
		}
	}
}

func TestSortFile_ReverseSorted(t *testing.T) {
	input := tmpPath(t, "input.txt")
	output := tmpPath(t, "output.txt")

	writeLines(t, input, []string{"z", "y", "x", "w", "v"})

	if err := SortFile(context.Background(), input, output); err != nil {
		t.Fatalf("SortFile: %v", err)
	}

	got := readLines(t, output)
	if !isSorted(got) {
		t.Errorf("output is not sorted: %v", got)
	}
}

func TestSortFile_SingleLine(t *testing.T) {
	input := tmpPath(t, "input.txt")
	output := tmpPath(t, "output.txt")

	writeLines(t, input, []string{"only"})

	if err := SortFile(context.Background(), input, output); err != nil {
		t.Fatalf("SortFile: %v", err)
	}

	got := readLines(t, output)
	if len(got) != 1 || got[0] != "only" {
		t.Errorf("got %v, want [only]", got)
	}
}

func TestSortFile_DuplicateLines(t *testing.T) {
	input := tmpPath(t, "input.txt")
	output := tmpPath(t, "output.txt")

	writeLines(t, input, []string{"b", "a", "b", "a", "c"})

	if err := SortFile(context.Background(), input, output); err != nil {
		t.Fatalf("SortFile: %v", err)
	}

	got := readLines(t, output)
	if !isSorted(got) {
		t.Errorf("output is not sorted: %v", got)
	}
	if len(got) != 5 {
		t.Errorf("expected 5 lines, got %d", len(got))
	}
}

func TestSortFile_CreatesOutputDir(t *testing.T) {
	input := tmpPath(t, "input.txt")
	// nested dir that does not exist yet
	output := filepath.Join(t.TempDir(), "nested", "deep", "output.txt")

	writeLines(t, input, []string{"b", "a"})

	if err := SortFile(context.Background(), input, output); err != nil {
		t.Fatalf("SortFile: %v", err)
	}

	if _, err := os.Stat(output); err != nil {
		t.Errorf("output file not created: %v", err)
	}
}

func TestSortFile_InputNotFound(t *testing.T) {
	err := SortFile(context.Background(), "/nonexistent/input.txt", tmpPath(t, "output.txt"))
	if err == nil {
		t.Fatal("expected error for missing input, got nil")
	}
}

func TestSortFile_CancelledContext(t *testing.T) {
	input := tmpPath(t, "input.txt")
	writeLines(t, input, []string{"b", "a"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := SortFile(ctx, input, tmpPath(t, "output.txt"))
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

func TestSortFile_ExternalSort(t *testing.T) {
	input := tmpPath(t, "input.txt")
	output := tmpPath(t, "output.txt")

	const n = 300_000
	lines := make([]string, n)
	for i := range lines {
		lines[i] = fmt.Sprintf("%010d", n-i)
	}
	writeLines(t, input, lines)

	if err := SortFile(context.Background(), input, output); err != nil {
		t.Fatalf("SortFile: %v", err)
	}

	got := readLines(t, output)
	if len(got) != n {
		t.Fatalf("line count mismatch: got %d, want %d", len(got), n)
	}
	if !isSorted(got) {
		t.Errorf("output is not sorted (checked %d lines)", n)
	}
}

func TestSortFile_PreservesAllLines(t *testing.T) {
	input := tmpPath(t, "input.txt")
	output := tmpPath(t, "output.txt")

	lines := []string{"delta", "alpha", "gamma", "beta", "alpha", "epsilon"}
	writeLines(t, input, lines)

	if err := SortFile(context.Background(), input, output); err != nil {
		t.Fatalf("SortFile: %v", err)
	}

	got := readLines(t, output)
	if len(got) != len(lines) {
		t.Errorf("line count: got %d, want %d", len(got), len(lines))
	}

	// verify every input line appears in output with correct multiplicity
	wantCounts := map[string]int{}
	for _, l := range lines {
		wantCounts[l]++
	}
	gotCounts := map[string]int{}
	for _, l := range got {
		gotCounts[l]++
	}
	for k, wc := range wantCounts {
		if gotCounts[k] != wc {
			t.Errorf("line %q: got count %d, want %d", k, gotCounts[k], wc)
		}
	}
}
