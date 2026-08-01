package slipstreamprobe

import (
	"net/netip"
	"testing"
	"time"

	"bgscan/internal/core/result"
)

// mustParseIP parses an IP address string, failing the test if invalid.
func mustParseIP(t *testing.T, s string) netip.Addr {
	t.Helper()

	ip, err := netip.ParseAddr(s)
	if err != nil {
		t.Fatalf("ParseAddr(%q) error = %v", s, err)
	}

	return ip
}

func TestSchema(t *testing.T) {
	t.Parallel()

	if Schema.Name != "Slipstream" {
		t.Fatalf("Schema.Name = %q, want %q", Schema.Name, "Slipstream")
	}

	if Schema.Directory != "slipstream" {
		t.Fatalf("Schema.Directory = %q, want %q", Schema.Directory, "slipstream")
	}

	if len(Schema.Columns) != 3 {
		t.Fatalf("len(Schema.Columns) = %d, want %d", len(Schema.Columns), 3)
	}

	wantCols := []struct {
		name  string
		width int
	}{
		{name: "IP", width: 45},
		{name: "Latency", width: 35},
		{name: "Port", width: 20},
	}

	for i, want := range wantCols {
		if Schema.Columns[i].Name != want.name {
			t.Fatalf("Schema.Columns[%d].Name = %q, want %q", i, Schema.Columns[i].Name, want.name)
		}
		if Schema.Columns[i].Width != want.width {
			t.Fatalf("Schema.Columns[%d].Width = %d, want %d", i, Schema.Columns[i].Width, want.width)
		}
	}

	if Schema.Parser == nil {
		t.Fatal("Schema.Parser is nil")
	}
}

func TestSlipstreamResultKey(t *testing.T) {
	t.Parallel()

	r := SlipstreamResult{
		IP: mustParseIP(t, "192.0.2.10"),
	}

	if got := r.Key(); got != "192.0.2.10" {
		t.Fatalf("Key() = %q, want %q", got, "192.0.2.10")
	}
}

func TestSlipstreamResultKeyType(t *testing.T) {
	t.Parallel()

	r := SlipstreamResult{}

	if got := r.KeyType(); got != result.KeyIP {
		t.Fatalf("KeyType() = %v, want %v", got, result.KeyIP)
	}
}

func TestSlipstreamResultEqual(t *testing.T) {
	t.Parallel()

	r1 := SlipstreamResult{IP: mustParseIP(t, "198.51.100.1")}
	r2 := SlipstreamResult{IP: mustParseIP(t, "198.51.100.1")}
	r3 := SlipstreamResult{IP: mustParseIP(t, "198.51.100.2")}

	if !r1.Equal(r2) {
		t.Fatal("Equal() = false, want true for same IP")
	}

	if r1.Equal(r3) {
		t.Fatal("Equal() = true, want false for different IP")
	}
}

func TestSlipstreamResultToRecord(t *testing.T) {
	t.Parallel()

	r := SlipstreamResult{
		IP:      mustParseIP(t, "203.0.113.5"),
		Latency: 25 * time.Millisecond,
		Port:    1080,
	}

	got := r.ToRecord()
	want := []string{"203.0.113.5", "25ms", "1080"}

	if len(got) != len(want) {
		t.Fatalf("len(ToRecord()) = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ToRecord()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSlipstreamResultScore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   SlipstreamResult
		want float64
	}{
		{
			name: "1ms",
			in: SlipstreamResult{
				Latency: 1 * time.Millisecond,
			},
			want: 1000,
		},
		{
			name: "10ms",
			in: SlipstreamResult{
				Latency: 10 * time.Millisecond,
			},
			want: 100,
		},
		{
			name: "sub-millisecond clamped to 1ms",
			in: SlipstreamResult{
				Latency: 500 * time.Microsecond,
			},
			want: 1000,
		},
		{
			name: "zero clamped to 1ms",
			in: SlipstreamResult{
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

// TestParseSlipstreamResult verifies parsing logic, including backward
// compatibility for legacy two-field records and handling of invalid inputs.
func TestParseSlipstreamResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		record    []string
		want      SlipstreamResult
		wantError bool
	}{
		{
			name:   "valid full record",
			record: []string{"192.0.2.1", "25ms", "1080"},
			want: SlipstreamResult{
				IP:      mustParseIP(t, "192.0.2.1"),
				Latency: 25 * time.Millisecond,
				Port:    1080,
			},
		},
		{
			name:   "legacy two field record",
			record: []string{"192.0.2.2", "40ms"},
			want: SlipstreamResult{
				IP:      mustParseIP(t, "192.0.2.2"),
				Latency: 40 * time.Millisecond,
				Port:    0,
			},
		},
		{
			name:   "extra fields ignored",
			record: []string{"192.0.2.3", "12ms", "9050", "extra"},
			want: SlipstreamResult{
				IP:      mustParseIP(t, "192.0.2.3"),
				Latency: 12 * time.Millisecond,
				Port:    9050,
			},
		},
		{
			name:      "too few fields",
			record:    []string{"192.0.2.4"},
			wantError: true,
		},
		{
			name:      "invalid ip",
			record:    []string{"not-an-ip", "10ms", "1080"},
			wantError: true,
		},
		{
			name:      "invalid latency",
			record:    []string{"192.0.2.5", "bad-duration", "1080"},
			wantError: true,
		},
		{
			name:      "invalid port",
			record:    []string{"192.0.2.6", "10ms", "bad-port"},
			wantError: true,
		},
		{
			name:      "negative port",
			record:    []string{"192.0.2.7", "10ms", "-1"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseSlipstreamResult(tt.record)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseSlipstreamResult() error = %v", err)
			}

			res, ok := got.(SlipstreamResult)
			if !ok {
				t.Fatalf("parseSlipstreamResult() type = %T, want SlipstreamResult", got)
			}

			if res.IP != tt.want.IP {
				t.Fatalf("IP = %v, want %v", res.IP, tt.want.IP)
			}
			if res.Latency != tt.want.Latency {
				t.Fatalf("Latency = %v, want %v", res.Latency, tt.want.Latency)
			}
			if res.Port != tt.want.Port {
				t.Fatalf("Port = %d, want %d", res.Port, tt.want.Port)
			}
		})
	}
}
