package tcpprobe

import (
	"net/netip"
	"testing"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/target"
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

// mustEntry wraps an address in the single-address row a probe now receives.
func mustEntry(t *testing.T, s string) target.Entry {
	t.Helper()

	addr := mustAddr(t, s)

	return target.NewEnabled(netip.PrefixFrom(addr, addr.BitLen()))
}

func TestSchema(t *testing.T) {
	t.Parallel()

	if Schema.Name != "TCP" {
		t.Fatalf("Schema.Name = %q, want %q", Schema.Name, "TCP")
	}

	if Schema.Directory != "tcp" {
		t.Fatalf("Schema.Directory = %q, want %q", Schema.Directory, "tcp")
	}

	if len(Schema.Columns) != 5 {
		t.Fatalf("len(Schema.Columns) = %d, want %d", len(Schema.Columns), 5)
	}

	if Schema.Columns[0].Name != "Target" {
		t.Fatalf("Schema.Columns[0].Name = %q, want %q", Schema.Columns[0].Name, "Target")
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
	if Schema.Columns[4].Name != "ResolvedFrom" {
		t.Fatalf("Schema.Columns[4].Name = %q, want %q", Schema.Columns[4].Name, "ResolvedFrom")
	}

	if Schema.Parser == nil {
		t.Fatal("Schema.Parser is nil")
	}
}

func TestTCPResultKey(t *testing.T) {
	t.Parallel()

	entry := mustEntry(t, "192.0.2.10")
	r := TCPResult{Entry: entry}

	if got, want := r.Key(), entry.Key(); got != want {
		t.Fatalf("Key() = %q, want %q", got, want)
	}
}

func TestTCPResultKeyType(t *testing.T) {
	t.Parallel()

	r := TCPResult{}

	if got := r.KeyType(); got != target.KindIP {
		t.Fatalf("KeyType() = %v, want %v", got, target.KindIP)
	}
}

func TestTCPResultEqual(t *testing.T) {
	t.Parallel()

	r1 := TCPResult{Entry: mustEntry(t, "198.51.100.7")}
	r2 := TCPResult{Entry: mustEntry(t, "198.51.100.7")}
	r3 := TCPResult{Entry: mustEntry(t, "198.51.100.8")}

	if !r1.Equal(r2) {
		t.Fatal("Equal() = false, want true for same target")
	}

	if r1.Equal(r3) {
		t.Fatal("Equal() = true, want false for different target")
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
				Entry:   mustEntry(t, "203.0.113.5"),
				Port:    443,
				Latency: 23 * time.Millisecond,
				Tries:   4,
			},
			want: []string{"203.0.113.5/32", "0", "", "1", "23.00ms", "443", "4"},
		},
		{
			name: "zero port becomes dash",
			in: TCPResult{
				Entry:   mustEntry(t, "203.0.113.6"),
				Port:    0,
				Latency: 50 * time.Millisecond,
				Tries:   2,
			},
			want: []string{"203.0.113.6/32", "0", "", "1", "50.00ms", "-", "2"},
		},
		{
			name: "zero tries defaults to one",
			in: TCPResult{
				Entry:   mustEntry(t, "203.0.113.7"),
				Port:    80,
				Latency: 75 * time.Millisecond,
				Tries:   0,
			},
			want: []string{"203.0.113.7/32", "0", "", "1", "75.00ms", "80", "1"},
		},
	}

	for _, tt := range tests {
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
			record: []string{"192.0.2.1", "8443", "/probe", "1", "25ms", "443", "3"},
			want: TCPResult{
				Entry: target.Entry{
					Kind:   target.KindIP,
					IP:     netip.MustParsePrefix("192.0.2.1/32"),
					Port:   8443,
					Path:   "/probe",
					Enable: true,
				},
				Latency: 25 * time.Millisecond,
				Port:    443,
				Tries:   3,
			},
		},
		{
			name:   "record with port and tries",
			record: []string{"192.0.2.2", "0", "", "1", "40ms", "80", "1"},
			want: TCPResult{
				Entry:   mustEntry(t, "192.0.2.2"),
				Latency: 40 * time.Millisecond,
				Port:    80,
				Tries:   1,
			},
		},
		{
			name:   "invalid port falls back to zero",
			record: []string{"192.0.2.3", "0", "", "1", "12ms", "-", "2"},
			want: TCPResult{
				Entry:   mustEntry(t, "192.0.2.3"),
				Latency: 12 * time.Millisecond,
				Port:    0,
				Tries:   2,
			},
		},
		{
			name:   "latency rounded to nearest millisecond",
			record: []string{"192.0.2.4", "0", "", "1", "1500us", "22", "1"},
			want: TCPResult{
				Entry:   mustEntry(t, "192.0.2.4"),
				Latency: 2 * time.Millisecond,
				Port:    22,
				Tries:   1,
			},
		},
		{
			name:   "latency minimum one millisecond",
			record: []string{"192.0.2.5", "0", "", "1", "100us", "53", "1"},
			want: TCPResult{
				Entry:   mustEntry(t, "192.0.2.5"),
				Latency: 1 * time.Millisecond,
				Port:    53,
				Tries:   1,
			},
		},
		{
			name:   "zero tries coerced to one",
			record: []string{"192.0.2.6", "0", "", "1", "18ms", "8080", "0"},
			want: TCPResult{
				Entry:   mustEntry(t, "192.0.2.6"),
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
			record:    []string{"not-an-ip", "0", "", "1", "10ms", "443", "1"},
			wantError: true,
		},
		{
			name:      "invalid latency",
			record:    []string{"192.0.2.8", "0", "", "1", "not-a-duration", "443", "1"},
			wantError: true,
		},
		{
			name:      "invalid port",
			record:    []string{"192.0.2.10", "0", "", "1", "10ms", "bad", "1"},
			wantError: true,
		},
		{
			name:      "invalid tries",
			record:    []string{"192.0.2.9", "0", "", "1", "10ms", "443", "bad"},
			wantError: true,
		},
	}

	for _, tt := range tests {
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

			if res.Entry.Key() != tt.want.Entry.Key() {
				t.Fatalf("Entry = %v, want %v", res.Entry, tt.want.Entry)
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
