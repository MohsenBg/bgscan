package config

// ICMPConfig defines configuration for ICMP probing.
type ICMPConfig struct {
	Workers      int        `toml:"workers"`
	Timeout      DurationMS `toml:"timeout"`
	Tries        uint16     `toml:"tries"`
	PrefixOutput string     `toml:"prefix_output"`
}

// TCPConfig defines configuration for TCP probing.
type TCPConfig struct {
	Workers      int        `toml:"workers"`
	Port         int        `toml:"port"`
	Timeout      DurationMS `toml:"timeout"`
	Tries        uint16     `toml:"tries"`
	PrefixOutput string     `toml:"prefix_output"`
}

// HTTPConfig defines configuration for HTTP probing and TLS validation.
type HTTPConfig struct {
	Workers             int        `toml:"workers"`
	Host                string     `toml:"host"`
	ServerName          string     `toml:"server_name"`
	Port                int        `toml:"port"`
	Protocol            string     `toml:"protocol"`
	Version             string     `toml:"version"`
	TLSValidation       bool       `toml:"tls_validation"`
	MinTLSVersion       string     `toml:"min_tls_version"`
	MaxTLSVersion       string     `toml:"max_tls_version"`
	Timeout             DurationMS `toml:"timeout"`
	PrefixOutput        string     `toml:"prefix_output"`
	AcceptedStatusCodes []int      `toml:"accepted_status_codes"`
}

// XrayConfig defines configuration for Xray connectivity testing.
type XrayConfig struct {
	Workers              int              `toml:"workers"`
	ConnectivityTestType ConnectivityTest `toml:"connectivity_test_type"`
	DownloadSpeed        int              `toml:"download_speed"`
	UploadSpeed          int              `toml:"upload_speed"`
	Timeout              DurationMS       `toml:"timeout"`
	PrefixOutput         string           `toml:"prefix_output"`
	PreScanType          string           `toml:"pre_scan_type"`
}
