package result

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mockResult is a lightweight Result implementation used across all test files.
type mockResult struct {
	key     string
	keyType KeyType
	score   float64
	record  []string
}

func newMockResult(key string, score float64, record ...string) *mockResult {
	return &mockResult{
		key:     key,
		keyType: KeyIP,
		score:   score,
		record:  record,
	}
}

func (m *mockResult) Key() string        { return m.key }
func (m *mockResult) KeyType() KeyType   { return m.keyType }
func (m *mockResult) ToRecord() []string { return m.record }
func (m *mockResult) Equal(other Result) bool {
	o, ok := other.(*mockResult)
	if !ok {
		return false
	}
	return m.key == o.key && m.score == o.score
}
func (m *mockResult) Score() float64 { return m.score }

// --- Shared test helpers ---

// validSchema returns a ResultSchema whose parser recognises mockResult records
// with the format: key,score
func validSchema(t *testing.T) ResultSchema {
	t.Helper()
	return ResultSchema{
		Name:      "test",
		Directory: "test",
		Columns: []ColumnDef{
			{Name: "key", Width: 20},
			{Name: "score", Width: 10},
		},
		Parser: func(rec []string) (Result, error) {
			if len(rec) < 2 {
				return nil, fmt.Errorf("expected 2 fields, got %d", len(rec))
			}
			var score float64
			_, err := fmt.Sscanf(rec[1], "%f", &score)
			if err != nil {
				return nil, fmt.Errorf("invalid score %q: %w", rec[1], err)
			}
			return &mockResult{
				key:     rec[0],
				keyType: KeyIP,
				score:   score,
				record:  []string{rec[0], rec[1]},
			}, nil
		},
	}
}

// writeTempCSV creates a temporary CSV file with the given content lines
// and returns its path. The caller is responsible for cleaning up
// (use t.Cleanup).
func writeTempCSV(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "data.csv")
	err := os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("writeTempCSV: %v", err)
	}
	return path
}

// parseRecords parses CSV text (one record per line, comma-separated)
// into a slice of string slices.
func parseRecords(t *testing.T, csvText string) [][]string {
	t.Helper()
	var out [][]string
	for _, line := range strings.Split(strings.TrimSpace(csvText), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, strings.Split(line, ","))
	}
	return out
}

// defaultTestConfig returns a Config suitable for fast tests.
func defaultTestConfig() Config {
	return Config{
		MergeFlushInterval: MinMergeFlushInterval,
		ChanSize:           16,
		BatchSize:          4,
	}
}
