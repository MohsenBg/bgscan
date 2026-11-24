package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestStreamCSVStrict_AbortsOnBadRow covers stream csvstrict aborts on bad row.
func TestStreamCSVStrict_AbortsOnBadRow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "strict.csv")

	if err := os.WriteFile(path, []byte("a,b\n\"unclosed,1\n"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	var rows int

	err := StreamCSVStrict(path, CSVConfig{}, func(rec []string, line int64) error {
		rows++
		return nil
	})
	if err == nil {
		t.Fatal("expected a parse error, got nil")
	}
	if rows != 1 {
		t.Fatalf("rows before error = %d, want 1", rows)
	}
}

// TestStreamCSVStrict_ReportsPhysicalLine covers stream csvstrict reports physical line.
func TestStreamCSVStrict_ReportsPhysicalLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "strict.csv")

	content := "value,port\nok,1\n\"bad\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	err := StreamCSVStrict(path, CSVConfig{HasHeader: true}, func(rec []string, line int64) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected a parse error, got nil")
	}
	if !strings.Contains(err.Error(), "line 3") {
		t.Fatalf("error = %q, want a line 3 reference", err)
	}
}

// TestStreamCSVStrict_SkipsBlankLines covers stream csvstrict skips blank lines.
func TestStreamCSVStrict_SkipsBlankLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "strict.csv")

	content := "a,b\n\nc,d\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	var rows int
	var lastLine int64

	err := StreamCSVStrict(path, CSVConfig{}, func(rec []string, line int64) error {
		rows++
		lastLine = line
		return nil
	})
	if err != nil {
		t.Fatalf("StreamCSVStrict: %v", err)
	}
	if rows != 2 {
		t.Fatalf("rows = %d, want 2", rows)
	}
	if lastLine != 2 {
		t.Fatalf("last record line = %d, want 2", lastLine)
	}
}

// TestStreamCSVWithOffset_HonorsHeader covers stream csvwith offset honors header.
func TestStreamCSVWithOffset_HonorsHeader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "indexed.csv")

	content := "value,port\n1.1.1.1,80\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	var records [][]string
	var offsets []int64

	err := StreamCSVWithOffset(path, CSVConfig{HasHeader: true}, func(rec []string, offset int64) error {
		records = append(records, rec)
		offsets = append(offsets, offset)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamCSVWithOffset: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("records = %d, want 1", len(records))
	}
	if got := strings.Join(records[0], ","); got != "1.1.1.1,80" {
		t.Fatalf("record = %q, want %q", got, "1.1.1.1,80")
	}

	wantOffset := int64(len("value,port\n"))
	if offsets[0] != wantOffset {
		t.Fatalf("offset = %d, want %d", offsets[0], wantOffset)
	}
}
