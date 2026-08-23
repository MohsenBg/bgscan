package result

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"bgscan/internal/core/config"
)

type mockResult struct {
	key     string
	keyType KeyType
	score   float64
	record  []string
}

func newMockResult(key string, score float64, record ...string) *mockResult {
	return &mockResult{key: key, keyType: KeyIP, score: score, record: record}
}

func (m *mockResult) Key() string        { return m.key }
func (m *mockResult) KeyType() KeyType   { return m.keyType }
func (m *mockResult) ToRecord() []string { return m.record }
func (m *mockResult) Score() float64     { return m.score }

func (m *mockResult) Equal(other Result) bool {
	o, ok := other.(*mockResult)
	return ok && m.key == o.key && m.score == o.score
}

func validSchema(t *testing.T) ResultSchema {
	t.Helper()

	return ResultSchema{
		Name:      "test",
		Directory: "test",
		Columns:   []ColumnDef{{Name: "key", Width: 20}, {Name: "score", Width: 10}},
		Parser: func(record []string) (Result, error) {
			if len(record) < 2 {
				return nil, fmt.Errorf("expected 2 fields, got %d", len(record))
			}

			var score float64
			if _, err := fmt.Sscanf(record[1], "%f", &score); err != nil {
				return nil, fmt.Errorf("parse score %q: %w", record[1], err)
			}

			return newMockResult(record[0], score, record[0], record[1]), nil
		},
	}
}

func writeTempCSV(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "data.csv")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temporary CSV: %v", err)
	}

	return path
}

func defaultTestWriterConfig(t *testing.T) config.WriterConfig {
	t.Helper()

	cfg := config.DefaultWriterConfig()
	setBaseDir(t)
	return cfg
}

// setBaseDir redirects the application base directory to a fresh temp dir
// for the duration of the test.
func setBaseDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	old := baseDirOverride
	baseDirOverride = dir

	t.Cleanup(func() {
		baseDirOverride = old
	})

	return dir
}
