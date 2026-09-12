package thefeedprobe

import (
	"net/netip"
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
)

func TestSchemaValidate(t *testing.T) {
	if err := Schema.Validate(); err != nil {
		t.Fatalf("Schema.Validate() error = %v", err)
	}

	if Schema.Name != "TheFeed" {
		t.Errorf("schema name = %q, want TheFeed", Schema.Name)
	}
	if Schema.Directory != "thefeed" {
		t.Errorf("schema directory = %q, want thefeed", Schema.Directory)
	}
	if len(Schema.Columns) != 4 {
		t.Errorf("schema columns = %d, want 4", len(Schema.Columns))
	}
	if Schema.Parser == nil {
		t.Fatal("schema parser must not be nil")
	}
}

func TestTheFeedResult_Methods(t *testing.T) {
	ip := netip.MustParseAddr("2.2.2.2")
	r := TheFeedResult{
		IP:        ip,
		Latency:   42 * time.Millisecond,
		Domain:    "t.example.com",
		QueryMode: "single",
	}

	if r.Key() != "2.2.2.2" {
		t.Errorf("Key() = %q, want 2.2.2.2", r.Key())
	}

	if r.KeyType() != result.KeyIP {
		t.Errorf("KeyType() = %v, want %v", r.KeyType(), result.KeyIP)
	}

	if !r.Equal(TheFeedResult{IP: ip}) {
		t.Error("Equal() = false for same IP, want true")
	}

	if r.Equal(TheFeedResult{IP: netip.MustParseAddr("1.1.1.1")}) {
		t.Error("Equal() = true for different IP, want false")
	}

	want := []string{
		"2.2.2.2",
		result.FormatDuration(42 * time.Millisecond),
		"t.example.com",
		"single",
	}
	got := r.ToRecord()
	if len(got) != len(want) {
		t.Fatalf("ToRecord() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ToRecord()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestTheFeedResult_Score(t *testing.T) {
	// 1000ms -> score 1.0
	if got := (TheFeedResult{Latency: time.Second}).Score(); got != 1.0 {
		t.Errorf("Score for 1s = %v, want 1.0", got)
	}

	// 100ms -> score 10.0
	if got := (TheFeedResult{Latency: 100 * time.Millisecond}).Score(); got != 10.0 {
		t.Errorf("Score for 100ms = %v, want 10.0", got)
	}

	// Sub-1ms values must be clamped to 1ms to avoid division by zero.
	if got := (TheFeedResult{Latency: 0}).Score(); got != 1000.0 {
		t.Errorf("Score for 0 latency = %v, want 1000.0", got)
	}
}

func TestParseTheFeedResult(t *testing.T) {
	record := []string{
		"2.2.2.2",
		result.FormatDuration(42 * time.Millisecond),
		"t.example.com",
		"single",
	}

	res, err := parseTheFeedResult(record)
	if err != nil {
		t.Fatalf("parseTheFeedResult() error = %v", err)
	}

	got, ok := res.(TheFeedResult)
	if !ok {
		t.Fatalf("expected TheFeedResult, got %T", res)
	}

	want := TheFeedResult{
		IP:        netip.MustParseAddr("2.2.2.2"),
		Latency:   42 * time.Millisecond,
		Domain:    "t.example.com",
		QueryMode: "single",
	}

	if got.IP != want.IP {
		t.Errorf("IP = %v, want %v", got.IP, want.IP)
	}
	if got.Latency != want.Latency {
		t.Errorf("Latency = %v, want %v", got.Latency, want.Latency)
	}
	if got.Domain != want.Domain {
		t.Errorf("Domain = %q, want %q", got.Domain, want.Domain)
	}
	if got.QueryMode != want.QueryMode {
		t.Errorf("QueryMode = %q, want %q", got.QueryMode, want.QueryMode)
	}
}

func TestParseTheFeedResult_TooShort(t *testing.T) {
	if _, err := parseTheFeedResult([]string{"2.2.2.2"}); err == nil {
		t.Fatal("expected error for short record")
	}
}

func TestParseTheFeedResult_BadIP(t *testing.T) {
	record := []string{
		"not-an-ip",
		result.FormatDuration(42 * time.Millisecond),
		"t.example.com",
		"single",
	}

	if _, err := parseTheFeedResult(record); err == nil {
		t.Fatal("expected error for invalid IP")
	}
}

func TestParseTheFeedResult_BadLatency(t *testing.T) {
	record := []string{
		"2.2.2.2",
		"not-a-duration",
		"t.example.com",
		"single",
	}

	if _, err := parseTheFeedResult(record); err == nil {
		t.Fatal("expected error for invalid latency")
	}
}

func TestTheFeedResult_RoundTrip(t *testing.T) {
	want := TheFeedResult{
		IP:        netip.MustParseAddr("9.9.9.9"),
		Latency:   123 * time.Millisecond,
		Domain:    "feed.example.com",
		QueryMode: "double",
	}

	parsed, err := parseTheFeedResult(want.ToRecord())
	if err != nil {
		t.Fatalf("parseTheFeedResult() error = %v", err)
	}

	if !want.Equal(parsed) {
		t.Errorf("round-trip mismatch: want %+v, got %+v", want, parsed)
	}
}
