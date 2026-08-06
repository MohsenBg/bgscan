package config

import "time"

// XrayConfig defines configuration for Xray connectivity testing.
type XrayConfig struct {
	Workers              int              `toml:"workers"`
	ConnectivityTestType ConnectivityTest `toml:"connectivity_test_type"`
	DownloadSpeed        int              `toml:"download_speed"`
	UploadSpeed          int              `toml:"upload_speed"`
	Timeout              DurationMS       `toml:"timeout"`
	OutputPrefix         string           `toml:"prefix_output"`
	PreScanType          string           `toml:"pre_scan_type"`
}

// DefaultXrayConfig returns the default configuration for Xray connectivity testing.
func DefaultXrayConfig() XrayConfig {
	return XrayConfig{
		Workers:              32,
		ConnectivityTestType: Both,
		DownloadSpeed:        100,
		UploadSpeed:          50,
		PreScanType:          "tcp",
		Timeout:              NewDurationMS(4 * time.Second),
		OutputPrefix:         "xray_",
	}
}
