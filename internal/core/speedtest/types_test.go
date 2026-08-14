package speedtest

import (
	"math"
	"testing"
)

func TestBitsPerSec_String(t *testing.T) {
	tests := []struct {
		input    BitsPerSec
		expected string
	}{
		{0, "0 bps"},
		{Bps * 500, "500 bps"},
		{Kbps, "1.00 Kbps"},
		{Kbps * 1234 / 10, "123.40 Kbps"}, // 123.4 Kbps
		{Mbps, "1.00 Mbps"},
		{Mbps * 5678 / 100, "56.78 Mbps"},
		{Gbps, "1.00 Gbps"},
		{Gbps * 25 / 10, "2.50 Gbps"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.input.String(); got != tt.expected {
				t.Errorf("BitsPerSec.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestBitsPerSec_Bps(t *testing.T) {
	val := BitsPerSec(5000)
	if val.Bps() != 5000 {
		t.Errorf("Bps() = %d, want 5000", val.Bps())
	}
}

func TestBelowMinimum(t *testing.T) {
	tests := []struct {
		name     string
		result   SpeedResult
		expected bool
	}{
		{
			name:     "no minimum set",
			result:   SpeedResult{Speed: 100 * Mbps, MinSpeed: 0},
			expected: false,
		},
		{
			name:     "above minimum",
			result:   SpeedResult{Speed: 150 * Mbps, MinSpeed: 100 * Mbps},
			expected: false,
		},
		{
			name:     "equal to minimum",
			result:   SpeedResult{Speed: 100 * Mbps, MinSpeed: 100 * Mbps},
			expected: false,
		},
		{
			name:     "below minimum",
			result:   SpeedResult{Speed: 80 * Mbps, MinSpeed: 100 * Mbps},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.BelowMinimum(); got != tt.expected {
				t.Errorf("BelowMinimum() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBitsPerSecCalc(t *testing.T) {
	tests := []struct {
		name     string
		bytes    uint64
		seconds  float64
		expected BitsPerSec
	}{
		{"zero bytes", 0, 1.0, 0},
		{"zero seconds", 1000, 0.0, 0},
		{"negative seconds", 1000, -1.5, 0},
		{"normal calculation", 125_000, 1.0, 1 * Mbps}, // 125000 bytes * 8 bits = 1,000,000 bits = 1 Mbps
		{"fractional seconds", 125_000, 0.5, 2 * Mbps},
		{"overflow protection", math.MaxUint64, 0.1, BitsPerSec(math.MaxUint64)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bitsPerSec(tt.bytes, tt.seconds)
			if got != tt.expected {
				t.Errorf("bitsPerSec() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseBitsPerSec(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected BitsPerSec
		wantErr  bool
	}{
		{"bps", "100 bps", 100 * Bps, false},
		{"kbps", "1.5 Kbps", 1500 * Bps, false},
		{"mbps", "50 Mbps", 50 * Mbps, false},
		{"gbps", "2.5 Gbps", 2500 * Mbps, false},
		{"spaces and case", "  10 mbps  ", 10 * Mbps, false},

		{"zero", "0 bps", 0, false},
		{"negative", "-1 Mbps", 0, true},

		// Overflow tests
		{"overflow bps", "18446744073709551616 bps", 0, true},
		{"overflow gbps", "18446744074 Gbps", 0, true},
		{"huge overflow", "5000000000000000000000000000000000000000 gbps", 0, true},

		{"invalid number", "invalid_num mbps", 0, true},
		{"unknown unit", "100 unknown", 0, true},
		{"missing unit", "no_unit", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBitsPerSec(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseBitsPerSec() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got != tt.expected {
				t.Errorf("ParseBitsPerSec() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseBitsPerSec_OverflowBoundary(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected BitsPerSec
	}{
		{
			name:     "exact max uint64",
			input:    "18446744073709551615 bps",
			expected: BitsPerSec(math.MaxUint64),
			wantErr:  false,
		},
		{
			name:    "one above max uint64",
			input:   "18446744073709551616 bps",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBitsPerSec(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseBitsPerSec() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got != tt.expected {
				t.Fatalf("got %v, want %v", got, tt.expected)
			}
		})
	}
}
