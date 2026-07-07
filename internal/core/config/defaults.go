package config

import "time"

// DefaultGeneralConfig returns the default configuration for general scanner behavior.
func DefaultGeneralConfig() *GeneralConfig {
	return &GeneralConfig{
		StatusInterval: NewDurationMS(1 * time.Second),
		StopAfterFound: 0,
		MaxIPsToTest:   0,
		MaxIPsPerStage: 100_000,
		BatchSize:      5_000,
		Shuffled:       false,
		PipelineMode:   "streaming",
	}
}

// DefaultWriterConfig returns the default configuration for the result writer.
func DefaultWriterConfig() *WriterConfig {
	return &WriterConfig{
		MergeFlushInterval: NewDurationMS(2 * time.Second),
		ChanSize:           4096,
		BatchSize:          4096,
	}
}

// DefaultICMPConfig returns the default configuration for ICMP scanning.
func DefaultICMPConfig() *ICMPConfig {
	return &ICMPConfig{
		Workers:      200,
		Timeout:      NewDurationMS(2 * time.Second),
		Tries:        1,
		PrefixOutput: "icmp_",
	}
}

// DefaultTCPConfig returns the default configuration for TCP scanning.
func DefaultTCPConfig() *TCPConfig {
	return &TCPConfig{
		Workers:      200,
		Port:         80,
		Timeout:      NewDurationMS(3 * time.Second),
		Tries:        1,
		PrefixOutput: "tcp_",
	}
}

// DefaultHTTPConfig returns the default configuration for HTTP probing.
func DefaultHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		Workers:             50,
		Host:                "example.com",
		Port:                443,
		Protocol:            "https",
		TLSValidation:       true,
		Version:             "h1,h2",
		AcceptedStatusCodes: []int{},
		MinTLSVersion:       "tls1.1",
		MaxTLSVersion:       "tls1.3",
		Timeout:             NewDurationMS(4 * time.Second),
		PrefixOutput:        "http_",
	}
}

// DefaultXrayConfig returns the default configuration for Xray connectivity testing.
func DefaultXrayConfig() *XrayConfig {
	return &XrayConfig{
		Workers:              32,
		ConnectivityTestType: ConnectivityOnly,
		DownloadSpeed:        100,
		UploadSpeed:          50,
		PreScanType:          "none",
		Timeout:              NewDurationMS(6 * time.Second),
		PrefixOutput:         "xray_",
	}
}

// DefaultDNSConfig returns the default configuration for DNS‑based scanning methods.
func DefaultDNSConfig() *DNSConfig {
	return &DNSConfig{
		Resolver: &ResolverConfig{
			Workers:         100,
			Protocol:        "udp",
			Domain:          "google.com",
			Port:            53,
			CheckTypes:      []string{"A"},
			EDNSBufSize:     1234,
			Timeout:         NewDurationMS(2 * time.Second),
			Tries:           1,
			RandomSubdomain: true,
			AcceptedRCodes:  []string{"noerror", "nxdomain"},
			CheckDPI:        true,
			DPITimeout:      NewDurationMS(500 * time.Millisecond),
			DPITries:        2,
			PrefixOutput:    "dns_resolver_",
		},
		DNSTT: &DNSTTConfig{
			Enabled:      false,
			Workers:      20,
			Domain:       "ns.example.com",
			PublicKey:    "",
			Timeout:      NewDurationMS(8 * time.Second),
			PrefixOutput: "dns_dnstt_",
		},
		SlipStream: &SlipStreamConfig{
			Enabled:      false,
			Workers:      20,
			Domain:       "ns.example.com",
			CertPath:     "",
			Timeout:      NewDurationMS(8 * time.Second),
			PrefixOutput: "dns_slipstream_",
		},
	}
}
