package vaydnsprobe

import (
	"net/netip"
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/result"
)

func TestVayDNSResult_ToRecord(t *testing.T) {
	r := VayDNSResult{
		IP:                netip.MustParseAddr("1.2.3.4"),
		Latency:           125 * time.Millisecond,
		Transport:         dns.ResolverTypeUDP,
		Port:              1080,
		AuthMethod:        dns.AuthNone,
		ResolverProxyType: dns.ResolverProxySOCKS,
	}

	got := r.ToRecord()

	want := []string{
		"1.2.3.4",
		"125.00ms",
		string(dns.ResolverTypeUDP),
		"1080",
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

func TestVayDNSResult_Key(t *testing.T) {
	r := VayDNSResult{
		IP: netip.MustParseAddr("2001:db8::1"),
	}

	if got := r.Key(); got != "2001:db8::1" {
		t.Errorf("Key() = %q, want %q", got, "2001:db8::1")
	}
}

func TestVayDNSResult_KeyType(t *testing.T) {
	r := VayDNSResult{}

	if got := r.KeyType(); got != result.KeyIP {
		t.Errorf("KeyType() = %v, want %v", got, result.KeyIP)
	}
}

func TestVayDNSResult_Equal(t *testing.T) {
	r := VayDNSResult{
		IP: netip.MustParseAddr("1.2.3.4"),
	}

	tests := []struct {
		name string
		rs   result.Result
		want bool
	}{
		{
			name: "same IP",
			rs: VayDNSResult{
				IP: netip.MustParseAddr("1.2.3.4"),
			},
			want: true,
		},
		{
			name: "different IP",
			rs: VayDNSResult{
				IP: netip.MustParseAddr("5.6.7.8"),
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
				"1.2.3.4",
				"125ms",
				string(dns.ResolverTypeUDP),
				"1080",
			},
			want: VayDNSResult{
				IP:        netip.MustParseAddr("1.2.3.4"),
				Latency:   125 * time.Millisecond,
				Transport: dns.ResolverTypeUDP,
				Port:      1080,
			},
		},
		{
			name: "legacy record",
			record: []string{
				"1.2.3.4",
				"125ms",
			},
			want: VayDNSResult{
				IP:      netip.MustParseAddr("1.2.3.4"),
				Latency: 125 * time.Millisecond,
			},
		},
		{
			name: "IPv6",
			record: []string{
				"2001:db8::1",
				"250ms",
				string(dns.ResolverTypeUDP),
				"5353",
			},
			want: VayDNSResult{
				IP:        netip.MustParseAddr("2001:db8::1"),
				Latency:   250 * time.Millisecond,
				Transport: dns.ResolverTypeUDP,
				Port:      5353,
			},
		},
		{
			name:      "missing fields",
			record:    []string{"1.2.3.4"},
			wantError: true,
		},
		{
			name: "invalid IP",
			record: []string{
				"not-an-ip",
				"100ms",
			},
			wantError: true,
		},
		{
			name: "invalid latency",
			record: []string{
				"1.2.3.4",
				"not-a-duration",
			},
			wantError: true,
		},
		{
			name: "invalid port",
			record: []string{
				"1.2.3.4",
				"100ms",
				string(dns.ResolverTypeUDP),
				"invalid",
			},
			wantError: true,
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

			if result.IP != tt.want.IP {
				t.Errorf("IP = %v, want %v", result.IP, tt.want.IP)
			}

			if result.Latency != tt.want.Latency {
				t.Errorf("Latency = %v, want %v", result.Latency, tt.want.Latency)
			}

			if result.Transport != tt.want.Transport {
				t.Errorf("Transport = %v, want %v", result.Transport, tt.want.Transport)
			}

			if result.Port != tt.want.Port {
				t.Errorf("Port = %d, want %d", result.Port, tt.want.Port)
			}
		})
	}
}
