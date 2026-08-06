package config

import "time"

// TCPConfig defines configuration for TCP probing.
type TCPConfig struct {
	Workers      int        `toml:"workers"`
	Port         int        `toml:"port"`
	Timeout      DurationMS `toml:"timeout"`
	Tries        uint16     `toml:"tries"`
	OutputPrefix string     `toml:"prefix_output"`
}

// DefaultTCPConfig returns the default configuration for TCP scanning.
func DefaultTCPConfig() TCPConfig {
	return TCPConfig{
		Workers:      400,
		Port:         443,
		Timeout:      NewDurationMS(2 * time.Second),
		Tries:        1,
		OutputPrefix: "tcp_",
	}
}
