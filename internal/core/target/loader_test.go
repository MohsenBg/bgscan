package target

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

func readAllCSV(t *testing.T, path string) []Entry {
	t.Helper()

	var out []Entry
	err := ReadCSV(path, func(entry Entry, _ int64) error {
		out = append(out, entry)
		return nil
	})
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	return out
}

func mustTargetPath(t *testing.T, name string, kind Kind) string {
	t.Helper()

	path, err := ResolveTargetFilePath(name, kind)
	if err != nil {
		t.Fatalf("ResolveTargetFilePath(%q, %v): %v", name, kind, err)
	}
	return path
}

func TestImport_MapPath_DedupesAndMasks(t *testing.T) {
	setBaseDir(t)

	src := filepath.Join(t.TempDir(), "src.csv")

	writeTestCSV(t, src, []string{
		"10.0.0.1/24,0,,1",
		"10.0.0.0/24,0,,1",
		"1.1.1.1,0,,1",
		"1.1.1.1/32,0,,1",
		"2.2.2.2,0,,0",
		"invalid-ip,0,,1",
	})

	result, err := Import(context.Background(), src, "mylist", ImportOption{
		MaxMapFileSize:     1 << 20, // force map path for this small file
		MaxInMemoryEntries: 100,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	if result.IPs != 3 || result.Domains != 0 {
		t.Fatalf("Import() result = %+v, want 3 IPs and 0 domains", result)
	}

	got := readAllCSV(t, mustTargetPath(t, "mylist", KindIP))
	want := []Entry{
		{IP: mustPrefix(t, "10.0.0.0/24"), Enable: true},
		{IP: mustPrefix(t, "1.1.1.1/32"), Enable: true},
		{IP: mustPrefix(t, "2.2.2.2/32"), Enable: false},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Import map mismatch\ngot:  %#v\nwant: %#v", got, want)
	}

	if TargetFileExists("mylist", KindDomain) {
		t.Fatal("IP-only import must not create a domain file")
	}
}

func TestImport_DiskPath_DedupesAndSorts(t *testing.T) {
	setBaseDir(t)

	src := filepath.Join(t.TempDir(), "src.csv")

	writeTestCSV(t, src, []string{
		"3.3.3.3,0,,1",
		"1.1.1.1,0,,1",
		"10.0.0.1/24,0,,1",
		"10.0.0.0/24,0,,1",
		"1.1.1.1/32,0,,1",
		"2.2.2.2,0,,0",
	})

	result, err := Import(context.Background(), src, "mylist", ImportOption{
		MaxMapFileSize: 0, // force disk path
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	if result.IPs != 4 || result.Domains != 0 {
		t.Fatalf("Import() result = %+v, want 4 IPs and 0 domains", result)
	}

	got := readAllCSV(t, mustTargetPath(t, "mylist", KindIP))
	want := []Entry{
		{IP: mustPrefix(t, "1.1.1.1/32"), Enable: true},
		{IP: mustPrefix(t, "10.0.0.0/24"), Enable: true},
		{IP: mustPrefix(t, "2.2.2.2/32"), Enable: false},
		{IP: mustPrefix(t, "3.3.3.3/32"), Enable: true},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Import disk mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestImport_DiskPath_DedupesKindPortAndPath(t *testing.T) {
	setBaseDir(t)

	src := filepath.Join(t.TempDir(), "src.csv")

	writeTestCSV(t, src, []string{
		"example.com,443,/a,1",
		"EXAMPLE.com,443,/a,1",  // same target: case-insensitive domain
		"example.com,443,/b,1",  // different path
		"example.com,8443,/a,1", // different port
		"8.8.8.8,53,,1",
		"8.8.8.8,53,,1",  // exact duplicate
		"8.8.8.8,853,,1", // different port
		"1.1.1.1,0,,1",
	})

	result, err := Import(context.Background(), src, "mixed", ImportOption{
		MaxMapFileSize: 0, // force disk path
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	if result.IPs != 3 || result.Domains != 3 {
		t.Fatalf("Import() result = %+v, want 3 IPs and 3 domains", result)
	}

	gotIPs := readAllCSV(t, mustTargetPath(t, "mixed", KindIP))
	wantIPs := []Entry{
		{IP: mustPrefix(t, "1.1.1.1/32"), Enable: true},
		{IP: mustPrefix(t, "8.8.8.8/32"), Port: 53, Enable: true},
		{IP: mustPrefix(t, "8.8.8.8/32"), Port: 853, Enable: true},
	}
	if !reflect.DeepEqual(gotIPs, wantIPs) {
		t.Fatalf("Import disk IP mismatch\ngot:  %#v\nwant: %#v", gotIPs, wantIPs)
	}

	gotDomains := readAllCSV(t, mustTargetPath(t, "mixed", KindDomain))
	wantDomains := []Entry{
		{Kind: KindDomain, Domain: "example.com", Port: 443, Path: "/a", Enable: true},
		{Kind: KindDomain, Domain: "example.com", Port: 443, Path: "/b", Enable: true},
		{Kind: KindDomain, Domain: "example.com", Port: 8443, Path: "/a", Enable: true},
	}
	if !reflect.DeepEqual(gotDomains, wantDomains) {
		t.Fatalf("Import disk domain mismatch\ngot:  %#v\nwant: %#v", gotDomains, wantDomains)
	}
}

func TestImport_MapPath_RespectsMaxInMemoryEntries(t *testing.T) {
	setBaseDir(t)

	src := filepath.Join(t.TempDir(), "src.csv")

	writeTestCSV(t, src, []string{
		"1.1.1.1,0,,1",
		"2.2.2.2,0,,1",
	})

	_, err := Import(context.Background(), src, "mylist", ImportOption{
		MaxMapFileSize:     1 << 20,
		MaxInMemoryEntries: 1,
	})
	if err == nil {
		t.Fatal("Import() error = nil, want memory limit error")
	}

	if !strings.Contains(err.Error(), "in-memory entry limit") {
		t.Fatalf("unexpected error: %v", err)
	}

	if TargetFileExists("mylist", KindIP) || TargetFileExists("mylist", KindDomain) {
		t.Fatal("failed import must not publish a destination file")
	}
}

func TestImport_MapPath_DedupesKindPortAndPath(t *testing.T) {
	setBaseDir(t)

	src := filepath.Join(t.TempDir(), "src.csv")

	writeTestCSV(t, src, []string{
		"example.com,443,/a,1",
		"EXAMPLE.com,443,/a,1",  // same target: case-insensitive domain
		"example.com,443,/b,1",  // different path
		"example.com,8443,/a,1", // different port
		"8.8.8.8,53,,1",
		"8.8.8.8,53,,1",  // exact duplicate
		"8.8.8.8,853,,1", // different port
	})

	result, err := Import(context.Background(), src, "mixed", ImportOption{
		MaxMapFileSize:     1 << 20,
		MaxInMemoryEntries: 100,
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	if result.IPs != 2 || result.Domains != 3 {
		t.Fatalf("Import() result = %+v, want 2 IPs and 3 domains", result)
	}

	gotIPs := readAllCSV(t, mustTargetPath(t, "mixed", KindIP))
	wantIPs := []Entry{
		{IP: mustPrefix(t, "8.8.8.8/32"), Port: 53, Enable: true},
		{IP: mustPrefix(t, "8.8.8.8/32"), Port: 853, Enable: true},
	}
	if !reflect.DeepEqual(gotIPs, wantIPs) {
		t.Fatalf("Import map IP mismatch\ngot:  %#v\nwant: %#v", gotIPs, wantIPs)
	}

	gotDomains := readAllCSV(t, mustTargetPath(t, "mixed", KindDomain))
	wantDomains := []Entry{
		{Kind: KindDomain, Domain: "example.com", Port: 443, Path: "/a", Enable: true},
		{Kind: KindDomain, Domain: "example.com", Port: 443, Path: "/b", Enable: true},
		{Kind: KindDomain, Domain: "example.com", Port: 8443, Path: "/a", Enable: true},
	}
	if !reflect.DeepEqual(gotDomains, wantDomains) {
		t.Fatalf("Import map domain mismatch\ngot:  %#v\nwant: %#v", gotDomains, wantDomains)
	}
}

func TestImport_RejectsExistingTarget(t *testing.T) {
	for _, existingKind := range []Kind{KindIP, KindDomain} {
		t.Run(existingKind.String(), func(t *testing.T) {
			base := setBaseDir(t)

			existing := filepath.Join(targetDir(base, existingKind), "existing.csv")
			mustWriteFile(t, existing, "1.1.1.1,0,,1\n")

			src := filepath.Join(t.TempDir(), "src.csv")
			writeTestCSV(t, src, []string{"1.1.1.1,0,,1"})

			_, err := Import(context.Background(), src, "existing", DefaultImportOption())
			if err == nil || !strings.Contains(err.Error(), "already exists") {
				t.Fatalf("Import() error = %v, want already exists", err)
			}

			otherKind := KindDomain
			if existingKind == KindDomain {
				otherKind = KindIP
			}
			if TargetFileExists("existing", otherKind) {
				t.Fatalf("Import() must not create the %v file when the %v file exists", otherKind, existingKind)
			}
		})
	}
}

func TestImport_NoValidTargets(t *testing.T) {
	setBaseDir(t)

	src := filepath.Join(t.TempDir(), "src.csv")
	writeTestCSV(t, src, []string{"not-a-target"})

	_, err := Import(context.Background(), src, "mylist", DefaultImportOption())
	if err == nil || !strings.Contains(err.Error(), "no valid targets") {
		t.Fatalf("Import() error = %v, want no valid targets", err)
	}

	if TargetFileExists("mylist", KindIP) || TargetFileExists("mylist", KindDomain) {
		t.Fatal("invalid import must not create destination files")
	}
}

func TestImport_DomainOnlyDoesNotCreateIPFile(t *testing.T) {
	setBaseDir(t)

	src := filepath.Join(t.TempDir(), "src.csv")
	writeTestCSV(t, src, []string{"example.com,443,,1"})

	result, err := Import(context.Background(), src, "domains", DefaultImportOption())
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	if result.IPs != 0 || result.Domains != 1 {
		t.Fatalf("Import() result = %+v, want 0 IPs and 1 domain", result)
	}

	if TargetFileExists("domains", KindIP) {
		t.Fatal("domain-only import must not create an IP file")
	}
	if !TargetFileExists("domains", KindDomain) {
		t.Fatal("domain-only import must create a domain file")
	}
}

func TestCountAll_DomainCountsAsOne(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "count-domain.csv")

	writeTestCSV(t, path, []string{
		"example.com,0,,1",
		"example.com,443,,1",
		"10.0.0.0/30,0,,1",
	})

	got, err := CountAll(path)
	if err != nil {
		t.Fatalf("CountAll() error = %v", err)
	}

	var want uint64 = 6 // domain + domain + 4 addresses
	if got != want {
		t.Fatalf("CountAll() = %d, want %d", got, want)
	}
}

func TestImport_AcceptsFlexibleInput(t *testing.T) {
	setBaseDir(t)

	src := filepath.Join(t.TempDir(), "plain.txt")
	writeTestCSV(t, src, []string{
		"1.1.1.1",
		"2.2.2.2,0",
		"example.com",
	})

	result, err := Import(context.Background(), src, "mixed", DefaultImportOption())
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if result.IPs != 2 || result.Domains != 1 {
		t.Fatalf("Import() result = %+v, want 2 IPs and 1 domain", result)
	}

	gotIPs := readAllCSV(t, mustTargetPath(t, "mixed", KindIP))
	wantIPs := []Entry{
		{IP: mustPrefix(t, "1.1.1.1/32"), Enable: true},
		{IP: mustPrefix(t, "2.2.2.2/32"), Enable: false},
	}
	if !reflect.DeepEqual(gotIPs, wantIPs) {
		t.Fatalf("IP import mismatch\ngot:  %#v\nwant: %#v", gotIPs, wantIPs)
	}

	gotDomains := readAllCSV(t, mustTargetPath(t, "mixed", KindDomain))
	wantDomains := []Entry{
		{Kind: KindDomain, Domain: "example.com", Enable: true},
	}
	if !reflect.DeepEqual(gotDomains, wantDomains) {
		t.Fatalf("domain import mismatch\ngot:  %#v\nwant: %#v", gotDomains, wantDomains)
	}
}

func TestLoadAll(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "input.csv")

	writeTestCSV(t, path, []string{
		"1.1.1.1,0,,1",
		"2.2.2.2,0,,0",
		"bad-value,0,,1", // skipped
		"10.0.0.0/31,0,,1",
	})

	got, err := LoadAll(path)
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	want := []Entry{
		{IP: mustPrefix(t, "1.1.1.1/32"), Enable: true},
		{IP: mustPrefix(t, "2.2.2.2/32"), Enable: false},
		{IP: mustPrefix(t, "10.0.0.0/31"), Enable: true},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadAll mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
}

func TestCountAll(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "count.csv")

	writeTestCSV(t, path, []string{
		"1.1.1.1,0,,1",     // 1
		"10.0.0.0/30,0,,1", // 4
		"2.2.2.2,0,,0",     // 1
		"bad-ip,0,,1",      // skipped
	})

	got, err := CountAll(path)
	if err != nil {
		t.Fatalf("CountAll() error = %v", err)
	}

	var want uint64 = 6
	if got != want {
		t.Fatalf("CountAll() = %d, want %d", got, want)
	}
}

func TestCountActive(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "active.csv")

	writeTestCSV(t, path, []string{
		"1.1.1.1,0,,1",     // 1
		"2.2.2.2,0,,0",     // 0
		"10.0.0.0/30,0,,1", // 4
		"10.0.0.8/31,0,,0", // 0
		"bad-ip,0,,1",      // skipped
	})

	got, err := CountActive(path)
	if err != nil {
		t.Fatalf("CountActive() error = %v", err)
	}

	var want uint64 = 5
	if got != want {
		t.Fatalf("CountActive() = %d, want %d", got, want)
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

			got := countEntry(New(tt.in, 1))
			if got != tt.want {
				t.Fatalf("countEntry(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
