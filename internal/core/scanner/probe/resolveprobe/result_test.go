package resolveprobe

import (
	"net/netip"
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
)

// mustAddr parses an IP address string, failing the test if invalid.
func mustAddr(t *testing.T, s string) netip.Addr {
	t.Helper()

	ip, err := netip.ParseAddr(s)
	if err != nil {
		t.Fatalf("ParseAddr(%q) error = %v", s, err)
	}

	return ip
}

func TestSchema(t *testing.T) {
	t.Parallel()

	if Schema.Name != "DNSResolver" {
		t.Fatalf("Schema.Name = %q, want %q", Schema.Name, "DNSResolver")
	}

	if Schema.Directory != "dns_resolver" {
		t.Fatalf("Schema.Directory = %q, want %q", Schema.Directory, "dns_resolver")
	}

	if len(Schema.Columns) != 6 {
		t.Fatalf("len(Schema.Columns) = %d, want %d", len(Schema.Columns), 6)
	}

	wantCols := []string{
		"IP",
		"Latency",
		"Record Type",
		"Tries",
		"Rcode",
		"DPI Check",
	}

	for i, want := range wantCols {
		if Schema.Columns[i].Name != want {
			t.Fatalf("Schema.Columns[%d].Name = %q, want %q", i, Schema.Columns[i].Name, want)
		}
	}

	if Schema.Parser == nil {
		t.Fatal("Schema.Parser is nil")
	}
}

func TestResolverResultKey(t *testing.T) {
	t.Parallel()

	r := ResolverResult{
		IP: mustAddr(t, "192.0.2.10"),
	}

	if got := r.Key(); got != "192.0.2.10" {
		t.Fatalf("Key() = %q, want %q", got, "192.0.2.10")
	}
}

func TestResolverResultKeyType(t *testing.T) {
	t.Parallel()

	r := ResolverResult{}

	if got := r.KeyType(); got != result.KeyIP {
		t.Fatalf("KeyType() = %v, want %v", got, result.KeyIP)
	}
}

func TestResolverResultEqual(t *testing.T) {
	t.Parallel()

	r1 := ResolverResult{IP: mustAddr(t, "198.51.100.7")}
	r2 := ResolverResult{IP: mustAddr(t, "198.51.100.7")}
	r3 := ResolverResult{IP: mustAddr(t, "198.51.100.8")}

	if !r1.Equal(r2) {
		t.Fatal("Equal() = false, want true for same IP")
	}

	if r1.Equal(r3) {
		t.Fatal("Equal() = true, want false for different IP")
	}
}

