package result

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergeResults_EmptyBatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	if err := mergeResults(path, 1024, validSchema(t), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("file should not be created for empty batch")
	}
}

func TestMergeResults_NewFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	results := []Result{
		newMockResult("1.1.1.1", 90, "1.1.1.1", "90"),
		newMockResult("2.2.2.2", 50, "2.2.2.2", "50"),
	}
	if err := mergeResults(path, 1024, validSchema(t), results); err != nil {
		t.Fatal(err)
	}
	rows := readCSVRows(t, path)
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}
	if rows[0][0] != "1.1.1.1" {
		t.Fatalf("want highest score first, got %s", rows[0][0])
	}
}

func TestMergeResults_SortedByScore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	results := []Result{
		newMockResult("a", 10, "a", "10"),
		newMockResult("b", 99, "b", "99"),
		newMockResult("c", 55, "c", "55"),
	}
	if err := mergeResults(path, 1024, validSchema(t), results); err != nil {
		t.Fatal(err)
	}
	rows := readCSVRows(t, path)
	want := []string{"b", "c", "a"}
	for i, r := range rows {
		if r[0] != want[i] {
			t.Fatalf("row %d: want %s got %s", i, want[i], r[0])
		}
	}
}

func TestMergeResults_AtomicReplacement(t *testing.T) {
	path := writeTempCSV(t, "a,100\n")
	delta := []Result{newMockResult("b", 200, "b", "200")}
	if err := mergeResults(path, 1024, validSchema(t), delta); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatal(".tmp file should be removed after atomic replace")
	}
}

// readCSVRows is a thin helper that reads the result file via ReadCSV.
func readCSVRows(t *testing.T, path string) [][]string {
	t.Helper()
	var rows [][]string
	err := ReadCSV(path, validSchema(t), func(r Result) error {
		rows = append(rows, r.ToRecord())
		return nil
	})
	if err != nil {
		t.Fatalf("readCSVRows: %v", err)
	}
	return rows
}
