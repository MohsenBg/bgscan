package result

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergeResults_EmptyBatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	if err := mergeResults(nil, path, validSchema(t), nil); err != nil {
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
	if err := mergeResults(nil, path, validSchema(t), results); err != nil {
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
	if err := mergeResults(nil, path, validSchema(t), results); err != nil {
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
	if err := mergeResults(nil, path, validSchema(t), delta); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatal(".tmp file should be removed after atomic replace")
	}
}

func TestMergeResults_DedupWithinBatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	results := []Result{
		newMockResult("a", 9, "a", "9"),
		newMockResult("a", 7, "a", "7"),
		newMockResult("b", 5, "b", "5"),
	}
	if err := mergeResults(nil, path, validSchema(t), results); err != nil {
		t.Fatal(err)
	}
	rows := readCSVRows(t, path)
	if len(rows) != 2 {
		t.Fatalf("want 2 deduped rows, got %d: %v", len(rows), rows)
	}
	if rows[0][0] != "a" || rows[0][1] != "9" {
		t.Fatalf("want highest score for a, got %s:%s", rows[0][0], rows[0][1])
	}
}

// TestMergeResults_DedupInterleavedWithinBatch is the regression test for
// batches where duplicates are separated in score order (a:9, b:5, a:3).
// Deduplication must happen independently of score sorting and keep the
// highest-scored record for each key.
func TestMergeResults_DedupInterleavedWithinBatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	results := []Result{
		newMockResult("a", 9, "a", "9"),
		newMockResult("b", 5, "b", "5"),
		newMockResult("a", 3, "a", "3"),
	}
	if err := mergeResults(nil, path, validSchema(t), results); err != nil {
		t.Fatal(err)
	}

	rows := readCSVRows(t, path)
	want := [][]string{{"a", "9"}, {"b", "5"}}
	if len(rows) != len(want) {
		t.Fatalf("want %v, got %v", want, rows)
	}
	for i, w := range want {
		if rows[i][0] != w[0] || rows[i][1] != w[1] {
			t.Fatalf("row %d: want %v got %v", i, w, rows[i])
		}
	}
}

// TestMergeResults_DedupInterleavedWithExistingFile exercises the merge lookup
// with an interleaved duplicate batch. The lookup must see unique keys and
// keep the highest-scored record for each key.
func TestMergeResults_DedupInterleavedWithExistingFile(t *testing.T) {
	path := writeTempCSV(t, "b,7\n")
	results := []Result{
		newMockResult("a", 9, "a", "9"),
		newMockResult("b", 5, "b", "5"),
		newMockResult("a", 3, "a", "3"),
	}
	if err := mergeResults(nil, path, validSchema(t), results); err != nil {
		t.Fatal(err)
	}

	rows := readCSVRows(t, path)
	want := [][]string{{"a", "9"}, {"b", "7"}}
	if len(rows) != len(want) {
		t.Fatalf("want %v, got %v", want, rows)
	}
	for i, w := range want {
		if rows[i][0] != w[0] || rows[i][1] != w[1] {
			t.Fatalf("row %d: want %v got %v", i, w, rows[i])
		}
	}
}

func TestMergeResults_ReplaceExistingWithHigherScore(t *testing.T) {
	path := writeTempCSV(t, "a,5\n")
	results := []Result{newMockResult("a", 9, "a", "9")}
	if err := mergeResults(nil, path, validSchema(t), results); err != nil {
		t.Fatal(err)
	}
	rows := readCSVRows(t, path)
	if len(rows) != 1 || rows[0][0] != "a" || rows[0][1] != "9" {
		t.Fatalf("want single row a:9, got %v", rows)
	}
}

// TestMergeResults_HigherIncomingScoreWins: existing A:70, incoming A:90 must
// yield exactly one A:90 row, with the file kept sorted by descending score.
func TestMergeResults_HigherIncomingScoreWins(t *testing.T) {
	path := writeTempCSV(t, "b,95\na,70\n")
	delta := []Result{newMockResult("a", 90, "a", "90")}
	if err := mergeResults(nil, path, validSchema(t), delta); err != nil {
		t.Fatal(err)
	}

	rows := readCSVRows(t, path)
	want := [][]string{{"b", "95"}, {"a", "90"}}
	if len(rows) != len(want) {
		t.Fatalf("want %v, got %v", want, rows)
	}
	for i, w := range want {
		if rows[i][0] != w[0] || rows[i][1] != w[1] {
			t.Fatalf("row %d: want %v got %v", i, w, rows[i])
		}
	}
}

