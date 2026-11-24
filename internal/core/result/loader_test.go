package result

import (
	"testing"
)

// TestLoadResult_SendsValidRecords covers load result sends valid records.
func TestLoadResult_SendsValidRecords(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\n5.6.7.8,0.5\n")
	schema := validSchema(t)
	out := make(chan Result, 10)

	res, err := LoadResult(nil, path, schema, out)
	if err != nil {
		t.Fatalf("LoadResult(nil, ) error = %v", err)
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

// TestLoadResult_SkipsInvalidRecords covers load result skips invalid records.
func TestLoadResult_SkipsInvalidRecords(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\nbad\n5.6.7.8,0.5\n")
	schema := validSchema(t)
	out := make(chan Result, 10)

	res, err := LoadResult(nil, path, schema, out)
	if err != nil {
		t.Fatalf("LoadResult(nil, ) error = %v", err)
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

// TestLoadResult_NonexistentFile covers load result nonexistent file.
func TestLoadResult_NonexistentFile(t *testing.T) {
	schema := validSchema(t)
	out := make(chan Result, 1)

	_, err := LoadResult(nil, "/nonexistent.csv", schema, out)
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

// TestCountResultKeys_MatchesCount covers count result keys matches count.
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

// TestCountResultKeys_AliasForCount covers count result keys alias for count.
func TestCountResultKeys_AliasForCount(t *testing.T) {
	path := writeTempCSV(t, "10.0.0.1,0.8\n")
	schema := validSchema(t)

	c1, _ := Count(path, schema)
	c2, _ := CountResultKeys(path, schema)
	if c1 != c2 {
		t.Errorf("Count()=%d != CountResultKeys()=%d", c1, c2)
	}
}

// TestLoadAll_WithinLimit covers load all within limit.
func TestLoadAll_WithinLimit(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\n5.6.7.8,0.5\n10.0.0.1,0.3\n")
	schema := validSchema(t)

	results, err := LoadAll(nil, path, schema, 10)
	if err != nil {
		t.Fatalf("LoadAll(nil, ) error = %v", err)
	}
	if len(results) != 3 {
		t.Errorf("got %d results, want 3", len(results))
	}
}

// TestLoadAll_ExceedsLimit covers load all exceeds limit.
func TestLoadAll_ExceedsLimit(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\n5.6.7.8,0.5\n10.0.0.1,0.3\n")
	schema := validSchema(t)

	results, err := LoadAll(nil, path, schema, 2)
	if err != nil {
		t.Fatalf("LoadAll(nil, ) error = %v", err)
	}
	if len(results) != 2 {
		t.Errorf("got %d results, want 2 (maxResults=2)", len(results))
	}
}

// TestLoadAll_LimitZero covers load all limit zero.
func TestLoadAll_LimitZero(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\n")
	schema := validSchema(t)

	results, err := LoadAll(nil, path, schema, 0)
	if err != nil {
		t.Fatalf("LoadAll(nil, ) error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0 (maxResults=0)", len(results))
	}
}

// TestLoadAll_LimitOne covers load all limit one.
func TestLoadAll_LimitOne(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\n5.6.7.8,0.5\n")
	schema := validSchema(t)

	results, err := LoadAll(nil, path, schema, 1)
	if err != nil {
		t.Fatalf("LoadAll(nil, ) error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
	if results[0].Key() != "1.2.3.4" {
		t.Errorf("first result key = %q, want %q", results[0].Key(), "1.2.3.4")
	}
}

// TestLoadAll_EmptyFile covers load all empty file.
func TestLoadAll_EmptyFile(t *testing.T) {
	path := writeTempCSV(t, "")
	schema := validSchema(t)

	results, err := LoadAll(nil, path, schema, 100)
	if err != nil {
		t.Fatalf("LoadAll(nil, ) error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("got %d results from empty file, want 0", len(results))
	}
}

// TestLoadAll_NonexistentFile covers load all nonexistent file.
func TestLoadAll_NonexistentFile(t *testing.T) {
	schema := validSchema(t)

	_, err := LoadAll(nil, "/nonexistent.csv", schema, 10)
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

// TestLoadAll_SkipsInvalidRecords covers load all skips invalid records.
func TestLoadAll_SkipsInvalidRecords(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\ninvalid\n5.6.7.8,0.5\n")
	schema := validSchema(t)

	results, err := LoadAll(nil, path, schema, 10)
	if err != nil {
		t.Fatalf("LoadAll(nil, ) error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("got %d results, want 2 (invalid skipped)", len(results))
	}
}

// TestLoadAll_EOFIsNotError covers load all eofis not error.
func TestLoadAll_EOFIsNotError(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\n5.6.7.8,0.5\n")
	schema := validSchema(t)

	results, err := LoadAll(nil, path, schema, 1)
	if err != nil {
		t.Fatalf("LoadAll(nil, ) should not return error when hitting limit, got: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
}
