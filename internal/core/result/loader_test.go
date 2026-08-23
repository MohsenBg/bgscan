package result

import (
	"testing"
)

func TestLoadResult_SendsValidRecords(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\n5.6.7.8,0.5\n")
	schema := validSchema(t)
	out := make(chan Result, 10)

	res, err := LoadResult(path, schema, out)
	if err != nil {
		t.Fatalf("LoadResult() error = %v", err)
	}

	close(out)
	var results []Result
	for r := range out {
		results = append(results, r)
	}
	if res.Loaded != 2 {
		t.Errorf("Loaded = %d, want 2", res.Loaded)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Key() != "1.2.3.4" {
		t.Errorf("results[0].Key() = %q", results[0].Key())
	}
	if results[1].Key() != "5.6.7.8" {
		t.Errorf("results[1].Key() = %q", results[1].Key())
	}
}

func TestLoadResult_SkipsInvalidRecords(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\nbad\n5.6.7.8,0.5\n")
	schema := validSchema(t)
	out := make(chan Result, 10)

	res, err := LoadResult(path, schema, out)
	if err != nil {
		t.Fatalf("LoadResult() error = %v", err)
	}

	close(out)
	count := 0
	for range out {
		count++
	}
	if count != 2 {
		t.Errorf("got %d results, want 2 (bad line skipped)", count)
	}
	if res.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", res.Skipped)
	}
}

func TestLoadResult_NonexistentFile(t *testing.T) {
	schema := validSchema(t)
	out := make(chan Result, 1)

	_, err := LoadResult("/nonexistent.csv", schema, out)
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestCountResultKeys_MatchesCount(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\n5.6.7.8,0.5\n")
	schema := validSchema(t)

	c, err := CountResultKeys(path, schema)
	if err != nil {
		t.Fatalf("CountResultKeys() error = %v", err)
	}
	if c != 2 {
		t.Errorf("CountResultKeys() = %d, want 2", c)
	}
}

func TestCountResultKeys_AliasForCount(t *testing.T) {
	path := writeTempCSV(t, "10.0.0.1,0.8\n")
	schema := validSchema(t)

	c1, _ := Count(path, schema)
	c2, _ := CountResultKeys(path, schema)
	if c1 != c2 {
		t.Errorf("Count()=%d != CountResultKeys()=%d", c1, c2)
	}
}

func TestLoadAll_WithinLimit(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\n5.6.7.8,0.5\n10.0.0.1,0.3\n")
	schema := validSchema(t)

	results, err := LoadAll(path, schema, 10)
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}
	if len(results) != 3 {
		t.Errorf("got %d results, want 3", len(results))
	}
}

func TestLoadAll_ExceedsLimit(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\n5.6.7.8,0.5\n10.0.0.1,0.3\n")
	schema := validSchema(t)

	results, err := LoadAll(path, schema, 2)
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}
	if len(results) != 2 {
		t.Errorf("got %d results, want 2 (maxResults=2)", len(results))
	}
}

func TestLoadAll_LimitZero(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\n")
	schema := validSchema(t)

	results, err := LoadAll(path, schema, 0)
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0 (maxResults=0)", len(results))
	}
}

func TestLoadAll_LimitOne(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\n5.6.7.8,0.5\n")
	schema := validSchema(t)

	results, err := LoadAll(path, schema, 1)
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
	if results[0].Key() != "1.2.3.4" {
		t.Errorf("first result key = %q, want %q", results[0].Key(), "1.2.3.4")
	}
}

func TestLoadAll_EmptyFile(t *testing.T) {
	path := writeTempCSV(t, "")
	schema := validSchema(t)

	results, err := LoadAll(path, schema, 100)
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("got %d results from empty file, want 0", len(results))
	}
}

func TestLoadAll_NonexistentFile(t *testing.T) {
	schema := validSchema(t)

	_, err := LoadAll("/nonexistent.csv", schema, 10)
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestLoadAll_SkipsInvalidRecords(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\ninvalid\n5.6.7.8,0.5\n")
	schema := validSchema(t)

	results, err := LoadAll(path, schema, 10)
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("got %d results, want 2 (invalid skipped)", len(results))
	}
}

func TestLoadAll_EOFIsNotError(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\n5.6.7.8,0.5\n")
	schema := validSchema(t)

	results, err := LoadAll(path, schema, 1)
	if err != nil {
		t.Fatalf("LoadAll() should not return error when hitting limit, got: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
}
