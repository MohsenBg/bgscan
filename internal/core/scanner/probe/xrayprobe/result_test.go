package xrayprobe

import (
	"net/netip"
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/speedtest"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

// entryOf wraps an address in the single-address row a probe now receives.
func entryOf(addr netip.Addr) target.Entry {
	return target.NewEnabled(netip.PrefixFrom(addr, addr.BitLen()))
}

// TestXrayResult_Key covers the xray result key scenario.
func TestXrayResult_Key(t *testing.T) {
	r := XrayResult{Entry: entryOf(netip.MustParseAddr("192.168.1.1"))}
	if got := r.Key(); got != "192.168.1.1/32:0:" {
		t.Errorf("Key() = %q, want %q", got, "0:192.168.1.1/32:0:")
	}
}

// TestXrayResult_Key_IPv6 covers the xray result key ipv6 scenario.
func TestXrayResult_Key_IPv6(t *testing.T) {
	r := XrayResult{Entry: entryOf(netip.MustParseAddr("::1"))}
	if got := r.Key(); got != "::1/128:0:" {
		t.Errorf("Key() = %q, want %q", got, "::1/128:0:")
	}
}

// TestXrayResult_KeyType covers the xray result key type scenario.
func TestXrayResult_KeyType(t *testing.T) {
	r := XrayResult{}
	if got := r.KeyType(); got != target.KindIP {
		t.Errorf("KeyType() = %v, want KeyIP", got)
	}
}

// TestXrayResult_Equal covers the xray result equal scenario.
func TestXrayResult_Equal(t *testing.T) {
	tests := []struct {
		name  string
		r     XrayResult
		other result.Result
		want  bool
	}{
		{
			name:  "same IP",
			r:     XrayResult{Entry: entryOf(netip.MustParseAddr("1.2.3.4"))},
			other: XrayResult{Entry: entryOf(netip.MustParseAddr("1.2.3.4")), Latency: 99 * time.Second},
			want:  true,
		},
		{
			name:  "different IP",
			r:     XrayResult{Entry: entryOf(netip.MustParseAddr("1.2.3.4"))},
			other: XrayResult{Entry: entryOf(netip.MustParseAddr("5.6.7.8"))},
			want:  false,
		},
		{
			name:  "same IPv6",
			r:     XrayResult{Entry: entryOf(netip.MustParseAddr("2001:db8::1"))},
			other: XrayResult{Entry: entryOf(netip.MustParseAddr("2001:db8::1"))},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Equal(tt.other); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestXrayResult_ToRecord covers the xray result to record scenario.
func TestXrayResult_ToRecord(t *testing.T) {
	r := XrayResult{
		Entry:    entryOf(netip.MustParseAddr("10.0.0.1")),
		Latency:  42 * time.Millisecond,
		Download: 100_000_000, // 100 Mbps
		Upload:   50_000_000,  // 50 Mbps
	}

	rec := r.ToRecord()

	if len(rec) != 7 {
		t.Fatalf("ToRecord() len = %d, want 7", len(rec))
	}
	if rec[0] != "10.0.0.1/32" {
		t.Errorf("record[0] = %q, want %q", rec[0], "10.0.0.1/32")
	}
	if rec[4] != "42.00ms" {
		t.Errorf("record[4] = %q, want %q", rec[4], "42.00ms")
	}

	wantDL := speedtest.BitsPerSec(100_000_000).String()
	if rec[5] != wantDL {
		t.Errorf("record[5] = %q, want %q", rec[5], wantDL)
	}
	wantUL := speedtest.BitsPerSec(50_000_000).String()
	if rec[6] != wantUL {
		t.Errorf("record[6] = %q, want %q", rec[6], wantUL)
	}
}

// TestXrayResult_ToRecord_ZeroValues covers the xray result to record zero values scenario.
func TestXrayResult_ToRecord_ZeroValues(t *testing.T) {
	r := XrayResult{
		Entry:   entryOf(netip.MustParseAddr("1.1.1.1")),
		Latency: 5 * time.Millisecond,
		// Download and Upload are zero (ConnectivityOnly mode)
	}

	rec := r.ToRecord()
	if rec[5] != speedtest.BitsPerSec(0).String() {
		t.Errorf("record[5] = %q, want zero value string", rec[5])
	}
	if rec[6] != speedtest.BitsPerSec(0).String() {
		t.Errorf("record[6] = %q, want zero value string", rec[6])
	}
}

// TestXrayResult_Score covers the xray result score scenario.
func TestXrayResult_Score(t *testing.T) {
	tests := []struct {
		name    string
		r       XrayResult
		checkFn func(t *testing.T, score float64)
	}{
		{
			name: "balanced result",
			r: XrayResult{
				Entry:    entryOf(netip.MustParseAddr("1.2.3.4")),
				Latency:  100 * time.Millisecond,
				Download: speedtest.BitsPerSec(100) * speedtest.Mbps,
				Upload:   speedtest.BitsPerSec(50) * speedtest.Mbps,
			},
			checkFn: func(t *testing.T, score float64) {
				want := 76.0
				if diff := score - want; diff > 0.001 || diff < -0.001 {
					t.Errorf("Score() = %f, want %f", score, want)
				}
			},
		},
		{
			name: "connectivity only - zero speeds",
			r: XrayResult{
				Entry:   entryOf(netip.MustParseAddr("1.2.3.4")),
				Latency: 50 * time.Millisecond,
			},
			checkFn: func(t *testing.T, score float64) {
				want := 2.0
				if diff := score - want; diff > 0.001 || diff < -0.001 {
					t.Errorf("Score() = %f, want %f", score, want)
				}
			},
		},
		{
			name: "sub-millisecond latency clamped to 1ms",
			r: XrayResult{
				Entry:    entryOf(netip.MustParseAddr("1.2.3.4")),
				Latency:  500 * time.Microsecond,
				Download: speedtest.BitsPerSec(10) * speedtest.Mbps,
			},
			checkFn: func(t *testing.T, score float64) {
				want := 106.0
				if diff := score - want; diff > 0.001 || diff < -0.001 {
					t.Errorf("Score() = %f, want %f", score, want)
				}
			},
		},
		{
			name: "zero latency clamped to 1ms",
			r: XrayResult{
				Entry:   entryOf(netip.MustParseAddr("1.2.3.4")),
				Latency: 0,
			},
			checkFn: func(t *testing.T, score float64) {
				want := 100.0
				if diff := score - want; diff > 0.001 || diff < -0.001 {
					t.Errorf("Score() = %f, want %f", score, want)
				}
			},
		},
		{
			name: "high latency low score",
			r: XrayResult{
				Entry:    entryOf(netip.MustParseAddr("1.2.3.4")),
				Latency:  1000 * time.Millisecond,
				Download: speedtest.BitsPerSec(10) * speedtest.Mbps,
				Upload:   speedtest.BitsPerSec(5) * speedtest.Mbps,
			},
			checkFn: func(t *testing.T, score float64) {
				want := 7.6
				if diff := score - want; diff > 0.001 || diff < -0.001 {
					t.Errorf("Score() = %f, want %f", score, want)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.checkFn(t, tt.r.Score())
		})
	}
}

// TestXrayResult_Score_Ordering verifies that a fast server with good throughput
// scores higher than a slow server, ensuring the weighting logic behaves as expected.
func TestXrayResult_Score_Ordering(t *testing.T) {
	fast := XrayResult{
		Entry:    entryOf(netip.MustParseAddr("1.1.1.1")),
		Latency:  20 * time.Millisecond,
		Download: speedtest.BitsPerSec(200) * speedtest.Mbps,
		Upload:   speedtest.BitsPerSec(100) * speedtest.Mbps,
	}
	slow := XrayResult{
		Entry:    entryOf(netip.MustParseAddr("2.2.2.2")),
		Latency:  300 * time.Millisecond,
		Download: speedtest.BitsPerSec(10) * speedtest.Mbps,
		Upload:   speedtest.BitsPerSec(5) * speedtest.Mbps,
	}

	if fast.Score() <= slow.Score() {
		t.Errorf("fast.Score() = %f should be > slow.Score() = %f", fast.Score(), slow.Score())
	}
}

// TestParseXrayResult_Valid covers the parse xray result valid scenario.
func TestParseXrayResult_Valid(t *testing.T) {
	dl := speedtest.BitsPerSec(100_000_000)
	ul := speedtest.BitsPerSec(50_000_000)

	// New format: entry fields (value, port, path, enable) + result fields (latency, download, upload)
	record := []string{
		"192.168.1.100/32", "0", "", "1",
		"42ms",
		dl.String(),
		ul.String(),
	}

	res, err := parseXrayResult(record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xr, ok := res.(XrayResult)
	if !ok {
		t.Fatalf("expected XrayResult, got %T", res)
	}

	if xr.Entry.IP.Addr() != netip.MustParseAddr("192.168.1.100") {
		t.Errorf("IP = %v, want 192.168.1.100", xr.Entry.IP.Addr())
	}
	if xr.Latency != 42*time.Millisecond {
		t.Errorf("Latency = %v, want 42ms", xr.Latency)
	}
	if xr.Download != dl {
		t.Errorf("Download = %v, want %v", xr.Download, dl)
	}
	if xr.Upload != ul {
		t.Errorf("Upload = %v, want %v", xr.Upload, ul)
	}
}

// TestParseXrayResult_LatencyRounded covers the parse xray result latency rounded scenario.
func TestParseXrayResult_LatencyRounded(t *testing.T) {
	record := []string{"1.2.3.4/32", "0", "", "1", "42.7ms", "0", "0"}

	res, err := parseXrayResult(record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xr := res.(XrayResult)
	if xr.Latency != 43*time.Millisecond {
		t.Errorf("Latency = %v, want 43ms (rounded)", xr.Latency)
	}
}

// TestParseXrayResult_LatencyClampedToMinimum covers the parse xray result latency clamped to minimum scenario.
func TestParseXrayResult_LatencyClampedToMinimum(t *testing.T) {
	record := []string{"1.2.3.4/32", "0", "", "1", "0s", "0", "0"}

	res, err := parseXrayResult(record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xr := res.(XrayResult)
	if xr.Latency != time.Millisecond {
		t.Errorf("Latency = %v, want 1ms (clamped)", xr.Latency)
	}
}

// TestParseXrayResult_TooFewFields covers the parse xray result too few fields scenario.
func TestParseXrayResult_TooFewFields(t *testing.T) {
	_, err := parseXrayResult([]string{"1.2.3.4/32", "0", "", "1", "42ms", "100"})
	if err == nil {
		t.Fatal("expected error for 6 fields (need 7)")
	}
}

// TestParseXrayResult_EmptyRecord covers the parse xray result empty record scenario.
func TestParseXrayResult_EmptyRecord(t *testing.T) {
	_, err := parseXrayResult(nil)
	if err == nil {
		t.Fatal("expected error for nil record")
	}
}

// TestParseXrayResult_InvalidIP covers the parse xray result invalid ip scenario.
func TestParseXrayResult_InvalidIP(t *testing.T) {
	_, err := parseXrayResult([]string{"not-an-ip", "0", "", "1", "42ms", "0", "0"})
	if err == nil {
		t.Fatal("expected error for invalid IP")
	}
}

// TestParseXrayResult_InvalidLatency covers the parse xray result invalid latency scenario.
func TestParseXrayResult_InvalidLatency(t *testing.T) {
	_, err := parseXrayResult([]string{"1.2.3.4/32", "0", "", "1", "not-a-duration", "0", "0"})
	if err == nil {
		t.Fatal("expected error for invalid latency")
	}
}

// TestParseXrayResult_InvalidSpeed_Ignored verifies that unparseable speed
// values are silently discarded and default to zero, matching the production behavior.
func TestParseXrayResult_InvalidSpeed_Ignored(t *testing.T) {
	record := []string{"1.2.3.4/32", "0", "", "1", "10ms", "garbage", "also-garbage"}

	res, err := parseXrayResult(record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xr := res.(XrayResult)
	if xr.Download != 0 {
		t.Errorf("Download = %v, want 0 for unparseable value", xr.Download)
	}
	if xr.Upload != 0 {
		t.Errorf("Upload = %v, want 0 for unparseable value", xr.Upload)
	}
}

// TestXrayResult_RoundTrip ensures that serializing and parsing a result
// yields the original values, validating the symmetry of ToRecord and parseXrayResult.
func TestXrayResult_RoundTrip(t *testing.T) {
	original := XrayResult{
		Entry:    entryOf(netip.MustParseAddr("172.16.0.1")),
		Latency:  123 * time.Millisecond,
		Download: speedtest.BitsPerSec(75_000_000),
		Upload:   speedtest.BitsPerSec(25_000_000),
	}

	record := original.ToRecord()
	parsed, err := parseXrayResult(record)
	if err != nil {
		t.Fatalf("round-trip parse failed: %v", err)
	}

	got := parsed.(XrayResult)

	if got.Entry.IP.Addr() != original.Entry.IP.Addr() {
		t.Errorf("IP = %v, want %v", got.Entry.IP.Addr(), original.Entry.IP.Addr())
	}
	if got.Latency != original.Latency {
		t.Errorf("Latency = %v, want %v", got.Latency, original.Latency)
	}
	if got.Download != original.Download {
		t.Errorf("Download = %v, want %v", got.Download, original.Download)
	}
	if got.Upload != original.Upload {
		t.Errorf("Upload = %v, want %v", got.Upload, original.Upload)
	}
}

// TestXrayResult_RoundTrip_IPv6 covers the xray result round trip ipv6 scenario.
func TestXrayResult_RoundTrip_IPv6(t *testing.T) {
	original := XrayResult{
		Entry:   entryOf(netip.MustParseAddr("2001:db8::42")),
		Latency: 7 * time.Millisecond,
	}

	record := original.ToRecord()
	parsed, err := parseXrayResult(record)
	if err != nil {
		t.Fatalf("round-trip parse failed: %v", err)
	}

	got := parsed.(XrayResult)
	if got.Entry.IP.Addr() != original.Entry.IP.Addr() {
		t.Errorf("IP = %v, want %v", got.Entry.IP.Addr(), original.Entry.IP.Addr())
	}
}

// TestSchema_Fields covers the schema fields scenario.
func TestSchema_Fields(t *testing.T) {
	if Schema.Name != "Xray" {
		t.Errorf("Schema.Name = %q, want %q", Schema.Name, "Xray")
	}
	if Schema.Directory != "xray" {
		t.Errorf("Schema.Directory = %q, want %q", Schema.Directory, "xray")
	}
	if len(Schema.Columns) != 4 {
		t.Fatalf("Schema.Columns len = %d, want 4", len(Schema.Columns))
	}

	wantCols := []struct {
		name  string
		width int
	}{
		{"Target", 40},
		{"Latency", 20},
		{"Download", 20},
		{"Upload", 20},
	}
	for i, wc := range wantCols {
		if Schema.Columns[i].Name != wc.name {
			t.Errorf("Columns[%d].Name = %q, want %q", i, Schema.Columns[i].Name, wc.name)
		}
		if Schema.Columns[i].Width != wc.width {
			t.Errorf("Columns[%d].Width = %d, want %d", i, Schema.Columns[i].Width, wc.width)
		}
	}
}

// TestSchema_ParserNotNil covers the schema parser not nil scenario.
func TestSchema_ParserNotNil(t *testing.T) {
	if Schema.Parser == nil {
		t.Fatal("Schema.Parser is nil")
	}
}

// TestSchema_ParserWorks covers the schema parser works scenario.
func TestSchema_ParserWorks(t *testing.T) {
	res, err := Schema.Parser([]string{"8.8.8.8/32", "0", "", "1", "15ms", "0", "0"})
	if err != nil {
		t.Fatalf("Schema.Parser error: %v", err)
	}
	xr := res.(XrayResult)
	if xr.Entry.IP.Addr() != netip.MustParseAddr("8.8.8.8") {
		t.Errorf("IP = %v, want 8.8.8.8", xr.Entry.IP.Addr())
	}
}
