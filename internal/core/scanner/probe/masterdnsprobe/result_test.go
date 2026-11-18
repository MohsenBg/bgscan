package masterdnsprobe

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

func TestMasterDNSResult_KeyAndType(t *testing.T) {
	r := MasterDNSResult{Entry: entryOf(netip.MustParseAddr("1.2.3.4"))}

	if got := r.Key(); got != "1.2.3.4/32:0:" {
		t.Errorf("Key() = %q, want 1.2.3.4/32:0:", got)
	}
	if got := r.KeyType(); got != target.KindIP {
		t.Errorf("KeyType() = %v, want %v", got, target.KindIP)
	}
}

func TestMasterDNSResult_Equal(t *testing.T) {
	base := MasterDNSResult{Entry: entryOf(netip.MustParseAddr("8.8.8.8"))}

	if !base.Equal(MasterDNSResult{Entry: entryOf(netip.MustParseAddr("8.8.8.8")), Latency: time.Second}) {
		t.Error("same IP should be equal")
	}
	if base.Equal(MasterDNSResult{Entry: entryOf(netip.MustParseAddr("1.1.1.1"))}) {
		t.Error("different IP should not be equal")
	}
}

func TestMasterDNSResult_ToRecord(t *testing.T) {
	r := MasterDNSResult{
		Entry:   entryOf(netip.MustParseAddr("2001:db8::1")),
		Latency: 1500 * time.Millisecond,
		Port:    20010,
		Enc:     dns.EncChaCha20,
	}

	want := []string{"2001:db8::1/128", "0", "", "1", "1.50s", "20010", "ChaCha20"}

	if got := r.ToRecord(); !reflect.DeepEqual(got, want) {
		t.Errorf("ToRecord() = %#v, want %#v", got, want)
	}
}

func TestMasterDNSResult_Score(t *testing.T) {
	tests := []struct {
		latency time.Duration
		want    float64
	}{
		{100 * time.Millisecond, 10},
		{0, 1000},
		{-time.Second, 1000},
	}

	for _, tt := range tests {
		got := MasterDNSResult{Latency: tt.latency}.Score()
		if got != tt.want {
			t.Errorf("Score(%v) = %v, want %v", tt.latency, got, tt.want)
		}
	}
}

func TestParseMasterDNSResult(t *testing.T) {
	ip := "1.2.3.4/32"

	tests := []struct {
		name    string
		record  []string
		want    MasterDNSResult
		wantErr bool
	}{
		{
			name:   "full record",
			record: []string{ip, "0", "", "1", "250ms", "20010", "AES-128-GCM"},
			want: MasterDNSResult{
				Entry:   entryOf(netip.MustParseAddr("1.2.3.4")),
				Latency: 250 * time.Millisecond,
				Port:    20010,
				Enc:     dns.EncAES128GCM,
			},
		},
		{
			name:   "port only",
			record: []string{ip, "0", "", "1", "250ms", "1234", "XOR"},
			want: MasterDNSResult{
				Entry:   entryOf(netip.MustParseAddr("1.2.3.4")),
				Latency: 250 * time.Millisecond,
				Port:    1234,
				Enc:     dns.EncXOR,
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
			record:  []string{"not-an-ip", "0", "", "1", "250ms", "1234", "XOR"},
			wantErr: true,
		},
		{
			name:    "bad latency",
			record:  []string{ip, "0", "", "1", "soon", "1234", "XOR"},
			wantErr: true,
		},
		{
			name:    "bad port",
			record:  []string{ip, "0", "", "1", "250ms", "port", "XOR"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMasterDNSResult(tt.record)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseMasterDNSResult(%v) = %#v, want error", tt.record, got)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseMasterDNSResult(%v) error = %v", tt.record, err)
			}

			if got != result.Result(tt.want) {
				t.Errorf("parseMasterDNSResult(%v) = %#v, want %#v", tt.record, got, tt.want)
			}
		})
	}
}

func TestMasterDNSResult_ParseRoundTrip(t *testing.T) {
	original := MasterDNSResult{
		Entry:   entryOf(netip.MustParseAddr("5.6.7.8")),
		Latency: 321 * time.Millisecond,
		Port:    20100,
		Enc:     dns.EncAES256GCM,
	}

	parsed, err := parseMasterDNSResult(original.ToRecord())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if parsed != result.Result(original) {
		t.Errorf("round trip = %#v, want %#v", parsed, original)
	}
}
