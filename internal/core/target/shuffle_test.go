package target

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func mustParseAddr(t *testing.T, s string) netip.Addr {
	t.Helper()

	ip, err := netip.ParseAddr(s)
	if err != nil {
		t.Fatalf("ParseAddr(%q): %v", s, err)
	}
	return ip
}

func mustWriteShuffleFile(t *testing.T, name, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
	return path
}

func collectFromChan(ch <-chan Entry) []Entry {
	var out []Entry
	for e := range ch {
		out = append(out, e)
	}
	return out
}

func entryAddrs(entries []Entry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.IP.Addr().String())
	}
	return out
}

// TestSaturatingAdd: saturatingAdd clamps at MaxUint64 instead of wrapping.
func TestSaturatingAdd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    uint64
		b    uint64
		want uint64
	}{
		{"normal", 10, 20, 30},
		{"zero", 0, 0, 0},
		{"overflow", ^uint64(0), 1, ^uint64(0)},
		{"near overflow", ^uint64(0) - 5, 10, ^uint64(0)},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := saturatingAdd(tt.a, tt.b)
			if got != tt.want {
				t.Fatalf("saturatingAdd(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestAddOffsetToAddr_IPv4: an offset lands on the expected IPv4 address.
func TestAddOffsetToAddr_IPv4(t *testing.T) {
	t.Parallel()

	ip := mustParseAddr(t, "192.168.1.1")
	got := addOffsetToAddr(ip, 5)
	want := mustParseAddr(t, "192.168.1.6")

	if got != want {
		t.Fatalf("addOffsetToAddr() = %v, want %v", got, want)
	}
}

// TestAddOffsetToAddr_CarryAcrossOctets: an offset carries across IPv4 octets.
func TestAddOffsetToAddr_CarryAcrossOctets(t *testing.T) {
	t.Parallel()

	ip := mustParseAddr(t, "10.0.0.255")
	got := addOffsetToAddr(ip, 1)
	want := mustParseAddr(t, "10.0.1.0")

	if got != want {
		t.Fatalf("addOffsetToAddr() = %v, want %v", got, want)
	}
}

// TestAddOffsetToAddr_IPv6: an offset lands on the expected IPv6 address.
func TestAddOffsetToAddr_IPv6(t *testing.T) {
	t.Parallel()

	ip := mustParseAddr(t, "2001:db8::1")
	got := addOffsetToAddr(ip, 2)
	want := mustParseAddr(t, "2001:db8::3")

	if got != want {
		t.Fatalf("addOffsetToAddr() = %v, want %v", got, want)
	}
}

// TestAddBigOffset: a big.Int offset lands on the expected address.
func TestAddBigOffset(t *testing.T) {
	t.Parallel()

	ip := mustParseAddr(t, "2001:db8::")
	offset := new(big.Int).Lsh(big.NewInt(1), 64)

	got := addBigOffset(ip, offset)
	want := mustParseAddr(t, "2001:db8:0:1::")

	if got != want {
		t.Fatalf("addBigOffset() = %v, want %v", got, want)
	}
}

// TestCIDRBlockAt: cidrBlockAt finds the block and address for a global index.
func TestCIDRBlockAt(t *testing.T) {
	t.Parallel()

	mi := &MasterIndexer{
		CIDRBlocks: []CIDRBlock{
			{
				StartIP:   mustParseAddr(t, "10.0.0.0"),
				TotalIPs:  4,
				GlobalIdx: 0,
			},
			{
				StartIP:   mustParseAddr(t, "10.0.1.0"),
				TotalIPs:  4,
				GlobalIdx: 4,
			},
		},
	}

	tests := []struct {
		idx  uint64
		want string
	}{
		{0, "10.0.0.0"},
		{1, "10.0.0.1"},
		{3, "10.0.0.3"},
		{4, "10.0.1.0"},
		{7, "10.0.1.3"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()

			_, got := mi.cidrBlockAt(tt.idx)
			if got.String() != tt.want {
				t.Fatalf("cidrBlockAt(%d) = %v, want %v", tt.idx, got, tt.want)
			}
		})
	}
}

// TestReadEntryAtCSVOffset: a row is read back from its recorded byte offset.
func TestReadEntryAtCSVOffset(t *testing.T) {
	t.Parallel()

	content := "" +
		"1.1.1.1,0,,1\n" +
		"2.2.2.2,8080,/probe,0\n" +
		"3.3.3.3,0,,1\n"

	path := mustWriteShuffleFile(t, "ips.csv", content)

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	defer func() {
		_ = f.Close()
	}()

	offset := int64(len("1.1.1.1,0,,1\n"))
	reader := bufio.NewReader(f)
	got, err := readEntryAtCSVOffset(f, reader, offset)
	if err != nil {
		t.Fatalf("readEntryAtCSVOffset: %v", err)
	}

	want := Entry{
		Kind:   KindIP,
		IP:     mustPrefix(t, "2.2.2.2/32"),
		Port:   8080,
		Path:   "/probe",
		Enable: false,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("readEntryAtCSVOffset() = %v, want %v", got, want)
	}
}

// TestReadEntryAtCSVOffset_InvalidIP: a row with a bad address at the offset returns an error.
func TestReadEntryAtCSVOffset_InvalidIP(t *testing.T) {
	t.Parallel()

	content := "not-an-ip,0,,1\n"
	path := mustWriteShuffleFile(t, "ips.csv", content)

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	defer func() {
		_ = f.Close()
	}()

	reader := bufio.NewReader(f)
	_, err = readEntryAtCSVOffset(f, reader, 0)
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
}

// TestNewMasterIndexer_EmptyFile: an empty file builds an index with a zero total.
func TestNewMasterIndexer_EmptyFile(t *testing.T) {
	t.Parallel()

	path := mustWriteShuffleFile(t, "empty.csv", "")

	idx, err := NewMasterIndexer(nil, path)
	if err != nil {
		t.Fatalf("NewMasterIndexer: %v", err)
	}

	if idx.GrandTotal != 0 {
		t.Fatalf("GrandTotal = %d, want 0", idx.GrandTotal)
	}
	if idx.TotalSingles != 0 {
		t.Fatalf("TotalSingles = %d, want 0", idx.TotalSingles)
	}
	if idx.TotalCIDRIPs != 0 {
		t.Fatalf("TotalCIDRIPs = %d, want 0", idx.TotalCIDRIPs)
	}
}

// TestNewMasterIndexer_SinglesOnly: a file of single addresses indexes each one as a point.
func TestNewMasterIndexer_SinglesOnly(t *testing.T) {
	t.Parallel()

	content := "" +
		"1.1.1.1,0,,1\n" +
		"2.2.2.2,0,,1\n" +
		"3.3.3.3,0,,0\n" +
		"bad-ip,0,,1\n"

	path := mustWriteShuffleFile(t, "single.csv", content)

	idx, err := NewMasterIndexer(nil, path)
	if err != nil {
		t.Fatalf("NewMasterIndexer: %v", err)
	}

	// with correct logic: 2 singles, 0 CIDR IPs
	if idx.TotalSingles != 2 {
		t.Fatalf("TotalSingles = %d, want 2", idx.TotalSingles)
	}
	if idx.TotalCIDRIPs != 0 {
		t.Fatalf("TotalCIDRIPs = %d, want 0", idx.TotalCIDRIPs)
	}
	if idx.GrandTotal != 2 {
		t.Fatalf("GrandTotal = %d, want 2", idx.GrandTotal)
	}
	if len(idx.SingleOffsets) != 2 {
		t.Fatalf("len(SingleOffsets) = %d, want 2", len(idx.SingleOffsets))
	}
}

// TestNewMasterIndexer_CIDRsAndSingles: ranges and single addresses share one index and total.
func TestNewMasterIndexer_CIDRsAndSingles(t *testing.T) {
	t.Parallel()

	content := "" +
		"10.0.0.1,0,,1\n" +
		"192.168.0.0/30,0,,1\n" +
		"192.168.1.0/31,0,,1\n" +
		"10.0.0.2,0,,0\n"

	path := mustWriteShuffleFile(t, "mixed.csv", content)

	idx, err := NewMasterIndexer(nil, path)
	if err != nil {
		t.Fatalf("NewMasterIndexer: %v", err)
	}

	// correct behavior:
	// single 10.0.0.1 = 1
	// /30 = 4
	// /31 = 2
	if idx.TotalSingles != 1 {
		t.Fatalf("TotalSingles = %d, want 1", idx.TotalSingles)
	}
	if idx.TotalCIDRIPs != 6 {
		t.Fatalf("TotalCIDRIPs = %d, want 6", idx.TotalCIDRIPs)
	}
	if idx.GrandTotal != 7 {
		t.Fatalf("GrandTotal = %d, want 7", idx.GrandTotal)
	}
	if len(idx.CIDRBlocks) != 2 {
		t.Fatalf("len(CIDRBlocks) = %d, want 2", len(idx.CIDRBlocks))
	}
}

// TestPermuter_IsBijection: the permuter maps each input to a distinct output.
func TestPermuter_IsBijection(t *testing.T) {
	t.Parallel()

	for _, nbits := range []uint{1, 2, 5, 8, 16} {
		nbits := nbits
		t.Run(fmt.Sprintf("%d bits", nbits), func(t *testing.T) {
			t.Parallel()

			p := newPermuter(nbits, rand.New(rand.NewSource(1)))
			limit := uint64(1) << nbits
			seen := make([]bool, limit)

			for i := uint64(0); i < limit; i++ {
				v := p.at(i)
				if v >= limit {
					t.Fatalf("at(%d) = %d, outside [0,%d)", i, v, limit)
				}
				if seen[v] {
					t.Fatalf("at(%d) = %d, value repeated", i, v)
				}
				seen[v] = true
			}
		})
	}
}

// TestPermuter_MixesOrder: the permuter does not return its inputs in order.
func TestPermuter_MixesOrder(t *testing.T) {
	t.Parallel()

	p := newPermuter(10, rand.New(rand.NewSource(1)))

	var fixed uint64
	for i := uint64(0); i < 1<<10; i++ {
		if p.at(i) == i {
			fixed++
		}
	}
	if fixed == 1<<10 {
		t.Fatal("permutation maps every value to itself")
	}
}

// TestPermuter_SingleTarget: a one-target range permutes to itself.
func TestPermuter_SingleTarget(t *testing.T) {
	t.Parallel()

	// GrandTotal == 1 leaves no room for a mask; this is why newPermuter
	// tolerates nbits 0.
	p := newPermuter(0, rand.New(rand.NewSource(1)))
	if got := p.at(0); got != 0 {
		t.Fatalf("at(0) = %d, want 0", got)
	}
}

// TestStreamActiveShuffled_Empty: shuffled streaming of an empty list sends nothing.
func TestStreamActiveShuffled_Empty(t *testing.T) {
	t.Parallel()

	path := mustWriteShuffleFile(t, "empty.csv", "")

	out := make(chan Entry, 10)
	errCh := make(chan error, 1)

	go func() {
		errCh <- streamActiveShuffled(nil, context.Background(), path, 0, out)
		close(out)
	}()

	got := collectFromChan(out)
	err := <-errCh
	if err != nil {
		t.Fatalf("streamActiveShuffled: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("len(got) = %d, want 0", len(got))
	}
}

// TestStreamActiveShuffled_SingleIPs_NoDuplicates: shuffled streaming sends each single address once.
func TestStreamActiveShuffled_SingleIPs_NoDuplicates(t *testing.T) {
	t.Parallel()

	content := "" +
		"1.1.1.1,0,,1\n" +
		"2.2.2.2,0,,1\n" +
		"3.3.3.3,0,,1\n"

	path := mustWriteShuffleFile(t, "single.csv", content)

	out := make(chan Entry, 10)
	errCh := make(chan error, 1)

	go func() {
		errCh <- streamActiveShuffled(nil, context.Background(), path, 0, out)
		close(out)
	}()

	got := collectFromChan(out)
	err := <-errCh
	if err != nil {
		t.Fatalf("streamActiveShuffled: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("len(got) = %d, want 3", len(got))
	}

	gotStr := entryAddrs(got)
	seen := make(map[string]struct{}, len(gotStr))
	for _, ip := range gotStr {
		seen[ip] = struct{}{}
	}
	if len(seen) != 3 {
		t.Fatalf("got duplicates: %v", gotStr)
	}

	wantSet := []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"}
	slices.Sort(gotStr)
	slices.Sort(wantSet)
	if !slices.Equal(gotStr, wantSet) {
		t.Fatalf("got %v, want %v", gotStr, wantSet)
	}
}

// TestStreamActiveShuffled_MixedDataset: shuffled streaming covers a mix of ranges and singles once each.
func TestStreamActiveShuffled_MixedDataset(t *testing.T) {
	t.Parallel()

	content := "" +
		"10.0.0.1,0,,1\n" +
		"192.168.1.0/30,0,,1\n" +
		"10.0.0.2,0,,0\n" +
		"bad-row,0,,1\n"

	path := mustWriteShuffleFile(t, "mixed.csv", content)

	out := make(chan Entry, 10)
	errCh := make(chan error, 1)

	go func() {
		errCh <- streamActiveShuffled(nil, context.Background(), path, 0, out)
		close(out)
	}()

	got := collectFromChan(out)
	err := <-errCh
	if err != nil {
		t.Fatalf("streamActiveShuffled: %v", err)
	}

	gotStr := entryAddrs(got)
	if len(gotStr) != 5 {
		t.Fatalf("len(got) = %d, want 5; got=%v", len(gotStr), gotStr)
	}

	wantSet := []string{
		"10.0.0.1",
		"192.168.1.0",
		"192.168.1.1",
		"192.168.1.2",
		"192.168.1.3",
	}
	slices.Sort(gotStr)
	slices.Sort(wantSet)
	if !slices.Equal(gotStr, wantSet) {
		t.Fatalf("got %v, want %v", gotStr, wantSet)
	}
}

// TestStreamActiveShuffled_DomainAndCIDRMixed: shuffled streaming covers domains and range addresses once each.
func TestStreamActiveShuffled_DomainAndCIDRMixed(t *testing.T) {
	t.Parallel()

	content := "" +
		"example.com,443,/dns-query,1\n" +
		"10.0.0.0/30,0,,1\n" +
		"10.0.0.9,0,,1\n"

	path := mustWriteShuffleFile(t, "domain-mixed.csv", content)

	out := make(chan Entry, 16)
	errCh := make(chan error, 1)

	go func() {
		errCh <- streamActiveShuffled(nil, context.Background(), path, 0, out)
		close(out)
	}()

	got := collectFromChan(out)
	if err := <-errCh; err != nil {
		t.Fatalf("streamActiveShuffled: %v", err)
	}

	// One domain + four addresses from the /30 + one single.
	if len(got) != 6 {
		t.Fatalf("len(got) = %d, want 6; got=%v", len(got), got)
	}

	var domains int
	seen := make(map[string]struct{}, len(got))
	for _, e := range got {
		if _, dup := seen[e.Key()]; dup {
			t.Fatalf("duplicate target %s", e.Key())
		}
		seen[e.Key()] = struct{}{}

		if e.Kind != KindDomain {
			continue
		}
		domains++
		if e.Domain != "example.com" || e.Port != 443 || e.Path != "/dns-query" {
			t.Fatalf("domain row lost its fields: %+v", e)
		}
	}
	if domains != 1 {
		t.Fatalf("streamed %d domains, want 1", domains)
	}
}

// TestStreamActiveShuffled_KeepsPortAndPath: shuffled streaming keeps the port and path of each row.
func TestStreamActiveShuffled_KeepsPortAndPath(t *testing.T) {
	t.Parallel()

	content := "" +
		"192.168.1.0/30,8443,/probe,1\n" +
		"10.0.0.1,1080,/x,1\n"

	path := mustWriteShuffleFile(t, "portpath.csv", content)

	out := make(chan Entry, 10)
	errCh := make(chan error, 1)

	go func() {
		errCh <- streamActiveShuffled(nil, context.Background(), path, 0, out)
		close(out)
	}()

	got := collectFromChan(out)
	if err := <-errCh; err != nil {
		t.Fatalf("streamActiveShuffled: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("len(got) = %d, want 5", len(got))
	}

	var single Entry
	for _, e := range got {
		addr := e.IP.Addr().String()
		if strings.HasPrefix(addr, "192.168.1.") {
			if e.Port != 8443 || e.Path != "/probe" {
				t.Fatalf("%s lost its port/path: %+v", addr, e)
			}
			continue
		}
		single = e
	}

	if single.Port != 1080 || single.Path != "/x" {
		t.Fatalf("single row lost its port/path: %+v", single)
	}
}

// TestStreamActiveShuffled_Limit: shuffled streaming stops at the limit.
func TestStreamActiveShuffled_Limit(t *testing.T) {
	t.Parallel()

	content := "" +
		"1.1.1.1,0,,1\n" +
		"2.2.2.2,0,,1\n" +
		"3.3.3.3,0,,1\n" +
		"4.4.4.4,0,,1\n"

	path := mustWriteShuffleFile(t, "limit.csv", content)

	out := make(chan Entry, 10)
	errCh := make(chan error, 1)

	go func() {
		errCh <- streamActiveShuffled(nil, context.Background(), path, 2, out)
		close(out)
	}()

	got := collectFromChan(out)
	err := <-errCh
	if err != nil {
		t.Fatalf("streamActiveShuffled: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
}

// TestStreamActiveShuffled_ContextCanceled: shuffled streaming stops when the context is cancelled.
func TestStreamActiveShuffled_ContextCanceled(t *testing.T) {
	t.Parallel()

	content := "10.0.0.0/24,0,,1\n"
	path := mustWriteShuffleFile(t, "cancel.csv", content)

	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan Entry)
	errCh := make(chan error, 1)

	go func() {
		errCh <- streamActiveShuffled(nil, ctx, path, 0, out)
		close(out)
	}()

	select {
	case <-out:
		cancel()
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first output")
	}

	err := <-errCh
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
