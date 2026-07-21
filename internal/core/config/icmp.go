package config

import "time"

// ICMPConfig defines configuration for ICMP probing.
type ICMPConfig struct {
	Workers      int        `toml:"workers"`
	Timeout      DurationMS `toml:"timeout"`
	Tries        uint16     `toml:"tries"`
	PrefixOutput string     `toml:"prefix_output"`
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
