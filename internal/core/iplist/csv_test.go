package iplist

import (
	"path/filepath"
	"testing"
)

func TestCSV_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.csv")

	input := []IPList{
		{IP: "1.1.1.1", Enable: true},
		{IP: "2.2.2.2", Enable: false},
		{IP: "10.0.0.0/24", Enable: true},
	}

	// Test Writing
	err := WriteCSV(path, func(write func(IPList) error) error {
		for _, ip := range input {
			if err := write(ip); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to write CSV: %v", err)
	}

	// Test Reading
	var output []IPList
	err = ReadCSV(path, func(entry IPList, offset int64) error {
		output = append(output, entry)
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	if len(output) != len(input) {
		t.Errorf("Expected %d records, got %d", len(input), len(output))
	}

	for i := range input {
		if output[i].IP != input[i].IP || output[i].Enable != input[i].Enable {
			t.Errorf("Mismatch at index %d: got %+v, want %+v", i, output[i], input[i])
		}
	}
}
