package tcpprobe

import (
	"net/netip"
	"testing"
	"time"

	"bgscan/internal/core/result"
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

	if Schema.Name != "TCP" {
		t.Fatalf("Schema.Name = %q, want %q", Schema.Name, "TCP")
	}

	if Schema.Directory != "tcp" {
		t.Fatalf("Schema.Directory = %q, want %q", Schema.Directory, "tcp")
	}

	if len(Schema.Columns) != 4 {
		t.Fatalf("len(Schema.Columns) = %d, want %d", len(Schema.Columns), 4)
	}

	if Schema.Columns[0].Name != "IP" {
		t.Fatalf("Schema.Columns[0].Name = %q, want %q", Schema.Columns[0].Name, "IP")
	}
	if Schema.Columns[1].Name != "Latency" {
		t.Fatalf("Schema.Columns[1].Name = %q, want %q", Schema.Columns[1].Name, "Latency")
	}
	if Schema.Columns[2].Name != "Port" {
		t.Fatalf("Schema.Columns[2].Name = %q, want %q", Schema.Columns[2].Name, "Port")
	}
	if Schema.Columns[3].Name != "Tries" {
		t.Fatalf("Schema.Columns[3].Name = %q, want %q", Schema.Columns[3].Name, "Tries")
	}

	if Schema.Parser == nil {
		t.Fatal("Schema.Parser is nil")
	}
}

func TestTCPResultKey(t *testing.T) {
	t.Parallel()

	r := TCPResult{
		IP: mustAddr(t, "192.0.2.10"),
	}

	if got := r.Key(); got != "192.0.2.10" {
		t.Fatalf("Key() = %q, want %q", got, "192.0.2.10")
	}
}

func TestTCPResultKeyType(t *testing.T) {
	t.Parallel()

	r := TCPResult{}

	if got := r.KeyType(); got != result.KeyIP {
		t.Fatalf("KeyType() = %v, want %v", got, result.KeyIP)
	}
}

func TestTCPResultEqual(t *testing.T) {
	t.Parallel()

	r1 := TCPResult{
		IP: mustAddr(t, "198.51.100.7"),
	}
	r2 := TCPResult{
		IP: mustAddr(t, "198.51.100.7"),
	}
	r3 := TCPResult{
		IP: mustAddr(t, "198.51.100.8"),
	}

	if !r1.Equal(r2) {
		t.Fatal("Equal() = false, want true for same IP")
	}

	if r1.Equal(r3) {
		t.Fatal("Equal() = true, want false for different IP")
	}
}

func TestTCPResultToRecord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   TCPResult
		want []string
	}{
		{
			name: "full values",
			in: TCPResult{
				IP:      mustAddr(t, "203.0.113.5"),
				Port:    443,
				Latency: 23 * time.Millisecond,
				Tries:   4,
			},
			want: []string{"203.0.113.5", "23ms", "443", "4"},
		},
		{
			name: "zero port becomes dash",
			in: TCPResult{
				IP:      mustAddr(t, "203.0.113.6"),
				Port:    0,
				Latency: 50 * time.Millisecond,
				Tries:   2,
			},
			want: []string{"203.0.113.6", "50ms", "-", "2"},
		},
		{
			name: "zero tries defaults to one",
			in: TCPResult{
				IP:      mustAddr(t, "203.0.113.7"),
				Port:    80,
				Latency: 75 * time.Millisecond,
				Tries:   0,
			},
			want: []string{"203.0.113.7", "75ms", "80", "1"},
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

func TestTCPResultScore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   TCPResult
		want float64
	}{
		{
			name: "1ms",
			in: TCPResult{
				Latency: 1 * time.Millisecond,
			},
			want: 1000,
		},
		{
			name: "10ms",
			in: TCPResult{
				Latency: 10 * time.Millisecond,
			},
			want: 100,
		},
		{
			name: "sub-millisecond is clamped to 1ms",
			in: TCPResult{
				Latency: 500 * time.Microsecond,
			},
			want: 1000,
		},
		{
			name: "zero is clamped to 1ms",
			in: TCPResult{
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

// TestParseTCPResult verifies parsing logic, including backward compatibility
// for legacy records and handling of invalid or edge-case inputs.
func TestParseTCPResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		record    []string
		want      TCPResult
		wantError bool
	}{
		{
			name:   "valid full record",
			record: []string{"192.0.2.1", "25ms", "443", "3"},
			want: TCPResult{
				IP:      mustAddr(t, "192.0.2.1"),
				Latency: 25 * time.Millisecond,
				Port:    443,
				Tries:   3,
			},
		},
		{
			name:   "backward compatible record without tries",
			record: []string{"192.0.2.2", "40ms", "80"},
			want: TCPResult{
				IP:      mustAddr(t, "192.0.2.2"),
				Latency: 40 * time.Millisecond,
				Port:    80,
				Tries:   1,
			},
		},
		{
			name:   "invalid port falls back to zero",
			record: []string{"192.0.2.3", "12ms", "-", "2"},
			want: TCPResult{
				IP:      mustAddr(t, "192.0.2.3"),
				Latency: 12 * time.Millisecond,
				Port:    0,
				Tries:   2,
			},
		},
		{
			name:   "latency rounded to nearest millisecond",
			record: []string{"192.0.2.4", "1500us", "22", "1"},
			want: TCPResult{
				IP:      mustAddr(t, "192.0.2.4"),
				Latency: 2 * time.Millisecond,
				Port:    22,
				Tries:   1,
			},
		},
		{
			name:   "latency minimum one millisecond",
			record: []string{"192.0.2.5", "100us", "53", "1"},
			want: TCPResult{
				IP:      mustAddr(t, "192.0.2.5"),
				Latency: 1 * time.Millisecond,
				Port:    53,
				Tries:   1,
			},
		},
		{
			name:   "zero tries coerced to one",
			record: []string{"192.0.2.6", "18ms", "8080", "0"},
			want: TCPResult{
				IP:      mustAddr(t, "192.0.2.6"),
				Latency: 18 * time.Millisecond,
				Port:    8080,
				Tries:   1,
			},
		},
		{
			name:      "too few fields",
			record:    []string{"192.0.2.7", "10ms"},
			wantError: true,
		},
		{
			name:      "invalid ip",
			record:    []string{"not-an-ip", "10ms", "443", "1"},
			wantError: true,
		},
		{
			name:      "invalid latency",
			record:    []string{"192.0.2.8", "not-a-duration", "443", "1"},
			wantError: true,
		},
		{
			name:      "invalid tries",
			record:    []string{"192.0.2.9", "10ms", "443", "bad"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseTCPResult(tt.record)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseTCPResult() error = %v", err)
			}

			res, ok := got.(TCPResult)
			if !ok {
				t.Fatalf("parseTCPResult() type = %T, want TCPResult", got)
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
			if res.Tries != tt.want.Tries {
				t.Fatalf("Tries = %d, want %d", res.Tries, tt.want.Tries)
			}
		})
	}
}
