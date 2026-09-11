package stormdnsprobe

import (
	"net/netip"
	"reflect"
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/result"
)

func TestStormDNSResult_KeyAndType(t *testing.T) {
	r := StormDNSResult{IP: netip.MustParseAddr("1.2.3.4")}

	if got := r.Key(); got != "1.2.3.4" {
		t.Errorf("Key() = %q, want 1.2.3.4", got)
	}
	if got := r.KeyType(); got != result.KeyIP {
		t.Errorf("KeyType() = %v, want %v", got, result.KeyIP)
	}
}

func TestStormDNSResult_Equal(t *testing.T) {
	base := StormDNSResult{IP: netip.MustParseAddr("8.8.8.8")}

	if !base.Equal(StormDNSResult{IP: netip.MustParseAddr("8.8.8.8"), Latency: time.Second}) {
		t.Error("same IP should be equal")
	}
	if base.Equal(StormDNSResult{IP: netip.MustParseAddr("1.1.1.1")}) {
		t.Error("different IP should not be equal")
	}
}

func TestStormDNSResult_ToRecord(t *testing.T) {
	r := StormDNSResult{
		IP:        netip.MustParseAddr("2001:db8::1"),
		Latency:   1500 * time.Millisecond,
		Port:      20110,
		QueryType: "rotate",
		Enc:       dns.EncChaCha20,
	}

	want := []string{"2001:db8::1", "1.50s", "20110", "ROTATE", "ChaCha20"}

	if got := r.ToRecord(); !reflect.DeepEqual(got, want) {
		t.Errorf("ToRecord() = %#v, want %#v", got, want)
	}
}

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

func TestParseStormDNSResult(t *testing.T) {
	ip := "1.2.3.4"

	tests := []struct {
		name    string
		record  []string
		want    StormDNSResult
		wantErr bool
	}{
		{
			name:   "full record",
			record: []string{ip, "250ms", "20110", "CNAME", "AES-128-GCM"},
			want: StormDNSResult{
				IP:        netip.MustParseAddr(ip),
				Latency:   250 * time.Millisecond,
				Port:      20110,
				QueryType: "CNAME",
				Enc:       dns.EncAES128GCM,
			},
		},
		{
			name:   "port and query",
			record: []string{ip, "250ms", "1234", "NS"},
			want: StormDNSResult{
				IP:        netip.MustParseAddr(ip),
				Latency:   250 * time.Millisecond,
				Port:      1234,
				QueryType: "NS",
				Enc:       dns.EncXOR,
			},
		},
		{
			name:   "port only",
			record: []string{ip, "250ms", "1234"},
			want: StormDNSResult{
				IP:        netip.MustParseAddr(ip),
				Latency:   250 * time.Millisecond,
				Port:      1234,
				QueryType: "TXT",
				Enc:       dns.EncXOR,
			},
		},
		{
			name:   "legacy record",
			record: []string{ip, "1.50s"},
			want: StormDNSResult{
				IP:        netip.MustParseAddr(ip),
				Latency:   1500 * time.Millisecond,
				QueryType: "TXT",
				Enc:       dns.EncXOR,
			},
		},
		{
			name:    "too short",
			record:  []string{ip},
			wantErr: true,
		},
		{
			name:    "bad ip",
			record:  []string{"not-an-ip", "250ms", "1234", "TXT", "XOR"},
			wantErr: true,
		},
		{
			name:    "bad latency",
			record:  []string{ip, "soon", "1234", "TXT", "XOR"},
			wantErr: true,
		},
		{
			name:    "bad port",
			record:  []string{ip, "250ms", "port"},
			wantErr: true,
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

func TestStormDNSResult_ParseRoundTrip(t *testing.T) {
	original := StormDNSResult{
		IP:        netip.MustParseAddr("5.6.7.8"),
		Latency:   321 * time.Millisecond,
		Port:      20200,
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
