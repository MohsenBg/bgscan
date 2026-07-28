package iplist

import (
	"net/netip"
	"path/filepath"
	"testing"
)

func TestCSV_RoundTrip(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.csv")

	input := []IPList{
		{IP: netip.MustParsePrefix("1.1.1.1/32"), Enable: true},
		{IP: netip.MustParsePrefix("2.2.2.2/32"), Enable: false},
		{IP: netip.MustParsePrefix("10.0.0.0/24"), Enable: true},
	}

	err := WriteCSV(path, func(write func(IPList) error) error {
		for _, item := range input {
			if err := write(item); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WriteCSV() error = %v", err)
	}

	var (
		output  []IPList
		offsets []int64
	)

	err = ReadCSV(path, func(entry IPList, offset int64) error {
		output = append(output, entry)
		offsets = append(offsets, offset)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadCSV() error = %v", err)
	}

	if got, want := len(output), len(input); got != want {
		t.Fatalf("record count mismatch: got %d, want %d", got, want)
	}

	for i := range input {
		if output[i].IP != input[i].IP {
			t.Errorf("record %d IP mismatch: got %s, want %s", i, output[i].IP, input[i].IP)
		}
		if output[i].Enable != input[i].Enable {
			t.Errorf("record %d Enable mismatch: got %v, want %v", i, output[i].Enable, input[i].Enable)
		}
	}

	for i := range offsets {
		if offsets[i] < 0 {
			t.Errorf("record %d has negative offset: %d", i, offsets[i])
		}
		if i > 0 && offsets[i] <= offsets[i-1] {
			t.Errorf("offsets must be strictly increasing: offsets[%d]=%d <= offsets[%d]=%d",
				i, offsets[i], i-1, offsets[i-1])
		}
	}
}

func TestCSV_Read_NormalizesMaskedPrefix(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "canonical.csv")

	input := IPList{
		IP:     netip.MustParsePrefix("10.0.0.123/24"),
		Enable: true,
	}

	err := WriteCSV(path, func(write func(IPList) error) error {
		return write(input)
	})
	if err != nil {
		t.Fatalf("WriteCSV() error = %v", err)
	}

	var got []IPList
	err = ReadCSV(path, func(entry IPList, offset int64) error {
		got = append(got, entry)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadCSV() error = %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("got %d records, want 1", len(got))
	}

	wantIP := input.IP.Masked()
	if got[0].IP != wantIP {
		t.Fatalf("normalized prefix mismatch: got %s, want %s", got[0].IP, wantIP)
	}
	if got[0].Enable != input.Enable {
		t.Fatalf("enable mismatch: got %v, want %v", got[0].Enable, input.Enable)
	}
}