func TestResolverResultToRecord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   ResolverResult
		want []string
	}{
		{
			name: "dpi passed",
			in: ResolverResult{
				IP:         mustAddr(t, "203.0.113.1"),
				Latency:    25 * time.Millisecond,
				RecordType: "A",
				Tries:      3,
				Rcode:      0,
				DPIChecked: true,
			},
			want: []string{"203.0.113.1", "25.00ms", "A", "3", "0", "passed"},
		},
		{
			name: "dpi skipped",
			in: ResolverResult{
				IP:         mustAddr(t, "203.0.113.2"),
				Latency:    40 * time.Millisecond,
				RecordType: "AAAA",
				Tries:      1,
				Rcode:      3,
				DPIChecked: false,
			},
			want: []string{"203.0.113.2", "40.00ms", "AAAA", "1", "3", "skipped"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.in.ToRecord()
			if len(got) != len(tt.want) {
				t.Fatalf("len(ToRecord()) = %d, want %d", len(got), len(tt.want))
			}

			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("ToRecord()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestResolverResultScore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   ResolverResult
		want float64
	}{
		{
			name: "1ms",
			in: ResolverResult{
				Latency: 1 * time.Millisecond,
			},
			want: 1000,
		},
		{
			name: "10ms",
			in: ResolverResult{
				Latency: 10 * time.Millisecond,
			},
			want: 100,
		},
		{
			name: "sub-millisecond clamped to 1ms",
			in: ResolverResult{
				Latency: 500 * time.Microsecond,
			},
			want: 1000,
		},
		{
			name: "zero clamped to 1ms",
			in: ResolverResult{
				Latency: 0,
			},
			want: 1000,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.in.Score()
			if got != tt.want {
				t.Fatalf("Score() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseResolverResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		record    []string
		want      ResolverResult
		wantError bool
	}{
		{
			name:   "valid full record dpi passed",
			record: []string{"192.0.2.1", "25ms", "A", "3", "0", "passed"},
			want: ResolverResult{
				IP:         mustAddr(t, "192.0.2.1"),
				Latency:    25 * time.Millisecond,
				RecordType: "A",
				Tries:      3,
				Rcode:      0,
				DPIChecked: true,
			},
		},
		{
			name:   "valid full record dpi skipped",
			record: []string{"192.0.2.2", "40ms", "AAAA", "2", "3", "skipped"},
			want: ResolverResult{
				IP:         mustAddr(t, "192.0.2.2"),
				Latency:    40 * time.Millisecond,
				RecordType: "AAAA",
				Tries:      2,
				Rcode:      3,
				DPIChecked: false,
			},
		},
		{
			name:   "legacy record",
			record: []string{"192.0.2.3", "12ms"},
			want: ResolverResult{
				IP:         mustAddr(t, "192.0.2.3"),
				Latency:    12 * time.Millisecond,
				RecordType: "?",
				Tries:      1,
				Rcode:      0,
				DPIChecked: false,
			},
		},
		{
			name:   "latency rounded to nearest millisecond",
			record: []string{"192.0.2.4", "1500us", "A", "1", "0", "passed"},
			want: ResolverResult{
				IP:         mustAddr(t, "192.0.2.4"),
				Latency:    2 * time.Millisecond,
				RecordType: "A",
				Tries:      1,
				Rcode:      0,
				DPIChecked: true,
			},
		},
		{
			name:   "latency minimum one millisecond",
			record: []string{"192.0.2.5", "100us", "TXT", "1", "0", "skipped"},
			want: ResolverResult{
				IP:         mustAddr(t, "192.0.2.5"),
				Latency:    1 * time.Millisecond,
				RecordType: "TXT",
				Tries:      1,
				Rcode:      0,
				DPIChecked: false,
			},
		},
		{
			name:   "tries zero coerced to one",
			record: []string{"192.0.2.6", "18ms", "MX", "0", "2", "passed"},
			want: ResolverResult{
				IP:         mustAddr(t, "192.0.2.6"),
				Latency:    18 * time.Millisecond,
				RecordType: "MX",
				Tries:      1,
				Rcode:      2,
				DPIChecked: true,
			},
		},
		{
			name:      "too few fields non-legacy",
			record:    []string{"192.0.2.7", "10ms", "A"},
			wantError: true,
		},
		{
			name:      "invalid ip",
			record:    []string{"not-an-ip", "10ms", "A", "1", "0", "passed"},
			wantError: true,
		},
		{
			name:      "invalid latency",
			record:    []string{"192.0.2.8", "not-a-duration", "A", "1", "0", "passed"},
			wantError: true,
		},
		{
			name:      "invalid tries",
			record:    []string{"192.0.2.9", "10ms", "A", "bad", "0", "passed"},
			wantError: true,
		},
		{
			name:      "invalid rcode",
			record:    []string{"192.0.2.10", "10ms", "A", "1", "bad", "passed"},
			wantError: true,
		},
		{
			name:   "unknown dpi string treated as false",
			record: []string{"192.0.2.11", "11ms", "AAAA", "1", "0", "unknown"},
			want: ResolverResult{
				IP:         mustAddr(t, "192.0.2.11"),
				Latency:    11 * time.Millisecond,
				RecordType: "AAAA",
				Tries:      1,
				Rcode:      0,
				DPIChecked: false,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseResolverResult(tt.record)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseResolverResult() error = %v", err)
			}

			res, ok := got.(ResolverResult)
			if !ok {
				t.Fatalf("parseResolverResult() type = %T, want ResolverResult", got)
			}

			if res.IP != tt.want.IP {
				t.Fatalf("IP = %v, want %v", res.IP, tt.want.IP)
			}
			if res.Latency != tt.want.Latency {
				t.Fatalf("Latency = %v, want %v", res.Latency, tt.want.Latency)
			}
			if res.RecordType != tt.want.RecordType {
				t.Fatalf("RecordType = %q, want %q", res.RecordType, tt.want.RecordType)
			}
			if res.Tries != tt.want.Tries {
				t.Fatalf("Tries = %d, want %d", res.Tries, tt.want.Tries)
			}
			if res.Rcode != tt.want.Rcode {
				t.Fatalf("Rcode = %d, want %d", res.Rcode, tt.want.Rcode)
			}
			if res.DPIChecked != tt.want.DPIChecked {
				t.Fatalf("DPIChecked = %v, want %v", res.DPIChecked, tt.want.DPIChecked)
			}
		})
	}
}

func TestParseResolverResultLegacy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		record    []string
		want      ResolverResult
		wantError bool
	}{
		{
			name:   "valid legacy record",
			record: []string{"198.51.100.1", "9ms"},
			want: ResolverResult{
				IP:         mustAddr(t, "198.51.100.1"),
				Latency:    9 * time.Millisecond,
				RecordType: "?",
				Tries:      1,
				Rcode:      0,
				DPIChecked: false,
			},
		},
		{
			name:   "legacy latency rounded up and clamped",
			record: []string{"198.51.100.2", "100us"},
			want: ResolverResult{
				IP:         mustAddr(t, "198.51.100.2"),
				Latency:    1 * time.Millisecond,
				RecordType: "?",
				Tries:      1,
				Rcode:      0,
				DPIChecked: false,
			},
		},
		{
			name:      "legacy invalid ip",
			record:    []string{"bad-ip", "9ms"},
			wantError: true,
		},
		{
			name:      "legacy invalid latency",
			record:    []string{"198.51.100.3", "bad-duration"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseResolverResultLegacy(tt.record)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseResolverResultLegacy() error = %v", err)
			}

			res, ok := got.(ResolverResult)
			if !ok {
				t.Fatalf("parseResolverResultLegacy() type = %T, want ResolverResult", got)
			}

			if res.IP != tt.want.IP {
				t.Fatalf("IP = %v, want %v", res.IP, tt.want.IP)
			}
			if res.Latency != tt.want.Latency {
				t.Fatalf("Latency = %v, want %v", res.Latency, tt.want.Latency)
			}
			if res.RecordType != tt.want.RecordType {
				t.Fatalf("RecordType = %q, want %q", res.RecordType, tt.want.RecordType)
			}
			if res.Tries != tt.want.Tries {
				t.Fatalf("Tries = %d, want %d", res.Tries, tt.want.Tries)
			}
			if res.Rcode != tt.want.Rcode {
				t.Fatalf("Rcode = %d, want %d", res.Rcode, tt.want.Rcode)
			}
			if res.DPIChecked != tt.want.DPIChecked {
				t.Fatalf("DPIChecked = %v, want %v", res.DPIChecked, tt.want.DPIChecked)
			}
		})
	}
}
