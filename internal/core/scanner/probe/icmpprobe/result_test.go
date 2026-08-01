package icmpprobe

import (
	"net/netip"
	"testing"
	"time"

	"bgscan/internal/core/result"
)

func mustAddr(t *testing.T, s string) netip.Addr {
	t.Helper()

	ip, err := netip.ParseAddr(s)
	if err != nil {
		t.Fatalf("ParseAddr(%q): %v", s, err)
	}

	return ip
}

func TestICMPResult_Key(t *testing.T) {
	t.Parallel()

	r := ICMPResult{
		IP: mustAddr(t, "1.2.3.4"),
	}

	got := r.Key()
	want := "1.2.3.4"

	if got != want {
		t.Fatalf("Key() = %q, want %q", got, want)
	}
}

func TestICMPResult_KeyType(t *testing.T) {
	t.Parallel()

	r := ICMPResult{}
	got := r.KeyType()

	if got != result.KeyIP {
		t.Fatalf("KeyType() = %v, want %v", got, result.KeyIP)
	}
}

func TestICMPResult_Equal(t *testing.T) {
	t.Parallel()

	base := ICMPResult{
		IP: mustAddr(t, "8.8.8.8"),
	}

	same := ICMPResult{
		IP: mustAddr(t, "8.8.8.8"),
	}

	diff := ICMPResult{
		IP: mustAddr(t, "1.1.1.1"),
	}

	if !base.Equal(same) {
		t.Fatal("Equal() = false, want true for same IP")
	}

	if base.Equal(diff) {
		t.Fatal("Equal() = true, want false for different IP")
	}
}

