package httpprobe

import (
	"net/netip"
	"strconv"
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

// entryOf wraps an address in the single-address row a probe now receives.
func entryOf(addr netip.Addr) target.Entry {
	return target.NewEnabled(netip.PrefixFrom(addr, addr.BitLen()))
}

func TestHTTPResult_Key(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want string
	}{
		{"IPv4", "192.168.1.1", "192.168.1.1/32:0:"},
		{"IPv6", "2001:db8::1", "2001:db8::1/128:0:"},
		{"loopback", "127.0.0.1", "127.0.0.1/32:0:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := HTTPResult{Entry: entryOf(netip.MustParseAddr(tt.ip))}
			if got := r.Key(); got != tt.want {
				t.Errorf("Key() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHTTPResult_KeyType(t *testing.T) {
	r := HTTPResult{}
	if got := r.KeyType(); got != target.KindIP {
		t.Errorf("KeyType() = %v, want KeyIP", got)
	}
}

func TestHTTPResult_Equal(t *testing.T) {
	tests := []struct {
		name  string
		r     HTTPResult
		other result.Result
		want  bool
	}{
		{
			name:  "same IP",
			r:     HTTPResult{Entry: entryOf(netip.MustParseAddr("10.0.0.1"))},
			other: HTTPResult{Entry: entryOf(netip.MustParseAddr("10.0.0.1")), StatusCode: 200},
			want:  true,
		},
		{
			name:  "different IP",
			r:     HTTPResult{Entry: entryOf(netip.MustParseAddr("10.0.0.1"))},
			other: HTTPResult{Entry: entryOf(netip.MustParseAddr("10.0.0.2"))},
			want:  false,
		},
		{
			name:  "same IPv6",
			r:     HTTPResult{Entry: entryOf(netip.MustParseAddr("::1"))},
			other: HTTPResult{Entry: entryOf(netip.MustParseAddr("::1"))},
			want:  true,
		},
		{
			name:  "different IPv6",
			r:     HTTPResult{Entry: entryOf(netip.MustParseAddr("::1"))},
			other: HTTPResult{Entry: entryOf(netip.MustParseAddr("::2"))},
			want:  false,
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

func TestHTTPResult_ToRecord_Full(t *testing.T) {
	r := HTTPResult{
		Entry:       entryOf(netip.MustParseAddr("93.184.216.34")),
		Latency:     150 * time.Millisecond,
		StatusCode:  200,
		HTTPVersion: "HTTP/2.0",
		UseTLS:      true,
	}

	rec := r.ToRecord()

	want := []string{"93.184.216.34/32", "0", "", "1", "150.00ms", "200", "HTTP/2.0", "true"}
	if len(rec) != len(want) {
		t.Fatalf("len = %d, want %d", len(rec), len(want))
	}

	for i, w := range want {
		if rec[i] != w {
			t.Errorf("record[%d] = %q, want %q", i, rec[i], w)
		}
	}
}

func TestHTTPResult_ToRecord_NoTLS(t *testing.T) {
	r := HTTPResult{
		Entry:       entryOf(netip.MustParseAddr("10.0.0.1")),
		Latency:     25 * time.Millisecond,
		StatusCode:  301,
		HTTPVersion: "HTTP/1.1",
		UseTLS:      false,
	}

	rec := r.ToRecord()

	if rec[5] != "301" {
		t.Errorf("status = %q, want %q", rec[5], "301")
	}
	if rec[7] != "false" {
		t.Errorf("tls = %q, want %q", rec[7], "false")
	}
}

// TestHTTPResult_ToRecord_ZeroStatusCode verifies that when StatusCode is 0,
// status and TLS fields are rendered as "-" regardless of the UseTLS value.
func TestHTTPResult_ToRecord_ZeroStatusCode(t *testing.T) {
	r := HTTPResult{
		Entry:       entryOf(netip.MustParseAddr("1.2.3.4")),
		Latency:     10 * time.Millisecond,
		StatusCode:  0,
		HTTPVersion: "HTTP/1.1",
		UseTLS:      true,
	}

	rec := r.ToRecord()

	if rec[5] != "-" {
		t.Errorf("status = %q, want %q for zero StatusCode", rec[5], "-")
	}
	if rec[7] != "-" {
		t.Errorf("tls = %q, want %q for zero StatusCode", rec[7], "-")
	}
}

func TestHTTPResult_ToRecord_VariousStatusCodes(t *testing.T) {
	codes := []int{100, 204, 301, 403, 404, 500, 503}
	for _, code := range codes {
		r := HTTPResult{
			Entry:       entryOf(netip.MustParseAddr("1.2.3.4")),
			Latency:     time.Millisecond,
			StatusCode:  code,
			HTTPVersion: "HTTP/1.1",
			UseTLS:      true,
		}
		rec := r.ToRecord()
		want := strconv.Itoa(code)
		if rec[5] != want {
			t.Errorf("StatusCode %d: record[5] = %q, want %q", code, rec[5], want)
		}
	}
}

func TestHTTPResult_Score(t *testing.T) {
	tests := []struct {
		name    string
		latency time.Duration
		want    float64
	}{
		{"100ms", 100 * time.Millisecond, 10.0},
		{"50ms", 50 * time.Millisecond, 20.0},
		{"10ms", 10 * time.Millisecond, 100.0},
		{"1ms", 1 * time.Millisecond, 1000.0},
		{"200ms", 200 * time.Millisecond, 5.0},
		{"1000ms", 1000 * time.Millisecond, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := HTTPResult{Latency: tt.latency}
			got := r.Score()
			if diff := got - tt.want; diff > 0.001 || diff < -0.001 {
				t.Errorf("Score() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestHTTPResult_Score_SubMillisecond_Clamped(t *testing.T) {
	r := HTTPResult{Latency: 500 * time.Microsecond}
	got := r.Score()
	if got != 1000.0 {
		t.Errorf("Score() = %f, want 1000.0 (clamped)", got)
	}
}

func TestHTTPResult_Score_ZeroLatency_Clamped(t *testing.T) {
	r := HTTPResult{Latency: 0}
	got := r.Score()
	if got != 1000.0 {
		t.Errorf("Score() = %f, want 1000.0 (clamped)", got)
	}
}

func TestHTTPResult_Score_Ordering(t *testing.T) {
	fast := HTTPResult{Latency: 10 * time.Millisecond}
	slow := HTTPResult{Latency: 500 * time.Millisecond}

	if fast.Score() <= slow.Score() {
		t.Errorf("fast (%f) should score higher than slow (%f)", fast.Score(), slow.Score())
	}
}

func TestParseHTTPResult_FullRecord(t *testing.T) {
	record := []string{"93.184.216.34", "0", "", "1", "150.00ms", "200", "HTTP/2.0", "true"}

	res, err := parseHTTPResult(record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hr := res.(HTTPResult)

	if hr.Entry.IP.Addr() != netip.MustParseAddr("93.184.216.34") {
		t.Errorf("IP = %v, want 93.184.216.34", hr.Entry.IP.Addr())
	}
	if hr.Latency != 150*time.Millisecond {
		t.Errorf("Latency = %v, want 150ms", hr.Latency)
	}
	if hr.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", hr.StatusCode)
	}
	if hr.HTTPVersion != "HTTP/2.0" {
		t.Errorf("HTTPVersion = %q, want %q", hr.HTTPVersion, "HTTP/2.0")
	}
	if !hr.UseTLS {
		t.Error("UseTLS = false, want true")
	}
}

// TestParseHTTPResult_LegacyShortRecordRejected verifies that records written
// before the entry columns were added are no longer accepted.
func TestParseHTTPResult_LegacyShortRecordRejected(t *testing.T) {
	record := []string{"10.0.0.1", "42ms"}

	if _, err := parseHTTPResult(record); err == nil {
		t.Fatal("expected error for legacy two-field record")
	}
}

func TestParseHTTPResult_FiveFieldsRejected(t *testing.T) {
	record := []string{"1.2.3.4", "10ms", "200", "HTTP/1.1", "true"}

	if _, err := parseHTTPResult(record); err == nil {
		t.Fatal("expected error for five-field record")
	}
}

func TestParseHTTPResult_NoTLS(t *testing.T) {
	record := []string{"172.16.0.1", "0", "", "1", "30ms", "404", "HTTP/1.1", "false"}

	res, err := parseHTTPResult(record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hr := res.(HTTPResult)
	if hr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", hr.StatusCode)
	}
	if hr.UseTLS {
		t.Error("UseTLS = true, want false")
	}
	if hr.HTTPVersion != "HTTP/1.1" {
		t.Errorf("HTTPVersion = %q, want %q", hr.HTTPVersion, "HTTP/1.1")
	}
}

func TestParseHTTPResult_LatencyRounded(t *testing.T) {
	record := []string{"1.2.3.4", "0", "", "1", "42.6ms", "200", "HTTP/1.1", "true"}

	res, err := parseHTTPResult(record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hr := res.(HTTPResult)
	if hr.Latency != 43*time.Millisecond {
		t.Errorf("Latency = %v, want 43ms (rounded)", hr.Latency)
	}
}

func TestParseHTTPResult_LatencyClampedToMinimum(t *testing.T) {
	record := []string{"1.2.3.4", "0", "", "1", "0s", "200", "HTTP/1.1", "true"}

	res, err := parseHTTPResult(record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hr := res.(HTTPResult)
	if hr.Latency != time.Millisecond {
		t.Errorf("Latency = %v, want 1ms (clamped)", hr.Latency)
	}
}

func TestParseHTTPResult_TooFewFields(t *testing.T) {
	_, err := parseHTTPResult([]string{"1.2.3.4"})
	if err == nil {
		t.Fatal("expected error for 1 field")
	}
}

func TestParseHTTPResult_EmptyRecord(t *testing.T) {
	_, err := parseHTTPResult(nil)
	if err == nil {
		t.Fatal("expected error for nil record")
	}
}

func TestParseHTTPResult_InvalidIP(t *testing.T) {
	_, err := parseHTTPResult([]string{"not-an-ip", "0", "", "1", "42ms", "200", "HTTP/1.1", "true"})
	if err == nil {
		t.Fatal("expected error for invalid IP")
	}
}

func TestParseHTTPResult_InvalidLatency(t *testing.T) {
	_, err := parseHTTPResult([]string{"1.2.3.4", "0", "", "1", "garbage", "200", "HTTP/1.1", "true"})
	if err == nil {
		t.Fatal("expected error for invalid latency")
	}
}

func TestParseHTTPResult_InvalidStatusCode(t *testing.T) {
	record := []string{"1.2.3.4", "0", "", "1", "10ms", "not-a-number", "HTTP/1.1", "true"}
	_, err := parseHTTPResult(record)
	if err == nil {
		t.Fatal("expected error for invalid status code")
	}
}

func TestParseHTTPResult_InvalidTLS(t *testing.T) {
	record := []string{"1.2.3.4", "0", "", "1", "10ms", "200", "HTTP/1.1", "maybe"}
	_, err := parseHTTPResult(record)
	if err == nil {
		t.Fatal("expected error for invalid TLS bool")
	}
}

func TestParseHTTPResult_IPv6(t *testing.T) {
	record := []string{"2001:db8::42", "0", "", "1", "55ms", "200", "HTTP/2.0", "true"}

	res, err := parseHTTPResult(record)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hr := res.(HTTPResult)
	if hr.Entry.IP.Addr() != netip.MustParseAddr("2001:db8::42") {
		t.Errorf("IP = %v, want 2001:db8::42", hr.Entry.IP.Addr())
	}
}

// TestHTTPResult_RoundTrip_Full ensures that serializing and parsing a result
// yields the original values, validating the symmetry of ToRecord and parseHTTPResult.
func TestHTTPResult_RoundTrip_Full(t *testing.T) {
	original := HTTPResult{
		Entry:       entryOf(netip.MustParseAddr("104.16.132.229")),
		Latency:     87 * time.Millisecond,
		StatusCode:  200,
		HTTPVersion: "HTTP/2.0",
		UseTLS:      true,
	}

	record := original.ToRecord()
	parsed, err := parseHTTPResult(record)
	if err != nil {
		t.Fatalf("round-trip parse failed: %v", err)
	}

	got := parsed.(HTTPResult)

	if got.Entry.IP.Addr() != original.Entry.IP.Addr() {
		t.Errorf("IP = %v, want %v", got.Entry.IP.Addr(), original.Entry.IP.Addr())
	}
	if got.Latency != original.Latency {
		t.Errorf("Latency = %v, want %v", got.Latency, original.Latency)
	}
	if got.StatusCode != original.StatusCode {
		t.Errorf("StatusCode = %d, want %d", got.StatusCode, original.StatusCode)
	}
	if got.HTTPVersion != original.HTTPVersion {
		t.Errorf("HTTPVersion = %q, want %q", got.HTTPVersion, original.HTTPVersion)
	}
	if got.UseTLS != original.UseTLS {
		t.Errorf("UseTLS = %v, want %v", got.UseTLS, original.UseTLS)
	}
}

// TestHTTPResult_RoundTrip_ZeroStatus verifies that failed requests survive a
// serialization round trip: "-" is parsed back as a zero status.
func TestHTTPResult_RoundTrip_ZeroStatus(t *testing.T) {
	original := HTTPResult{
		Entry:       entryOf(netip.MustParseAddr("8.8.8.8")),
		Latency:     12 * time.Millisecond,
		StatusCode:  0,
		HTTPVersion: "HTTP/1.1",
		UseTLS:      false,
	}

	record := original.ToRecord()

	parsed, err := parseHTTPResult(record)
	if err != nil {
		t.Fatalf("round-trip parse failed: %v", err)
	}

	got := parsed.(HTTPResult)
	if got.StatusCode != 0 {
		t.Errorf("StatusCode = %d, want 0", got.StatusCode)
	}
	if got.UseTLS {
		t.Error("UseTLS = true, want false")
	}
}

func TestHTTPResult_RoundTrip_IPv6(t *testing.T) {
	original := HTTPResult{
		Entry:       entryOf(netip.MustParseAddr("::1")),
		Latency:     3 * time.Millisecond,
		StatusCode:  204,
		HTTPVersion: "HTTP/1.1",
		UseTLS:      false,
	}

	record := original.ToRecord()
	parsed, err := parseHTTPResult(record)
	if err != nil {
		t.Fatalf("round-trip parse failed: %v", err)
	}

	got := parsed.(HTTPResult)
	if got.Entry.IP.Addr() != original.Entry.IP.Addr() {
		t.Errorf("IP = %v, want %v", got.Entry.IP.Addr(), original.Entry.IP.Addr())
	}
}

func TestSchema_Fields(t *testing.T) {
	if Schema.Name != "HTTP" {
		t.Errorf("Name = %q, want %q", Schema.Name, "HTTP")
	}
	if Schema.Directory != "http" {
		t.Errorf("Directory = %q, want %q", Schema.Directory, "http")
	}
}

func TestSchema_Columns(t *testing.T) {
	want := []struct {
		name  string
		width int
	}{
		{"Target", 39},
		{"Latency", 19},
		{"Status", 12},
		{"Version", 20},
		{"TLS", 10},
	}

	if len(Schema.Columns) != len(want) {
		t.Fatalf("Columns len = %d, want %d", len(Schema.Columns), len(want))
	}

	for i, w := range want {
		if Schema.Columns[i].Name != w.name {
			t.Errorf("Columns[%d].Name = %q, want %q", i, Schema.Columns[i].Name, w.name)
		}
		if Schema.Columns[i].Width != w.width {
			t.Errorf("Columns[%d].Width = %d, want %d", i, Schema.Columns[i].Width, w.width)
		}
	}
}

func TestSchema_ParserNotNil(t *testing.T) {
	if Schema.Parser == nil {
		t.Fatal("Schema.Parser is nil")
	}
}

func TestSchema_ParserWorks(t *testing.T) {
	res, err := Schema.Parser([]string{"1.1.1.1", "0", "", "1", "20ms", "200", "HTTP/1.1", "true"})
	if err != nil {
		t.Fatalf("Schema.Parser error: %v", err)
	}
	hr := res.(HTTPResult)
	if hr.Entry.IP.Addr() != netip.MustParseAddr("1.1.1.1") {
		t.Errorf("IP = %v, want 1.1.1.1", hr.Entry.IP.Addr())
	}
	if hr.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", hr.StatusCode)
	}
}
