package dnsttprobe

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

func mustAddr(t *testing.T, s string) netip.Addr {
	t.Helper()

	addr, err := netip.ParseAddr(s)
	if err != nil {
		t.Fatalf("ParseAddr(%q): %v", s, err)
	}
	return addr
}

// TestDNSTTResult_Key covers DNSTT result key.
func TestDNSTTResult_Key(t *testing.T) {
	t.Parallel()

	r := DNSTTResult{
		Entry: entryOf(mustAddr(t, "1.2.3.4")),
	}

	got := r.Key()
	want := "0:1.2.3.4/32:0:"

	if got != want {
		t.Fatalf("Key() = %q, want %q", got, want)
	}
}

// TestDNSTTResult_KeyType covers DNSTT result key type.
func TestDNSTTResult_KeyType(t *testing.T) {
	t.Parallel()

	r := DNSTTResult{}
	got := r.KeyType()

	if got != target.KindIP {
		t.Fatalf("KeyType() = %v, want %v", got, target.KindIP)
	}
}

// TestDNSTTResult_Equal covers DNSTT result equal.
func TestDNSTTResult_Equal(t *testing.T) {
	t.Parallel()

	base := DNSTTResult{
		Entry: entryOf(mustAddr(t, "8.8.8.8")),
	}

	tests := []struct {
		name  string
		other result.Result
		want  bool
	}{
		{
			name: "same ip",
			other: DNSTTResult{
				Entry: entryOf(mustAddr(t, "8.8.8.8")),
			},
			want: true,
		},
		{
			name: "different ip",
			other: DNSTTResult{
				Entry: entryOf(mustAddr(t, "1.1.1.1")),
			},
			want: false,
		},
		{
			name: "same key different fields",
			other: DNSTTResult{
				Entry:     entryOf(mustAddr(t, "8.8.8.8")),
				Latency:   250 * time.Millisecond,
				Transport: dns.ResolverType("doh"),
				Port:      1080,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := base.Equal(tt.other)
			if got != tt.want {
				t.Fatalf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDNSTTResult_ToRecord covers DNSTT result to record.
func TestDNSTTResult_ToRecord(t *testing.T) {
	t.Parallel()

	r := DNSTTResult{
		Entry:     entryOf(mustAddr(t, "2001:db8::1")),
		Latency:   1500 * time.Millisecond,
		Transport: dns.ResolverType("udp"),
		Port:      1080,
	}

	got := r.ToRecord()
	want := []string{
		"2001:db8::1/128", "0", "", "1",
		"1.50s",
		"udp",
		"1080",
		"",
		"",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ToRecord() = %#v, want %#v", got, want)
	}
}

// TestDNSTTResult_Score covers DNSTT result score.
func TestDNSTTResult_Score(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		latency time.Duration
		want    float64
	}{
		{
			name:    "1ms",
			latency: 1 * time.Millisecond,
			want:    1000.0,
		},
		{
			name:    "10ms",
			latency: 10 * time.Millisecond,
			want:    100.0,
		},
		{
			name:    "250ms",
			latency: 250 * time.Millisecond,
			want:    4.0,
		},
		{
			name:    "sub-millisecond rounded up to 1ms floor",
			latency: 500 * time.Microsecond,
			want:    1000.0,
		},
		{
			name:    "zero latency uses 1ms floor",
			latency: 0,
			want:    1000.0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := DNSTTResult{Latency: tt.latency}
			got := r.Score()

			if got != tt.want {
				t.Fatalf("Score() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseDNSTTResult_BackwardCompatibleOldFormat covers parsing the DNSTT result backward
// compatible old format.
func TestParseDNSTTResult_BackwardCompatibleOldFormat(t *testing.T) {
	t.Parallel()

	// Old format is no longer supported; this test verifies it's rejected.
	record := []string{
		"1.1.1.1",
		"250ms",
	}

	_, err := parseDNSTTResult(record)
	if err == nil {
		t.Fatal("expected error for legacy two-field record")
	}
}

// TestParseDNSTTResult_NewFormat covers parsing the DNSTT result new format.
func TestParseDNSTTResult_NewFormat(t *testing.T) {
	t.Parallel()

	// New format: entry fields (value, port, path, enable) + result fields (latency, transport, port, auth, proxy)
	record := []string{
		"8.8.8.8/32", "0", "", "1",
		"150ms",
		"udp",
		"1080",
		"none",
		"socks",
	}

	got, err := parseDNSTTResult(record)
	if err != nil {
		t.Fatalf("parseDNSTTResult() error = %v", err)
	}

	res, ok := got.(DNSTTResult)
	if !ok {
		t.Fatalf("parseDNSTTResult() type = %T, want DNSTTResult", got)
	}

	if res.Entry.IP.Addr() != mustAddr(t, "8.8.8.8") {
		t.Fatalf("IP = %v, want %v", res.Entry.IP.Addr(), mustAddr(t, "8.8.8.8"))
	}
	if res.Latency != 150*time.Millisecond {
		t.Fatalf("Latency = %v, want %v", res.Latency, 150*time.Millisecond)
	}
	if res.Transport != dns.ParseResolverType("udp") {
		t.Fatalf("Transport = %q, want %q", res.Transport, dns.ParseResolverType("udp"))
	}
	if res.Port != 1080 {
		t.Fatalf("Port = %d, want %d", res.Port, 1080)
	}
	if res.AuthMethod != dns.AuthNone {
		t.Fatalf("AuthMethod = %q, want %q", res.AuthMethod, dns.AuthNone)
	}
	if res.ResolverProxyType != dns.ResolverProxySOCKS {
		t.Fatalf("ResolverProxyType = %q, want %q", res.ResolverProxyType, dns.ResolverProxySOCKS)
	}
}

// TestParseDNSTTResult_IPv6 covers parsing the DNSTT result I pv6.
func TestParseDNSTTResult_IPv6(t *testing.T) {
	t.Parallel()

	record := []string{
		"2001:db8::5/128", "0", "", "1",
		"2.00s",
		"doh",
		"9050",
		"none",
		"socks",
	}

	got, err := parseDNSTTResult(record)
	if err != nil {
		t.Fatalf("parseDNSTTResult() error = %v", err)
	}

	res, ok := got.(DNSTTResult)
	if !ok {
		t.Fatalf("parseDNSTTResult() type = %T, want DNSTTResult", got)
	}

	if res.Entry.IP.Addr() != mustAddr(t, "2001:db8::5") {
		t.Fatalf("IP = %v, want %v", res.Entry.IP.Addr(), mustAddr(t, "2001:db8::5"))
	}
	if res.Latency != 2*time.Second {
		t.Fatalf("Latency = %v, want %v", res.Latency, 2*time.Second)
	}
	if res.Transport != dns.ParseResolverType("doh") {
		t.Fatalf("Transport = %q, want %q", res.Transport, dns.ParseResolverType("doh"))
	}
	if res.Port != 9050 {
		t.Fatalf("Port = %d, want %d", res.Port, 9050)
	}
}

// TestParseDNSTTResult_Errors covers parsing the DNSTT result errors.
func TestParseDNSTTResult_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		record      []string
		wantError   bool
		checkResult func(t *testing.T, res DNSTTResult)
	}{
		{
			name:      "too few fields",
			record:    []string{"1.1.1.1/32", "0", "", "1"},
			wantError: true,
		},
		{
			name:      "invalid ip",
			record:    []string{"not-an-ip", "0", "", "1", "100ms", "udp", "1080", "none", "socks"},
			wantError: true,
		},
		{
			name:      "invalid latency",
			record:    []string{"1.1.1.1/32", "0", "", "1", "not-a-duration", "udp", "1080", "none", "socks"},
			wantError: true,
		},
		{
			name:      "invalid port defaults to zero",
			record:    []string{"1.1.1.1/32", "0", "", "1", "100ms", "udp", "not-a-port", "none", "socks"},
			wantError: false,
			checkResult: func(t *testing.T, res DNSTTResult) {
				if res.Port != 0 {
					t.Fatalf("Port = %d, want 0 for invalid port", res.Port)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseDNSTTResult(tt.record)
			if tt.wantError {
				if err == nil {
					t.Fatalf("expected error, got result = %#v", got)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				res, ok := got.(DNSTTResult)
				if !ok {
					t.Fatalf("parseDNSTTResult() type = %T, want DNSTTResult", got)
				}
				if tt.checkResult != nil {
					tt.checkResult(t, res)
				}
			}
		})
	}
}

// TestSchema_Metadata covers schema metadata.
func TestSchema_Metadata(t *testing.T) {
	t.Parallel()

	if Schema.Name != "DNSTT" {
		t.Fatalf("Schema.Name = %q, want %q", Schema.Name, "DNSTT")
	}

	if Schema.Directory != "dnstt" {
		t.Fatalf("Schema.Directory = %q, want %q", Schema.Directory, "dnstt")
	}

	if len(Schema.Columns) != 6 {
		t.Fatalf("len(Schema.Columns) = %d, want 6", len(Schema.Columns))
	}

	wantNames := []string{"Target", "Latency", "Transport", "Port", "Auth", "Proxy"}
	for i, want := range wantNames {
		if Schema.Columns[i].Name != want {
			t.Fatalf("Schema.Columns[%d].Name = %q, want %q", i, Schema.Columns[i].Name, want)
		}
	}

	if Schema.Parser == nil {
		t.Fatal("Schema.Parser is nil")
	}
}

// TestSchema_Parser covers schema parser.
func TestSchema_Parser(t *testing.T) {
	t.Parallel()

	record := []string{"9.9.9.9/32", "0", "", "1", "300ms", "dot", "1081", "none", "socks"}

	got, err := Schema.Parser(record)
	if err != nil {
		t.Fatalf("Schema.Parser() error = %v", err)
	}

	res, ok := got.(DNSTTResult)
	if !ok {
		t.Fatalf("Schema.Parser() type = %T, want DNSTTResult", got)
	}

	if res.Entry.IP.Addr() != mustAddr(t, "9.9.9.9") {
		t.Fatalf("IP = %v, want %v", res.Entry.IP.Addr(), mustAddr(t, "9.9.9.9"))
	}
	if res.Latency != 300*time.Millisecond {
		t.Fatalf("Latency = %v, want %v", res.Latency, 300*time.Millisecond)
	}
	if res.Port != 1081 {
		t.Fatalf("Port = %d, want %d", res.Port, 1081)
	}
}
