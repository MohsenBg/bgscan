package config

import "testing"

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
