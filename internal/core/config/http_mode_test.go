package config

import "testing"

// TestParseHTTPMode checks that ParseHTTPMode accepts valid modes in any case, defaults empty input to host mode, and rejects unknown modes.
func TestParseHTTPMode(t *testing.T) {
	tests := []struct {
		in      string
		want    HTTPMode
		wantErr bool
	}{
		{"", HTTPModeHost, false},
		{"host", HTTPModeHost, false},
		{"HOST", HTTPModeHost, false},
		{" host ", HTTPModeHost, false},
		{"target", HTTPModeTarget, false},
		{" target ", HTTPModeTarget, false},
		{"bogus", "", true},
	}

	for _, tt := range tests {
		got, err := ParseHTTPMode(tt.in)
		if (err != nil) != tt.wantErr {
			t.Fatalf("ParseHTTPMode(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
		}
		if got != tt.want {
			t.Errorf("ParseHTTPMode(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
