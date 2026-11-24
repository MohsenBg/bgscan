package stormdnsprobe

import (
	"net/netip"
	"reflect"
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

// entryOf wraps an address in the single-address row a probe now receives.
func entryOf(addr netip.Addr) target.Entry {
	return target.NewEnabled(netip.PrefixFrom(addr, addr.BitLen()))
}

// TestStormDNSResult_KeyAndType checks storm dns result key and type.
func TestStormDNSResult_KeyAndType(t *testing.T) {
	r := StormDNSResult{Entry: entryOf(netip.MustParseAddr("1.2.3.4"))}

	if got := r.Key(); got != "1.2.3.4/32:0:" {
		t.Errorf("Key() = %q, want 1.2.3.4/32:0:", got)
	}
	if got := r.KeyType(); got != target.KindIP {
		t.Errorf("KeyType() = %v, want %v", got, target.KindIP)
	}
}

// TestStormDNSResult_Equal checks storm dns result equal.
func TestStormDNSResult_Equal(t *testing.T) {
	base := StormDNSResult{Entry: entryOf(netip.MustParseAddr("8.8.8.8"))}

	if !base.Equal(StormDNSResult{Entry: entryOf(netip.MustParseAddr("8.8.8.8")), Latency: time.Second}) {
		t.Error("same IP should be equal")
	}
	if base.Equal(StormDNSResult{Entry: entryOf(netip.MustParseAddr("1.1.1.1"))}) {
		t.Error("different IP should not be equal")
	}
}

// TestStormDNSResult_ToRecord checks storm dns result to record.
func TestStormDNSResult_ToRecord(t *testing.T) {
	r := StormDNSResult{
		Entry:     entryOf(netip.MustParseAddr("2001:db8::1")),
		Latency:   1500 * time.Millisecond,
		Port:      20110,
		QueryType: "rotate",
		Enc:       dns.EncChaCha20,
	}

	want := []string{"2001:db8::1/128", "0", "", "1", "1.50s", "20110", "ROTATE", "ChaCha20"}

	if got := r.ToRecord(); !reflect.DeepEqual(got, want) {
		t.Errorf("ToRecord() = %#v, want %#v", got, want)
	}
}

// TestStormDNSResult_Score checks storm dns result score.
func TestStormDNSResult_Score(t *testing.T) {
	tests := []struct {
		latency time.Duration
		want    float64
	}{
		{100 * time.Millisecond, 10},
		{0, 1000},
		{-time.Second, 1000},
	}

	for _, tt := range tests {
		got := StormDNSResult{Latency: tt.latency}.Score()
		if got != tt.want {
			t.Errorf("Score(%v) = %v, want %v", tt.latency, got, tt.want)
		}
	}
}

// TestNormalizeQueryType checks normalize query type.
func TestNormalizeQueryType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "TXT"},
		{"  ", "TXT"},
		{"txt", "TXT"},
		{" cname ", "CNAME"},
		{"ROTATE", "ROTATE"},
		{"ns", "NS"},
	}

	for _, tt := range tests {
		if got := normalizeQueryType(tt.input); got != tt.want {
			t.Errorf("normalizeQueryType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// TestParseStormDNSResult checks parse storm dns result.
func TestParseStormDNSResult(t *testing.T) {
	ip := "1.2.3.4/32"

	tests := []struct {
		name    string
		record  []string
		want    StormDNSResult
		wantErr bool
	}{
		{
			name:   "full record",
			record: []string{ip, "0", "", "1", "250ms", "20110", "CNAME", "AES-128-GCM"},
			want: StormDNSResult{
				Entry:     entryOf(netip.MustParseAddr("1.2.3.4")),
				Latency:   250 * time.Millisecond,
				Port:      20110,
				QueryType: "CNAME",
				Enc:       dns.EncAES128GCM,
			},
		},
		{
			name:   "query",
			record: []string{ip, "0", "", "1", "250ms", "1234", "NS", "XOR"},
			want: StormDNSResult{
				Entry:     entryOf(netip.MustParseAddr("1.2.3.4")),
				Latency:   250 * time.Millisecond,
				Port:      1234,
				QueryType: "NS",
				Enc:       dns.EncXOR,
			},
		},
		{
			name:    "legacy record rejected",
			record:  []string{ip, "0", "", "1", "1.50s"},
			wantErr: true,
		},
		{
			name:    "too short",
			record:  []string{ip, "0", "", "1"},
			wantErr: true,
		},
		{
			name:    "bad ip",
			record:  []string{"not-an-ip", "0", "", "1", "250ms", "1234", "TXT", "XOR"},
			wantErr: true,
		},
		{
			name:    "bad latency",
			record:  []string{ip, "0", "", "1", "soon", "1234", "TXT", "XOR"},
			wantErr: true,
		},
		{
			name:   "bad port defaults to zero",
			record: []string{ip, "0", "", "1", "250ms", "port", "TXT", "XOR"},
			want: StormDNSResult{
				Entry:     entryOf(netip.MustParseAddr("1.2.3.4")),
				Latency:   250 * time.Millisecond,
				Port:      0,
				QueryType: "TXT",
				Enc:       dns.EncXOR,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStormDNSResult(tt.record)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseStormDNSResult(%v) = %#v, want error", tt.record, got)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseStormDNSResult(%v) error = %v", tt.record, err)
			}

			if got != result.Result(tt.want) {
				t.Errorf("parseStormDNSResult(%v) = %#v, want %#v", tt.record, got, tt.want)
			}
		})
	}
}

// TestStormDNSResult_ParseRoundTrip checks storm dns result parse round trip.
func TestStormDNSResult_ParseRoundTrip(t *testing.T) {
	original := StormDNSResult{
		Entry:     entryOf(netip.MustParseAddr("5.6.7.8")),
		Latency:   321 * time.Millisecond,
		QueryType: "ROTATE",
		Enc:       dns.EncAES256GCM,
	}

	parsed, err := parseStormDNSResult(original.ToRecord())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if parsed != result.Result(original) {
		t.Errorf("round trip = %#v, want %#v", parsed, original)
	}
}
