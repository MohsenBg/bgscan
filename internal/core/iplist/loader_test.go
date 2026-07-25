package iplist

import (
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeTestCSV(t *testing.T, path string, rows []string) {
	t.Helper()
	data := ""
	for _, row := range rows {
		data += row + "\n"
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("write test csv: %v", err)
	}
}

func readAllCSV(t *testing.T, path string) []IPList {
	t.Helper()
	var out []IPList
	err := ReadCSV(path, func(entry IPList, _ int64) error {
		out = append(out, entry)
		return nil
	})
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	return out
}

func TestCopyIPFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src.csv")
	dst := filepath.Join(tmpDir, "dst.csv")

	writeTestCSV(t, src, []string{
		"1.1.1.1,1",
		"invalid-ip,1",
		"2.2.2.2,0",
		"10.0.0.0/30,1",
	})

	if err := CopyIPFile(src, dst); err != nil {
		t.Fatalf("CopyIPFile() error = %v", err)
	}

	got := readAllCSV(t, dst)
	want := []IPList{
		{IP: "1.1.1.1", Enable: true},
		{IP: "2.2.2.2", Enable: false},
		{IP: "10.0.0.0/30", Enable: true},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("copied content mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestLoadAll(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "input.csv")

	writeTestCSV(t, path, []string{
		"1.1.1.1,1",
		"2.2.2.2,0",
		"bad-value,1", // skipped
		"10.0.0.0/31,1",
	})

	got, err := LoadAll(path)
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	want := []IPList{
		{IP: "1.1.1.1", Enable: true},
		{IP: "2.2.2.2", Enable: false},
		{IP: "10.0.0.0/31", Enable: true},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadAll mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestCountIPs(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "count.csv")

	writeTestCSV(t, path, []string{
		"1.1.1.1,1",     // 1
		"10.0.0.0/30,1", // 4
		"2.2.2.2,0",     // 1 (CountIPs counts all valid entries regardless of enable)
		"bad-ip,1",      // skipped
	})

	got, err := CountIPs(path)
	if err != nil {
		t.Fatalf("CountIPs() error = %v", err)
	}

	var want uint64 = 6
	if got != want {
		t.Fatalf("CountIPs() = %d, want %d", got, want)
	}
}

func TestCountActiveIPs(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "active.csv")

	writeTestCSV(t, path, []string{
		"1.1.1.1,1",     // 1
		"2.2.2.2,0",     // 0
		"10.0.0.0/30,1", // 4
		"10.0.0.8/31,0", // 0
		"bad-ip,1",      // skipped
	})

	got, err := CountActiveIPs(path)
	if err != nil {
		t.Fatalf("CountActiveIPs() error = %v", err)
	}

	var want uint64 = 5
	if got != want {
		t.Fatalf("CountActiveIPs() = %d, want %d", got, want)
	}
}

func TestMergeFiles(t *testing.T) {
	tmpDir := t.TempDir()
	src1 := filepath.Join(tmpDir, "src1.csv")
	src2 := filepath.Join(tmpDir, "src2.csv")
	dst := filepath.Join(tmpDir, "merged.csv")

	writeTestCSV(t, src1, []string{
		"1.1.1.1,1",
		"2.2.2.2,0",
		"10.0.0.0/30,1",
	})

	writeTestCSV(t, src2, []string{
		"2.2.2.2,1",
		"3.3.3.3,1",
		"10.0.0.0/30,0",
	})

	if err := MergeFiles(dst, src1, src2); err != nil {
		t.Fatalf("MergeFiles() error = %v", err)
	}

	got := readAllCSV(t, dst)
	want := []IPList{
		{IP: "1.1.1.1", Enable: true},
		{IP: "2.2.2.2", Enable: false}, // first occurrence kept
		{IP: "10.0.0.0/30", Enable: true},
		{IP: "3.3.3.3", Enable: true},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MergeFiles mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestCountIPEntry(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want uint64
	}{
		{
			name: "single IPv4",
			in:   "1.1.1.1",
			want: 1,
		},
		{
			name: "IPv4 /32",
			in:   "10.0.0.1/32",
			want: 1,
		},
		{
			name: "IPv4 /31",
			in:   "10.0.0.0/31",
			want: 2,
		},
		{
			name: "IPv4 /30",
			in:   "10.0.0.0/30",
			want: 4,
		},
		{
			name: "IPv4 /24",
			in:   "10.0.0.0/24",
			want: 256,
		},
		{
			name: "IPv6 single",
			in:   "2001:db8::1",
			want: 1,
		},
		{
			name: "IPv6 /128",
			in:   "2001:db8::1/128",
			want: 1,
		},
		{
			name: "IPv6 /64 saturates to MaxUint64",
			in:   "2001:db8::/64",
			want: math.MaxUint64,
		},
		{
			name: "invalid IP",
			in:   "not-an-ip",
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countIPEntry(tt.in)
			if got != tt.want {
				t.Fatalf("countIPEntry(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
