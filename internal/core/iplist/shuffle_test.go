package iplist

import (
	"context"
	"errors"
	"math/big"
	"math/rand"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
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

func collectFromChan(ch <-chan netip.Addr) []netip.Addr {
	var out []netip.Addr
	for s := range ch {
		out = append(out, s)
	}
	return out
}

func addrStrings(addrs []netip.Addr) []string {
	out := make([]string, 0, len(addrs))
	for _, a := range addrs {
		out = append(out, a.String())
	}
	return out
}

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

func TestAddOffsetToAddr_IPv4(t *testing.T) {
	t.Parallel()

	ip := mustParseAddr(t, "192.168.1.1")
	got := addOffsetToAddr(ip, 5)
	want := mustParseAddr(t, "192.168.1.6")

	if got != want {
		t.Fatalf("addOffsetToAddr() = %v, want %v", got, want)
	}
}

func TestAddOffsetToAddr_CarryAcrossOctets(t *testing.T) {
	t.Parallel()

	ip := mustParseAddr(t, "10.0.0.255")
	got := addOffsetToAddr(ip, 1)
	want := mustParseAddr(t, "10.0.1.0")

	if got != want {
		t.Fatalf("addOffsetToAddr() = %v, want %v", got, want)
	}
}

func TestAddOffsetToAddr_IPv6(t *testing.T) {
	t.Parallel()

	ip := mustParseAddr(t, "2001:db8::1")
	got := addOffsetToAddr(ip, 2)
	want := mustParseAddr(t, "2001:db8::3")

	if got != want {
		t.Fatalf("addOffsetToAddr() = %v, want %v", got, want)
	}
}

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

func TestRandBigIntBelow(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewSource(123))
	max := big.NewInt(10)

	for i := 0; i < 1000; i++ {
		n := randBigIntBelow(rng, max)
		if n.Sign() < 0 {
			t.Fatalf("randBigIntBelow returned negative: %v", n)
		}
		if n.Cmp(max) >= 0 {
			t.Fatalf("randBigIntBelow returned %v, want < %v", n, max)
		}
	}
}

func TestGetIPFromCIDRBlocks(t *testing.T) {
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

			got := mi.getIPFromCIDRBlocks(tt.idx)
			if got.String() != tt.want {
				t.Fatalf("getIPFromCIDRBlocks(%d) = %v, want %v", tt.idx, got, tt.want)
			}
		})
	}
}

func TestReadIPAtCSVOffset(t *testing.T) {
	t.Parallel()

	content := "" +
		"1.1.1.1,1\n" +
		"2.2.2.2,0\n" +
		"3.3.3.3,1\n"

	path := mustWriteShuffleFile(t, "ips.csv", content)

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	defer func() {
		_ = f.Close()
	}()

	offset := int64(len("1.1.1.1,1\n"))
	got, err := readIPAtCSVOffset(f, offset)
	if err != nil {
		t.Fatalf("readIPAtCSVOffset: %v", err)
	}

	if got.String() != "2.2.2.2" {
		t.Fatalf("readIPAtCSVOffset() = %v, want %v", got, "2.2.2.2")
	}
}

func TestReadIPAtCSVOffset_InvalidIP(t *testing.T) {
	t.Parallel()

	content := "not-an-ip,1\n"
	path := mustWriteShuffleFile(t, "ips.csv", content)

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	defer func() {
		_ = f.Close()
	}()

	_, err = readIPAtCSVOffset(f, 0)
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
}

func TestNewMasterIndexer_EmptyFile(t *testing.T) {
	t.Parallel()

	path := mustWriteShuffleFile(t, "empty.csv", "")

	idx, err := NewMasterIndexer(path)
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

func TestNewMasterIndexer_SinglesOnly(t *testing.T) {
	t.Parallel()

	content := "" +
		"1.1.1.1,1\n" +
		"2.2.2.2,1\n" +
		"3.3.3.3,0\n" +
		"bad-ip,1\n"

	path := mustWriteShuffleFile(t, "single.csv", content)

	idx, err := NewMasterIndexer(path)
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

func TestNewMasterIndexer_CIDRsAndSingles(t *testing.T) {
	t.Parallel()

	content := "" +
		"10.0.0.1,1\n" +
		"192.168.0.0/30,1\n" +
		"192.168.1.0/31,1\n" +
		"10.0.0.2,0\n"

	path := mustWriteShuffleFile(t, "mixed.csv", content)

	idx, err := NewMasterIndexer(path)
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

func TestStreamActiveIPsShuffled_Empty(t *testing.T) {
	t.Parallel()

	path := mustWriteShuffleFile(t, "empty.csv", "")

	out := make(chan netip.Addr, 10)
	errCh := make(chan error, 1)

	go func() {
		errCh <- streamActiveIPsShuffled(context.Background(), path, 0, out)
		close(out)
	}()

	got := collectFromChan(out)
	err := <-errCh
	if err != nil {
		t.Fatalf("streamActiveIPsShuffled: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("len(got) = %d, want 0", len(got))
	}
}

func TestStreamActiveIPsShuffled_SingleIPs_NoDuplicates(t *testing.T) {
	t.Parallel()

	content := "" +
		"1.1.1.1,1\n" +
		"2.2.2.2,1\n" +
		"3.3.3.3,1\n"

	path := mustWriteShuffleFile(t, "single.csv", content)

	out := make(chan netip.Addr, 10)
	errCh := make(chan error, 1)

	go func() {
		errCh <- streamActiveIPsShuffled(context.Background(), path, 0, out)
		close(out)
	}()

	got := collectFromChan(out)
	err := <-errCh
	if err != nil {
		t.Fatalf("streamActiveIPsShuffled: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("len(got) = %d, want 3", len(got))
	}

	gotStr := addrStrings(got)
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

func TestStreamActiveIPsShuffled_MixedDataset(t *testing.T) {
	t.Parallel()

	content := "" +
		"10.0.0.1,1\n" +
		"192.168.1.0/30,1\n" +
		"10.0.0.2,0\n" +
		"bad-row,1\n"

	path := mustWriteShuffleFile(t, "mixed.csv", content)

	out := make(chan netip.Addr, 10)
	errCh := make(chan error, 1)

	go func() {
		errCh <- streamActiveIPsShuffled(context.Background(), path, 0, out)
		close(out)
	}()

	got := collectFromChan(out)
	err := <-errCh
	if err != nil {
		t.Fatalf("streamActiveIPsShuffled: %v", err)
	}

	gotStr := addrStrings(got)
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

func TestStreamActiveIPsShuffled_Limit(t *testing.T) {
	t.Parallel()

	content := "" +
		"1.1.1.1,1\n" +
		"2.2.2.2,1\n" +
		"3.3.3.3,1\n" +
		"4.4.4.4,1\n"

	path := mustWriteShuffleFile(t, "limit.csv", content)

	out := make(chan netip.Addr, 10)
	errCh := make(chan error, 1)

	go func() {
		errCh <- streamActiveIPsShuffled(context.Background(), path, 2, out)
		close(out)
	}()

	got := collectFromChan(out)
	err := <-errCh
	if err != nil {
		t.Fatalf("streamActiveIPsShuffled: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
}

func TestStreamActiveIPsShuffled_ContextCanceled(t *testing.T) {
	t.Parallel()

	content := "10.0.0.0/24,1\n"
	path := mustWriteShuffleFile(t, "cancel.csv", content)

	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan netip.Addr)
	errCh := make(chan error, 1)

	go func() {
		errCh <- streamActiveIPsShuffled(ctx, path, 0, out)
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
