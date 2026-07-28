package iplist

import (
	"context"
	"math"
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func mustPrefix(t *testing.T, s string) netip.Prefix {
	t.Helper()

	if strings.Contains(s, "/") {
		p, err := netip.ParsePrefix(s)
		if err != nil {
			t.Fatalf("parse prefix %q: %v", s, err)
		}
		return p
	}

	addr, err := netip.ParseAddr(s)
	if err != nil {
		t.Fatalf("parse addr %q: %v", s, err)
	}
	return netip.PrefixFrom(addr, addr.BitLen())
}

func writeTestCSV(t *testing.T, path string, rows []string) {
	t.Helper()

	data := strings.Join(rows, "\n")
	if data != "" {
		data += "\n"
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

func TestImportIPList_MapPath_DedupesAndMasks(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src.csv")
	dst := filepath.Join(tmpDir, "dst.csv")

	writeTestCSV(t, src, []string{
		"10.0.0.1/24,1",
		"10.0.0.0/24,1",
		"1.1.1.1,1",
		"1.1.1.1/32,1",
		"2.2.2.2,0",
		"invalid-ip,1",
	})

	err := ImportIPList(context.Background(), src, dst, ImportOption{
		MaxMapFileSize:     1 << 20, // force map path for this small file
		MaxInMemoryEntries: 100,
	})
	if err != nil {
		t.Fatalf("ImportIPList() error = %v", err)
	}

	got := readAllCSV(t, dst)
	want := []IPList{
		{IP: mustPrefix(t, "10.0.0.0/24"), Enable: true},
		{IP: mustPrefix(t, "1.1.1.1/32"), Enable: true},
		{IP: mustPrefix(t, "2.2.2.2/32"), Enable: false},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ImportIPList map mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestImportIPList_DiskPath_DedupesAndSorts(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src.csv")
	dst := filepath.Join(tmpDir, "dst.csv")

	writeTestCSV(t, src, []string{
		"3.3.3.3,1",
		"1.1.1.1,1",
		"10.0.0.1/24,1",
		"10.0.0.0/24,1",
		"1.1.1.1/32,1",
		"2.2.2.2,0",
	})

	err := ImportIPList(context.Background(), src, dst, ImportOption{
		MaxMapFileSize: 0, // force disk path
	})
	if err != nil {
		t.Fatalf("ImportIPList() error = %v", err)
	}

	got := readAllCSV(t, dst)
	want := []IPList{
		{IP: mustPrefix(t, "1.1.1.1/32"), Enable: true},
		{IP: mustPrefix(t, "10.0.0.0/24"), Enable: true},
		{IP: mustPrefix(t, "2.2.2.2/32"), Enable: false},
		{IP: mustPrefix(t, "3.3.3.3/32"), Enable: true},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ImportIPList disk mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestImportIPList_MapPath_RespectsMaxInMemoryEntries(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src.csv")
	dst := filepath.Join(tmpDir, "dst.csv")

	writeTestCSV(t, src, []string{
		"1.1.1.1,1",
		"2.2.2.2,1",
	})

	err := ImportIPList(context.Background(), src, dst, ImportOption{
		MaxMapFileSize:     1 << 20,
		MaxInMemoryEntries: 1,
	})
	if err == nil {
		t.Fatal("ImportIPList() error = nil, want memory limit error")
	}

	if !strings.Contains(err.Error(), "in-memory entry limit") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadAll(t *testing.T) {
	t.Parallel()

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
		{IP: mustPrefix(t, "1.1.1.1/32"), Enable: true},
		{IP: mustPrefix(t, "2.2.2.2/32"), Enable: false},
		{IP: mustPrefix(t, "10.0.0.0/31"), Enable: true},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadAll mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestCountIPs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "count.csv")

	writeTestCSV(t, path, []string{
		"1.1.1.1,1",     // 1
		"10.0.0.0/30,1", // 4
		"2.2.2.2,0",     // 1
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
	t.Parallel()

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

func TestCountIPEntry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   netip.Prefix
		want uint64
	}{
		{
			name: "single IPv4",
			in:   mustPrefix(t, "1.1.1.1"),
			want: 1,
		},
		{
			name: "IPv4 /32",
			in:   mustPrefix(t, "10.0.0.1/32"),
			want: 1,
		},
		{
			name: "IPv4 /31",
			in:   mustPrefix(t, "10.0.0.0/31"),
			want: 2,
		},
		{
			name: "IPv4 /30",
			in:   mustPrefix(t, "10.0.0.0/30"),
			want: 4,
		},
		{
			name: "IPv4 /24",
			in:   mustPrefix(t, "10.0.0.0/24"),
			want: 256,
		},
		{
			name: "IPv6 single",
			in:   mustPrefix(t, "2001:db8::1"),
			want: 1,
		},
		{
			name: "IPv6 /128",
			in:   mustPrefix(t, "2001:db8::1/128"),
			want: 1,
		},
		{
			name: "IPv6 /64 saturates to MaxUint64",
			in:   mustPrefix(t, "2001:db8::/64"),
			want: math.MaxUint64,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := countIPEntry(tt.in)
			if got != tt.want {
				t.Fatalf("countIPEntry(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
