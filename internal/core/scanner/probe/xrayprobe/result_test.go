package xrayprobe

import (
	"net/netip"
	"testing"
	"time"

	"bgscan/internal/core/result"
	"bgscan/internal/core/speedtest"
)

func TestXrayResult_Key(t *testing.T) {
	r := XrayResult{IP: netip.MustParseAddr("192.168.1.1")}
	if got := r.Key(); got != "192.168.1.1" {
		t.Errorf("Key() = %q, want %q", got, "192.168.1.1")
	}
}

func TestXrayResult_Key_IPv6(t *testing.T) {
	r := XrayResult{IP: netip.MustParseAddr("::1")}
	if got := r.Key(); got != "::1" {
		t.Errorf("Key() = %q, want %q", got, "::1")
	}
}

func TestXrayResult_KeyType(t *testing.T) {
	r := XrayResult{}
	if got := r.KeyType(); got != result.KeyIP {
		t.Errorf("KeyType() = %v, want KeyIP", got)
	}
}

func TestXrayResult_Equal(t *testing.T) {
	tests := []struct {
		name  string
		r     XrayResult
		other result.Result
		want  bool
	}{
		{
			name:  "same IP",
			r:     XrayResult{IP: netip.MustParseAddr("1.2.3.4")},
			other: XrayResult{IP: netip.MustParseAddr("1.2.3.4"), Latency: 99 * time.Second},
			want:  true,
		},
		{
			name:  "different IP",
			r:     XrayResult{IP: netip.MustParseAddr("1.2.3.4")},
			other: XrayResult{IP: netip.MustParseAddr("5.6.7.8")},
			want:  false,
		},
		{
			name:  "same IPv6",
			r:     XrayResult{IP: netip.MustParseAddr("2001:db8::1")},
			other: XrayResult{IP: netip.MustParseAddr("2001:db8::1")},
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

func TestXrayResult_ToRecord(t *testing.T) {
	r := XrayResult{
		IP:       netip.MustParseAddr("10.0.0.1"),
		Latency:  42 * time.Millisecond,
		Download: 100_000_000, // 100 Mbps
		Upload:   50_000_000,  // 50 Mbps
	}

	rec := r.ToRecord()

	if len(rec) != 4 {
		t.Fatalf("ToRecord() len = %d, want 4", len(rec))
	}
	if rec[0] != "10.0.0.1" {
		t.Errorf("record[0] = %q, want %q", rec[0], "10.0.0.1")
	}
	if rec[1] != "42ms" {
		t.Errorf("record[1] = %q, want %q", rec[1], "42ms")
	}

	wantDL := speedtest.BitsPerSec(100_000_000).String()
	if rec[2] != wantDL {
		t.Errorf("record[2] = %q, want %q", rec[2], wantDL)
	}
	wantUL := speedtest.BitsPerSec(50_000_000).String()
	if rec[3] != wantUL {
		t.Errorf("record[3] = %q, want %q", rec[3], wantUL)
	}
}

func TestXrayResult_ToRecord_ZeroValues(t *testing.T) {
	r := XrayResult{
		IP:      netip.MustParseAddr("1.1.1.1"),
		Latency: 5 * time.Millisecond,
		// Download and Upload are zero (ConnectivityOnly mode)
	}

	rec := r.ToRecord()
	if rec[2] != speedtest.BitsPerSec(0).String() {
		t.Errorf("record[2] = %q, want zero value string", rec[2])
	}
	if rec[3] != speedtest.BitsPerSec(0).String() {
		t.Errorf("record[3] = %q, want zero value string", rec[3])
	}
}

func TestXrayResult_Score(t *testing.T) {
	tests := []struct {
		name    string
		r       XrayResult
		checkFn func(t *testing.T, score float64)
	}{
		{
			name: "balanced result",
			r: XrayResult{
				IP:       netip.MustParseAddr("1.2.3.4"),
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
				IP:      netip.MustParseAddr("1.2.3.4"),
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
				IP:       netip.MustParseAddr("1.2.3.4"),
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
				IP:      netip.MustParseAddr("1.2.3.4"),
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
				IP:       netip.MustParseAddr("1.2.3.4"),
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
		IP:       netip.MustParseAddr("1.1.1.1"),
		Latency:  20 * time.Millisecond,
		Download: speedtest.BitsPerSec(200) * speedtest.Mbps,
		Upload:   speedtest.BitsPerSec(100) * speedtest.Mbps,
	}
	slow := XrayResult{
		IP:       netip.MustParseAddr("2.2.2.2"),
		Latency:  300 * time.Millisecond,
		Download: speedtest.BitsPerSec(10) * speedtest.Mbps,
		Upload:   speedtest.BitsPerSec(5) * speedtest.Mbps,
	}

	if fast.Score() <= slow.Score() {
		t.Errorf("fast.Score() = %f should be > slow.Score() = %f", fast.Score(), slow.Score())
	}
}

func TestParseXrayResult_Valid(t *testing.T) {
	dl := speedtest.BitsPerSec(100_000_000)
	ul := speedtest.BitsPerSec(50_000_000)

	record := []string{
		"192.168.1.100",
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

	if xr.IP != netip.MustParseAddr("192.168.1.100") {
		t.Errorf("IP = %v, want 192.168.1.100", xr.IP)
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

func TestParseXrayResult_LatencyRounded(t *testing.T) {
	record := []string{"1.2.3.4", "42.7ms", "0", "0"}

	res, err := parseXrayResult(record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xr := res.(XrayResult)
	if xr.Latency != 43*time.Millisecond {
		t.Errorf("Latency = %v, want 43ms (rounded)", xr.Latency)
	}
}

func TestParseXrayResult_LatencyClampedToMinimum(t *testing.T) {
	record := []string{"1.2.3.4", "0s", "0", "0"}

	res, err := parseXrayResult(record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xr := res.(XrayResult)
	if xr.Latency != time.Millisecond {
		t.Errorf("Latency = %v, want 1ms (clamped)", xr.Latency)
	}
}

func TestParseXrayResult_TooFewFields(t *testing.T) {
	_, err := parseXrayResult([]string{"1.2.3.4", "42ms", "100"})
	if err == nil {
		t.Fatal("expected error for 3 fields")
	}
}

func TestParseXrayResult_EmptyRecord(t *testing.T) {
	_, err := parseXrayResult(nil)
	if err == nil {
		t.Fatal("expected error for nil record")
	}
}

func TestParseXrayResult_InvalidIP(t *testing.T) {
	_, err := parseXrayResult([]string{"not-an-ip", "42ms", "0", "0"})
	if err == nil {
		t.Fatal("expected error for invalid IP")
	}
}

func TestParseXrayResult_InvalidLatency(t *testing.T) {
	_, err := parseXrayResult([]string{"1.2.3.4", "not-a-duration", "0", "0"})
	if err == nil {
		t.Fatal("expected error for invalid latency")
	}
}

// TestParseXrayResult_InvalidSpeed_Ignored verifies that unparseable speed
// values are silently discarded and default to zero, matching the production behavior.
func TestParseXrayResult_InvalidSpeed_Ignored(t *testing.T) {
	record := []string{"1.2.3.4", "10ms", "garbage", "also-garbage"}

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
		IP:       netip.MustParseAddr("172.16.0.1"),
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

	if got.IP != original.IP {
		t.Errorf("IP = %v, want %v", got.IP, original.IP)
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

func TestXrayResult_RoundTrip_IPv6(t *testing.T) {
	original := XrayResult{
		IP:      netip.MustParseAddr("2001:db8::42"),
		Latency: 7 * time.Millisecond,
	}

	record := original.ToRecord()
	parsed, err := parseXrayResult(record)
	if err != nil {
		t.Fatalf("round-trip parse failed: %v", err)
	}

	got := parsed.(XrayResult)
	if got.IP != original.IP {
		t.Errorf("IP = %v, want %v", got.IP, original.IP)
	}
}

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
		{"IP", 40},
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

func TestSchema_ParserNotNil(t *testing.T) {
	if Schema.Parser == nil {
		t.Fatal("Schema.Parser is nil")
	}
}

func TestSchema_ParserWorks(t *testing.T) {
	res, err := Schema.Parser([]string{"8.8.8.8", "15ms", "0", "0"})
	if err != nil {
		t.Fatalf("Schema.Parser error: %v", err)
	}
	xr := res.(XrayResult)
	if xr.IP != netip.MustParseAddr("8.8.8.8") {
		t.Errorf("IP = %v, want 8.8.8.8", xr.IP)
	}
}
