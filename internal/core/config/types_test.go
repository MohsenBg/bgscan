package config

import (
	"testing"
	"time"
)

func TestNewDurationMS(t *testing.T) {
	tests := []struct {
		name  string
		input time.Duration
		want  DurationMS
	}{
		{"zero", 0, 0},
		{"1 second", 1 * time.Second, 1000},
		{"500ms", 500 * time.Millisecond, 500},
		{"2 minutes", 2 * time.Minute, 120_000},
		{"1 hour", 1 * time.Hour, 3_600_000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDurationMS(tt.input)
			if got != tt.want {
				t.Errorf("NewDurationMS(%v) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestDurationMS_Duration(t *testing.T) {
	tests := []struct {
		name  string
		input DurationMS
		want  time.Duration
	}{
		{"zero", 0, 0},
		{"1000ms", 1000, 1 * time.Second},
		{"500ms", 500, 500 * time.Millisecond},
		{"negative", -100, -100 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Duration()
			if got != tt.want {
				t.Errorf("DurationMS(%d).Duration() = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestDurationMS_SetDuration(t *testing.T) {
	var d DurationMS
	d.SetDuration(3 * time.Second)
	if d != 3000 {
		t.Errorf("after SetDuration(3s), got %d, want 3000", d)
	}

	d.SetDuration(0)
	if d != 0 {
		t.Errorf("after SetDuration(0), got %d, want 0", d)
	}
}

func TestDurationMS_RoundTrip(t *testing.T) {
	original := 7 * time.Second
	d := NewDurationMS(original)
	if d.Duration() != original {
		t.Errorf("round-trip failed: got %v, want %v", d.Duration(), original)
	}
}

func TestDurationMS_String(t *testing.T) {
	tests := []struct {
		input DurationMS
		want  string
	}{
		{0, "0s"},
		{1000, "1s"},
		{1500, "1.5s"},
		{500, "500ms"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.input.String()
			if got != tt.want {
				t.Errorf("DurationMS(%d).String() = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConnectivityTest_String(t *testing.T) {
	tests := []struct {
		input ConnectivityTest
		want  string
	}{
		{ConnectivityOnly, "Connectivity Only"},
		{DownloadSpeedOnly, "Download Speed Only"},
		{UploadSpeedOnly, "Upload Speed Only"},
		{Both, "Both"},
		{ConnectivityTest(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.input.String()
			if got != tt.want {
				t.Errorf("ConnectivityTest(%d).String() = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConnectivityTest_IsValid(t *testing.T) {
	valid := []ConnectivityTest{ConnectivityOnly, DownloadSpeedOnly, UploadSpeedOnly, Both}
	for _, v := range valid {
		if !v.IsValid() {
			t.Errorf("ConnectivityTest(%d).IsValid() = false, want true", v)
		}
	}

	invalid := []ConnectivityTest{ConnectivityTest(4), ConnectivityTest(99), ConnectivityTest(255)}
	for _, v := range invalid {
		if v.IsValid() {
			t.Errorf("ConnectivityTest(%d).IsValid() = true, want false", v)
		}
	}
}

func TestConnectivityTest_Constants(t *testing.T) {
	// Ensure iota order hasn't shifted — these values matter for TOML serialization.
	if ConnectivityOnly != 0 {
		t.Errorf("ConnectivityOnly = %d, want 0", ConnectivityOnly)
	}
	if DownloadSpeedOnly != 1 {
		t.Errorf("DownloadSpeedOnly = %d, want 1", DownloadSpeedOnly)
	}
	if UploadSpeedOnly != 2 {
		t.Errorf("UploadSpeedOnly = %d, want 2", UploadSpeedOnly)
	}
	if Both != 3 {
		t.Errorf("Both = %d, want 3", Both)
	}
}
