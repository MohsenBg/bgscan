package target

import (
	"net/netip"
	"path/filepath"
	"reflect"
	"testing"
)

// TestCSV_RoundTrip: entries written with WriteCSV read back unchanged with ReadCSV.
func TestCSV_RoundTrip(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.csv")

	input := []Entry{
		{IP: netip.MustParsePrefix("1.1.1.1/32"), Enable: true},
		{IP: netip.MustParsePrefix("2.2.2.2/32"), Enable: false},
		{IP: netip.MustParsePrefix("10.0.0.0/24"), Enable: true},
		{IP: netip.MustParsePrefix("8.8.8.8/32"), Port: 8080, Enable: true},
		{IP: netip.MustParsePrefix("1.2.3.4/32"), Port: 443, Path: "/login", Enable: false},
		{Kind: KindDomain, Domain: "example.com", Port: 443, Path: "/dns-query", Enable: true},
	}

	err := WriteCSV(path, func(write func(Entry) error) error {
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
		output  []Entry
		offsets []int64
	)

	err = ReadCSV(path, func(entry Entry, offset int64) error {
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

	if !reflect.DeepEqual(output, input) {
		t.Fatalf("round-trip mismatch\ngot:  %#v\nwant: %#v", output, input)
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

// TestCSV_Read_NormalizesMaskedPrefix: reading a row masks its prefix before it reaches the callback.
func TestCSV_Read_NormalizesMaskedPrefix(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "canonical.csv")

	input := Entry{
		IP:     netip.MustParsePrefix("10.0.0.123/24"),
		Enable: true,
	}

	err := WriteCSV(path, func(write func(Entry) error) error {
		return write(input)
	})
	if err != nil {
		t.Fatalf("WriteCSV() error = %v", err)
	}

	var got []Entry
	err = ReadCSV(path, func(entry Entry, offset int64) error {
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
