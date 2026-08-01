package dnsttprobe

import (
	"net/netip"
	"reflect"
	"testing"
	"time"

	"bgscan/internal/core/dns"
	"bgscan/internal/core/result"
)

func mustAddr(t *testing.T, s string) netip.Addr {
	t.Helper()

	addr, err := netip.ParseAddr(s)
	if err != nil {
		t.Fatalf("ParseAddr(%q): %v", s, err)
	}
	return addr
}

func TestDNSTTResult_Key(t *testing.T) {
	t.Parallel()

	r := DNSTTResult{
		IP: mustAddr(t, "1.2.3.4"),
	}

	got := r.Key()
	want := "1.2.3.4"

	if got != want {
		t.Fatalf("Key() = %q, want %q", got, want)
	}
}

func TestDNSTTResult_KeyType(t *testing.T) {
	t.Parallel()

	r := DNSTTResult{}
	got := r.KeyType()

	if got != result.KeyIP {
		t.Fatalf("KeyType() = %v, want %v", got, result.KeyIP)
	}
}

func TestDNSTTResult_Equal(t *testing.T) {
	t.Parallel()

	base := DNSTTResult{
		IP: mustAddr(t, "8.8.8.8"),
	}

	tests := []struct {
		name  string
		other result.Result
		want  bool
	}{
		{
			name: "same ip",
			other: DNSTTResult{
				IP: mustAddr(t, "8.8.8.8"),
			},
			want: true,
		},
		{
			name: "different ip",
			other: DNSTTResult{
				IP: mustAddr(t, "1.1.1.1"),
			},
			want: false,
		},
		{
			name: "same key different fields",
			other: DNSTTResult{
				IP:        mustAddr(t, "8.8.8.8"),
				Latency:   250 * time.Millisecond,
				Transport: dns.Transport("doh"),
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

func TestDNSTTResult_ToRecord(t *testing.T) {
	t.Parallel()

	r := DNSTTResult{
		IP:        mustAddr(t, "2001:db8::1"),
		Latency:   1500 * time.Millisecond,
		Transport: dns.Transport("udp"),
		Port:      1080,
	}

	got := r.ToRecord()
	want := []string{
		"2001:db8::1",
		"1.5s",
		"udp",
		"1080",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ToRecord() = %#v, want %#v", got, want)
	}
}

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

func TestParseDNSTTResult_BackwardCompatibleOldFormat(t *testing.T) {
	t.Parallel()

	record := []string{
		"1.1.1.1",
		"250ms",
	}

	got, err := parseDNSTTResult(record)
	if err != nil {
		t.Fatalf("parseDNSTTResult() error = %v", err)
	}

	res, ok := got.(DNSTTResult)
	if !ok {
		t.Fatalf("parseDNSTTResult() type = %T, want DNSTTResult", got)
	}

	if res.IP != mustAddr(t, "1.1.1.1") {
		t.Fatalf("IP = %v, want %v", res.IP, mustAddr(t, "1.1.1.1"))
	}
	if res.Latency != 250*time.Millisecond {
		t.Fatalf("Latency = %v, want %v", res.Latency, 250*time.Millisecond)
	}
	if res.Transport != "" {
		t.Fatalf("Transport = %q, want empty", res.Transport)
	}
	if res.Port != 0 {
		t.Fatalf("Port = %d, want 0", res.Port)
	}
}

func TestParseDNSTTResult_NewFormat(t *testing.T) {
	t.Parallel()

	record := []string{
		"8.8.8.8",
		"150ms",
		"udp",
		"1080",
	}

	got, err := parseDNSTTResult(record)
	if err != nil {
		t.Fatalf("parseDNSTTResult() error = %v", err)
	}

	res, ok := got.(DNSTTResult)
	if !ok {
		t.Fatalf("parseDNSTTResult() type = %T, want DNSTTResult", got)
	}

	want := DNSTTResult{
		IP:        mustAddr(t, "8.8.8.8"),
		Latency:   150 * time.Millisecond,
		Transport: dns.ParseTransport("udp"),
		Port:      1080,
	}

	if res != want {
		t.Fatalf("parseDNSTTResult() = %#v, want %#v", res, want)
	}
}

func TestParseDNSTTResult_IPv6(t *testing.T) {
	t.Parallel()

	record := []string{
		"2001:db8::5",
		"2s",
		"doh",
		"9050",
	}

	got, err := parseDNSTTResult(record)
	if err != nil {
		t.Fatalf("parseDNSTTResult() error = %v", err)
	}

	res, ok := got.(DNSTTResult)
	if !ok {
		t.Fatalf("parseDNSTTResult() type = %T, want DNSTTResult", got)
	}

	if res.IP != mustAddr(t, "2001:db8::5") {
		t.Fatalf("IP = %v, want %v", res.IP, mustAddr(t, "2001:db8::5"))
	}
	if res.Latency != 2*time.Second {
		t.Fatalf("Latency = %v, want %v", res.Latency, 2*time.Second)
	}
	if res.Transport != dns.ParseTransport("doh") {
		t.Fatalf("Transport = %q, want %q", res.Transport, dns.ParseTransport("doh"))
	}
	if res.Port != 9050 {
		t.Fatalf("Port = %d, want %d", res.Port, 9050)
	}
}

func TestParseDNSTTResult_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		record []string
	}{
		{
			name:   "too few fields",
			record: []string{"1.1.1.1"},
		},
		{
			name:   "invalid ip",
			record: []string{"not-an-ip", "100ms"},
		},
		{
			name:   "invalid latency",
			record: []string{"1.1.1.1", "not-a-duration"},
		},
		{
			name:   "invalid port",
			record: []string{"1.1.1.1", "100ms", "udp", "not-a-port"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseDNSTTResult(tt.record)
			if err == nil {
				t.Fatalf("expected error, got result = %#v", got)
			}
		})
	}
}

func TestSchema_Metadata(t *testing.T) {
	t.Parallel()

	if Schema.Name != "DNSTT" {
		t.Fatalf("Schema.Name = %q, want %q", Schema.Name, "DNSTT")
	}

	if Schema.Directory != "dnstt" {
		t.Fatalf("Schema.Directory = %q, want %q", Schema.Directory, "dnstt")
	}

	if len(Schema.Columns) != 4 {
		t.Fatalf("len(Schema.Columns) = %d, want 4", len(Schema.Columns))
	}

	wantNames := []string{"IP", "Latency", "Transport", "Port"}
	for i, want := range wantNames {
		if Schema.Columns[i].Name != want {
			t.Fatalf("Schema.Columns[%d].Name = %q, want %q", i, Schema.Columns[i].Name, want)
		}
	}

	if Schema.Parser == nil {
		t.Fatal("Schema.Parser is nil")
	}
}

func TestSchema_Parser(t *testing.T) {
	t.Parallel()

	record := []string{"9.9.9.9", "300ms", "dot", "1081"}

	got, err := Schema.Parser(record)
	if err != nil {
		t.Fatalf("Schema.Parser() error = %v", err)
	}

	res, ok := got.(DNSTTResult)
	if !ok {
		t.Fatalf("Schema.Parser() type = %T, want DNSTTResult", got)
	}

	if res.IP != mustAddr(t, "9.9.9.9") {
		t.Fatalf("IP = %v, want %v", res.IP, mustAddr(t, "9.9.9.9"))
	}
	if res.Latency != 300*time.Millisecond {
		t.Fatalf("Latency = %v, want %v", res.Latency, 300*time.Millisecond)
	}
	if res.Port != 1081 {
		t.Fatalf("Port = %d, want %d", res.Port, 1081)
	}
}