// TestMergeResults_HigherExistingScoreWins: existing A:90, incoming A:70 must
// keep A:90 exactly once instead of regressing to the lower score.
func TestMergeResults_HigherExistingScoreWins(t *testing.T) {
	path := writeTempCSV(t, "b,95\na,90\n")
	delta := []Result{newMockResult("a", 70, "a", "70")}
	if err := mergeResults(nil, path, validSchema(t), delta); err != nil {
		t.Fatal(err)
	}

	rows := readCSVRows(t, path)
	want := [][]string{{"b", "95"}, {"a", "90"}}
	if len(rows) != len(want) {
		t.Fatalf("want %v, got %v", want, rows)
	}
	for i, w := range want {
		if rows[i][0] != w[0] || rows[i][1] != w[1] {
			t.Fatalf("row %d: want %v got %v", i, w, rows[i])
		}
	}
}

// TestMergeResults_EqualScoreKeepsExisting: existing A:80, incoming A:80 must
// preserve the existing record exactly once, without duplicates.
func TestMergeResults_EqualScoreKeepsExisting(t *testing.T) {
	path := writeTempCSV(t, "b,95\na,80\n")
	delta := []Result{newMockResult("a", 80, "a", "80")}
	if err := mergeResults(nil, path, validSchema(t), delta); err != nil {
		t.Fatal(err)
	}

	rows := readCSVRows(t, path)
	want := [][]string{{"b", "95"}, {"a", "80"}}
	if len(rows) != len(want) {
		t.Fatalf("want %v, got %v", want, rows)
	}
	for i, w := range want {
		if rows[i][0] != w[0] || rows[i][1] != w[1] {
			t.Fatalf("row %d: want %v got %v", i, w, rows[i])
		}
	}
}

// TestMergeResults_KeepExistingWhenIncomingLower verifies that an incoming
// record with a lower score never replaces the higher-scoring existing record.
func TestMergeResults_KeepExistingWhenIncomingLower(t *testing.T) {
	path := writeTempCSV(t, "a,5\nb,1\n")
	results := []Result{newMockResult("a", 3, "a", "3")}
	if err := mergeResults(nil, path, validSchema(t), results); err != nil {
		t.Fatal(err)
	}
	rows := readCSVRows(t, path)
	want := [][]string{{"a", "5"}, {"b", "1"}}
	if len(rows) != len(want) {
		t.Fatalf("want %v, got %v", want, rows)
	}
	for i, w := range want {
		if rows[i][0] != w[0] || rows[i][1] != w[1] {
			t.Fatalf("row %d: want %v got %v", i, w, rows[i])
		}
	}
}

func TestMergeResults_LookupOutsideScorePrefix(t *testing.T) {
	// The matching batch record sits between two other records in score
	// order, so the positional boundary check of the old implementation
	// could never see it. The existing record keeps the higher score.
	path := writeTempCSV(t, "a,5\n")
	results := []Result{
		newMockResult("b", 9, "b", "9"),
		newMockResult("a", 3, "a", "3"),
	}
	if err := mergeResults(nil, path, validSchema(t), results); err != nil {
		t.Fatal(err)
	}
	rows := readCSVRows(t, path)
	want := [][]string{{"b", "9"}, {"a", "5"}}
	if len(rows) != len(want) {
		t.Fatalf("want %v, got %v", want, rows)
	}
	for i, w := range want {
		if rows[i][0] != w[0] || rows[i][1] != w[1] {
			t.Fatalf("row %d: want %v got %v", i, w, rows[i])
		}
	}
}

// readCSVRows is a thin helper that reads the result file via ReadCSV.
func readCSVRows(t *testing.T, path string) [][]string {
	t.Helper()
	var rows [][]string
	_, err := ReadCSV(nil, path, validSchema(t), func(r Result) error {
		rows = append(rows, r.ToRecord())
		return nil
	})
	if err != nil {
		t.Fatalf("readCSVRows: %v", err)
	}
	return rows
}
