package config

import "time"

// DurationMS stores configuration durations as integer milliseconds.
type DurationMS int64

// NewDurationMS converts a time.Duration to milliseconds.
func NewDurationMS(d time.Duration) DurationMS {
	return DurationMS(d.Milliseconds())
}

// Duration converts milliseconds to a time.Duration.
func (d DurationMS) Duration() time.Duration {
	return time.Duration(d) * time.Millisecond
}

func (d *DurationMS) SetDuration(v time.Duration) {
	*d = DurationMS(v.Milliseconds())
}

// String implements fmt.Stringer.
func (d DurationMS) String() string {
	return d.Duration().String()
}

// ConnectivityTest identifies the Xray connectivity test to run.
type ConnectivityTest uint8

const (
	// ConnectivityOnly performs only a connectivity check.
	ConnectivityOnly ConnectivityTest = iota

	// DownloadSpeedOnly measures download speed only.
	DownloadSpeedOnly

	// UploadSpeedOnly measures upload speed only.
	UploadSpeedOnly

	// Both performs connectivity and speed tests.
	Both
)

// String implements fmt.Stringer.
func (c ConnectivityTest) String() string {
	names := [...]string{
		"Connectivity Only",
		"Download Speed Only",
		"Upload Speed Only",
		"Both",
	}

	if int(c) < len(names) {
		return names[c]
	}
	return "Unknown"
}

// IsValid reports whether c is a defined ConnectivityTest value.
func (c ConnectivityTest) IsValid() bool {
	return c <= Both
}
