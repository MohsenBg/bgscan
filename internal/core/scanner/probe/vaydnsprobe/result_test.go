package vaydnsprobe

import (
	"net/netip"
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

// mustEntry wraps an address in the single-address row a probe now receives.
func mustEntry(t *testing.T, s string) target.Entry {
	t.Helper()

	addr := netip.MustParseAddr(s)

	return target.NewEnabled(netip.PrefixFrom(addr, addr.BitLen()))
}

// TestVayDNSResult_ToRecord covers the vay dnsresult to record scenario.
func TestVayDNSResult_ToRecord(t *testing.T) {
	r := VayDNSResult{
		Entry:             mustEntry(t, "1.2.3.4"),
		Latency:           125 * time.Millisecond,
		Port:              1080,
		Transport:         dns.ResolverTypeUDP,
		AuthMethod:        dns.AuthNone,
		ResolverProxyType: dns.ResolverProxySOCKS,
	}

	got := r.ToRecord()

	want := []string{
		"1.2.3.4/32",
		"0",
		"",
		"1",
		"125.00ms",
		"1080",
		string(dns.ResolverTypeUDP),
		"none",
		"socks",
	}

	if len(got) != len(want) {
		t.Fatalf("ToRecord() returned %d fields, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ToRecord()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestVayDNSResult_Score covers the vay dnsresult score scenario.
func TestVayDNSResult_Score(t *testing.T) {
	tests := []struct {
		name    string
		latency time.Duration
		want    float64
	}{
		{
			name:    "1 millisecond",
			latency: 1 * time.Millisecond,
			want:    1000,
		},
		{
			name:    "10 milliseconds",
			latency: 10 * time.Millisecond,
			want:    100,
		},
		{
			name:    "100 milliseconds",
			latency: 100 * time.Millisecond,
			want:    10,
		},
		{
			name:    "1 second",
			latency: 1 * time.Second,
			want:    1,
		},
		{
			name:    "zero latency is clamped",
			latency: 0,
			want:    1000,
		},
		{
			name:    "negative latency is clamped",
			latency: -100 * time.Millisecond,
			want:    1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := VayDNSResult{Latency: tt.latency}

			if got := r.Score(); got != tt.want {
				t.Errorf("Score() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestVayDNSResult_Key covers the vay dnsresult key scenario.
func TestVayDNSResult_Key(t *testing.T) {
	entry := mustEntry(t, "2001:db8::1")
	r := VayDNSResult{
		Entry: entry,
	}

	if got := r.Key(); got != entry.Key() {
		t.Errorf("Key() = %q, want %q", got, entry.Key())
	}
}

// TestVayDNSResult_KeyType covers the vay dnsresult key type scenario.
func TestVayDNSResult_KeyType(t *testing.T) {
	r := VayDNSResult{}

	if got := r.KeyType(); got != target.KindIP {
		t.Errorf("KeyType() = %v, want %v", got, target.KindIP)
	}
}

// TestVayDNSResult_Equal covers the vay dnsresult equal scenario.
func TestVayDNSResult_Equal(t *testing.T) {
	r := VayDNSResult{
		Entry: mustEntry(t, "1.2.3.4"),
	}

	tests := []struct {
		name string
		rs   result.Result
		want bool
	}{
		{
			name: "same IP",
			rs: VayDNSResult{
				Entry: mustEntry(t, "1.2.3.4"),
			},
			want: true,
		},
		{
			name: "different IP",
			rs: VayDNSResult{
				Entry: mustEntry(t, "5.6.7.8"),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := r.Equal(tt.rs); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseVayDNSResult covers the parse vay dnsresult scenario.
func TestParseVayDNSResult(t *testing.T) {
	tests := []struct {
		name      string
		record    []string
		want      VayDNSResult
		wantError bool
	}{
		{
			name: "full record",
			record: []string{
				"1.2.3.4/32", "0", "", "1",
				"125ms",
				"1080",
				string(dns.ResolverTypeUDP),
				string(dns.AuthNone),
				string(dns.ResolverProxySOCKS),
			},
			want: VayDNSResult{
				Entry:             mustEntry(t, "1.2.3.4"),
				Latency:           125 * time.Millisecond,
				Port:              1080,
				Transport:         dns.ResolverTypeUDP,
				AuthMethod:        dns.AuthNone,
				ResolverProxyType: dns.ResolverProxySOCKS,
			},
		},
		{
			name: "record with port and path",
			record: []string{
				"192.0.2.10/32", "8443", "/probe", "1",
				"125ms",
				"1080",
				string(dns.ResolverTypeUDP),
				string(dns.AuthPassword),
				string(dns.ResolverProxySSH),
			},
			want: VayDNSResult{
				Entry: target.Entry{
					Kind:   target.KindIP,
					IP:     netip.MustParsePrefix("192.0.2.10/32"),
					Port:   8443,
					Path:   "/probe",
					Enable: true,
				},
				Latency:           125 * time.Millisecond,
				Port:              1080,
				Transport:         dns.ResolverTypeUDP,
				AuthMethod:        dns.AuthPassword,
				ResolverProxyType: dns.ResolverProxySSH,
			},
		},
		{
			name: "IPv6",
			record: []string{
				"2001:db8::1/128", "0", "", "1",
				"250ms",
				"5353",
				string(dns.ResolverTypeUDP),
				string(dns.AuthNone),
				string(dns.ResolverProxySOCKS),
			},
			want: VayDNSResult{
				Entry:             mustEntry(t, "2001:db8::1"),
				Latency:           250 * time.Millisecond,
				Port:              5353,
				Transport:         dns.ResolverTypeUDP,
				AuthMethod:        dns.AuthNone,
				ResolverProxyType: dns.ResolverProxySOCKS,
			},
		},
		{
			name:      "missing fields",
			record:    []string{"1.2.3.4/32", "0", "", "1"},
			wantError: true,
		},
		{
			name: "legacy record rejected",
			record: []string{
				"1.2.3.4/32", "0", "", "1",
				"125ms",
			},
			wantError: true,
		},
		{
			name: "invalid IP",
			record: []string{
				"not-an-ip", "0", "", "1",
				"100ms",
				"1080",
				string(dns.ResolverTypeUDP),
				string(dns.AuthNone),
				string(dns.ResolverProxySOCKS),
			},
			wantError: true,
		},
		{
			name: "invalid latency",
			record: []string{
				"1.2.3.4/32", "0", "", "1",
				"not-a-duration",
				"1080",
				string(dns.ResolverTypeUDP),
				string(dns.AuthNone),
				string(dns.ResolverProxySOCKS),
			},
			wantError: true,
		},
		{
			name: "invalid port defaults to zero",
			record: []string{
				"1.2.3.4/32", "0", "", "1",
				"100ms",
				"invalid",
				string(dns.ResolverTypeUDP),
				string(dns.AuthNone),
				string(dns.ResolverProxySOCKS),
			},
			want: VayDNSResult{
				Entry:             mustEntry(t, "1.2.3.4"),
				Latency:           100 * time.Millisecond,
				Port:              0,
				Transport:         dns.ResolverTypeUDP,
				AuthMethod:        dns.AuthNone,
				ResolverProxyType: dns.ResolverProxySOCKS,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVayDNSResult(tt.record)

			if tt.wantError {
				if err == nil {
					t.Fatal("parseVayDNSResult() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseVayDNSResult() unexpected error: %v", err)
			}

			result, ok := got.(VayDNSResult)
			if !ok {
				t.Fatalf("parseVayDNSResult() returned %T, want VayDNSResult", got)
			}

			if result.Entry.Key() != tt.want.Entry.Key() {
				t.Errorf("Entry = %v, want %v", result.Entry, tt.want.Entry)
			}

			if result.Latency != tt.want.Latency {
				t.Errorf("Latency = %v, want %v", result.Latency, tt.want.Latency)
			}

			if result.Transport != tt.want.Transport {
				t.Errorf("Transport = %v, want %v", result.Transport, tt.want.Transport)
			}

		})
	}
}