func TestICMPResult_ToRecord(t *testing.T) {
	t.Parallel()

	r := ICMPResult{
		IP:      mustAddr(t, "192.0.2.1"),
		Latency: 25 * time.Millisecond,
		Tries:   3,
		Mode:    "raw",
	}

	got := r.ToRecord()
	want := []string{
		"192.0.2.1",
		"25ms",
		"3",
		"raw",
	}

	if len(got) != len(want) {
		t.Fatalf("len(ToRecord()) = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ToRecord()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestICMPResult_ToRecord_DefaultTries(t *testing.T) {
	t.Parallel()

	r := ICMPResult{
		IP:      mustAddr(t, "192.0.2.2"),
		Latency: 10 * time.Millisecond,
		Tries:   0,
		Mode:    "udp",
	}

	got := r.ToRecord()
	want := []string{
		"192.0.2.2",
		"10ms",
		"1",
		"udp",
	}

	if len(got) != len(want) {
		t.Fatalf("len(ToRecord()) = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ToRecord()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestICMPResult_Score(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    ICMPResult
		want float64
	}{
		{
			name: "1ms",
			r: ICMPResult{
				Latency: 1 * time.Millisecond,
			},
			want: 1000.0,
		},
		{
			name: "10ms",
			r: ICMPResult{
				Latency: 10 * time.Millisecond,
			},
			want: 100.0,
		},
		{
			name: "250ms",
			r: ICMPResult{
				Latency: 250 * time.Millisecond,
			},
			want: 4.0,
		},
		{
			name: "sub-millisecond clamps to 1ms",
			r: ICMPResult{
				Latency: 500 * time.Microsecond,
			},
			want: 1000.0,
		},
		{
			name: "zero clamps to 1ms",
			r: ICMPResult{
				Latency: 0,
			},
			want: 1000.0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.r.Score()
			if got != tt.want {
				t.Fatalf("Score() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseICMPResult_LegacyTwoFields(t *testing.T) {
	t.Parallel()

	rec := []string{"203.0.113.10", "23ms"}

	got, err := parseICMPResult(rec)
	if err != nil {
		t.Fatalf("parseICMPResult() error = %v", err)
	}

	r, ok := got.(ICMPResult)
	if !ok {
		t.Fatalf("parseICMPResult() type = %T, want ICMPResult", got)
	}

	if r.IP != mustAddr(t, "203.0.113.10") {
		t.Fatalf("IP = %v, want %v", r.IP, mustAddr(t, "203.0.113.10"))
	}

	if r.Latency != 23*time.Millisecond {
		t.Fatalf("Latency = %v, want %v", r.Latency, 23*time.Millisecond)
	}

	if r.Tries != 1 {
		t.Fatalf("Tries = %d, want %d", r.Tries, 1)
	}

	if r.Mode != "" {
		t.Fatalf("Mode = %q, want empty string", r.Mode)
	}
}

func TestParseICMPResult_FourFields(t *testing.T) {
	t.Parallel()

	rec := []string{"203.0.113.11", "42ms", "4", "udp"}

	got, err := parseICMPResult(rec)
	if err != nil {
		t.Fatalf("parseICMPResult() error = %v", err)
	}

	r, ok := got.(ICMPResult)
	if !ok {
		t.Fatalf("parseICMPResult() type = %T, want ICMPResult", got)
	}

	if r.IP != mustAddr(t, "203.0.113.11") {
		t.Fatalf("IP = %v, want %v", r.IP, mustAddr(t, "203.0.113.11"))
	}

	if r.Latency != 42*time.Millisecond {
		t.Fatalf("Latency = %v, want %v", r.Latency, 42*time.Millisecond)
	}

	if r.Tries != 4 {
		t.Fatalf("Tries = %d, want %d", r.Tries, 4)
	}

	if r.Mode != "udp" {
		t.Fatalf("Mode = %q, want %q", r.Mode, "udp")
	}
}

func TestParseICMPResult_RoundsLatencyToMilliseconds(t *testing.T) {
	t.Parallel()

	rec := []string{"203.0.113.12", "1500us"}

	got, err := parseICMPResult(rec)
	if err != nil {
		t.Fatalf("parseICMPResult() error = %v", err)
	}

	r := got.(ICMPResult)

	if r.Latency != 2*time.Millisecond {
		t.Fatalf("Latency = %v, want %v", r.Latency, 2*time.Millisecond)
	}
}

func TestParseICMPResult_MinimumLatencyIsOneMillisecond(t *testing.T) {
	t.Parallel()

	rec := []string{"203.0.113.13", "200us"}

	got, err := parseICMPResult(rec)
	if err != nil {
		t.Fatalf("parseICMPResult() error = %v", err)
	}

	r := got.(ICMPResult)

	if r.Latency != 1*time.Millisecond {
		t.Fatalf("Latency = %v, want %v", r.Latency, 1*time.Millisecond)
	}
}

func TestParseICMPResult_TriesMinimumIsOne(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rec  []string
	}{
		{
			name: "tries zero",
			rec:  []string{"203.0.113.14", "5ms", "0", "raw"},
		},
		{
			name: "tries negative",
			rec:  []string{"203.0.113.15", "5ms", "-3", "raw"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseICMPResult(tt.rec)
			if err != nil {
				t.Fatalf("parseICMPResult() error = %v", err)
			}

			r := got.(ICMPResult)

			if r.Tries != 1 {
				t.Fatalf("Tries = %d, want %d", r.Tries, 1)
			}
		})
	}
}

func TestParseICMPResult_InvalidRecordLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rec  []string
	}{
		{
			name: "zero fields",
			rec:  []string{},
		},
		{
			name: "one field",
			rec:  []string{"203.0.113.10"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseICMPResult(tt.rec)
			if err == nil {
				t.Fatalf("expected error, got result = %#v", got)
			}
		})
	}
}

func TestParseICMPResult_InvalidLatency(t *testing.T) {
	t.Parallel()

	rec := []string{"203.0.113.16", "not-a-duration"}

	got, err := parseICMPResult(rec)
	if err == nil {
		t.Fatalf("expected error, got result = %#v", got)
	}
}

func TestParseICMPResult_InvalidTries(t *testing.T) {
	t.Parallel()

	rec := []string{"203.0.113.17", "10ms", "NaN", "udp"}

	got, err := parseICMPResult(rec)
	if err == nil {
		t.Fatalf("expected error, got result = %#v", got)
	}
}

func TestSchema(t *testing.T) {
	t.Parallel()

	if Schema.Name != "ICMP" {
		t.Fatalf("Schema.Name = %q, want %q", Schema.Name, "ICMP")
	}

	if Schema.Directory != "icmp" {
		t.Fatalf("Schema.Directory = %q, want %q", Schema.Directory, "icmp")
	}

	if len(Schema.Columns) != 4 {
		t.Fatalf("len(Schema.Columns) = %d, want %d", len(Schema.Columns), 4)
	}

	wantCols := []struct {
		name  string
		width int
	}{
		{"IP", 60},
		{"Latency", 20},
		{"Tries", 10},
		{"Mode", 10},
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

func TestSchemaParser(t *testing.T) {
	t.Parallel()

	got, err := Schema.Parser([]string{"198.51.100.20", "8ms", "2", "raw"})
	if err != nil {
		t.Fatalf("Schema.Parser() error = %v", err)
	}

	r, ok := got.(ICMPResult)
	if !ok {
		t.Fatalf("Schema.Parser() type = %T, want ICMPResult", got)
	}

	if r.IP != mustAddr(t, "198.51.100.20") {
		t.Fatalf("IP = %v, want %v", r.IP, mustAddr(t, "198.51.100.20"))
	}

	if r.Latency != 8*time.Millisecond {
		t.Fatalf("Latency = %v, want %v", r.Latency, 8*time.Millisecond)
	}

	if r.Tries != 2 {
		t.Fatalf("Tries = %d, want %d", r.Tries, 2)
	}

	if r.Mode != "raw" {
		t.Fatalf("Mode = %q, want %q", r.Mode, "raw")
	}
}

func TestICMPResult_ImplementsResult(t *testing.T) {
	t.Parallel()

	var _ result.Result = ICMPResult{}
}
