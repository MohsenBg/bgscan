package result

import (
	"testing"
)

func TestCount_ValidRecords(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\n5.6.7.8,0.5\n")
	schema := validSchema(t)

	count, err := Count(path, schema)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 2 {
		t.Errorf("Count() = %d, want 2", count)
	}
}

func TestCount_EmptyFile(t *testing.T) {
	path := writeTempCSV(t, "")
	schema := validSchema(t)

	count, err := Count(path, schema)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Count() = %d, want 0", count)
	}
}

func TestCount_SkipsInvalidRecords(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9\nbadfield\n5.6.7.8,0.5\n")
	schema := validSchema(t)

	count, err := Count(path, schema)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 2 {
		t.Errorf("Count() = %d, want 2 (invalid records skipped)", count)
	}
}

func TestCount_NonexistentFile(t *testing.T) {
	schema := validSchema(t)

	_, err := Count("/nonexistent/path.csv", schema)
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestCount_LargeFile(t *testing.T) {
	var lines string
	const n = 500
	for i := 0; i < n; i++ {
		lines += "10.0.0.1,0.5\n"
	}
	path := writeTempCSV(t, lines)
	schema := validSchema(t)

	count, err := Count(path, schema)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != uint64(n) {
		t.Errorf("Count() = %d, want %d", count, n)
	}
}

func TestCount_AllInvalidRecords(t *testing.T) {
	path := writeTempCSV(t, "bad\nalso_bad\n")
	schema := validSchema(t)

	count, err := Count(path, schema)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Count() = %d, want 0", count)
	}
}

func TestCount_SingleRecord(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,1.0\n")
	schema := validSchema(t)

	count, err := Count(path, schema)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 1 {
		t.Errorf("Count() = %d, want 1", count)
	}
}

func TestCount_ExtraFieldsIgnored(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0.9,extra,fields\n")
	schema := validSchema(t)

	count, err := Count(path, schema)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 1 {
		t.Errorf("Count() = %d, want 1", count)
	}
}

func TestCount_ScoreZero(t *testing.T) {
	path := writeTempCSV(t, "1.2.3.4,0\n")
	schema := validSchema(t)

	count, err := Count(path, schema)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 1 {
		t.Errorf("Count() = %d, want 1", count)
	}
}
