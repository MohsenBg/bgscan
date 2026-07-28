package config

import "time"

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
	OutputPrefix        string     `toml:"prefix_output"`
	AcceptedStatusCodes []int      `toml:"accepted_status_codes"`
}

// DefaultHTTPConfig returns the default configuration for HTTP probing.
func DefaultHTTPConfig() HTTPConfig {
	return HTTPConfig{
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
		OutputPrefix:        "http_",
	}
}
