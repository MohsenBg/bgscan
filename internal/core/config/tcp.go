package config

import (
	"time"

	"bgscan/internal/logger"
)

// TCPConfig defines configuration for TCP probing.
type TCPConfig struct {
	Workers      int        `toml:"workers" comment:"Number of concurrent TCP connection attempts."`
	Port         int        `toml:"port" comment:"TCP port to probe on each IP address."`
	Timeout      DurationMS `toml:"timeout" comment:"Maximum time to wait for a TCP connection, in milliseconds."`
	Tries        uint16     `toml:"tries" comment:"Maximum number of connection attempts per target."`
	OutputPrefix string     `toml:"output_prefix" comment:"Filename prefix for TCP scan results."`
}

func DefaultTCPConfig() TCPConfig {
	platform := DetectPlatform()
	tier := SelectTier(CheckResources())

	tiers, ok := tcpDefaults[platform]
	if !ok {
		logger.CoreWarn("bgscan: no TCP defaults for platform %q, falling back to %q", platform, Desktop)
		platform = Desktop
		tiers = tcpDefaults[platform]
	}

	cfg, ok := tiers[tier]
	if !ok {
		logger.CoreWarn("bgscan: no TCP defaults for platform %q tier %q, falling back to %q", platform, tier, Mid)
		cfg = tiers[Mid]
	}

	return cfg
}

var tcpDefaults = map[Platform]map[Tier]TCPConfig{
	Server: {
		Low:  TCPConfig{Workers: 100, Port: 443, Timeout: NewDurationMS(2 * time.Second), Tries: 1, OutputPrefix: "tcp_"},
		Mid:  TCPConfig{Workers: 400, Port: 443, Timeout: NewDurationMS(2 * time.Second), Tries: 1, OutputPrefix: "tcp_"},
		High: TCPConfig{Workers: 1000, Port: 443, Timeout: NewDurationMS(1500 * time.Millisecond), Tries: 1, OutputPrefix: "tcp_"},
	},
	Desktop: {
		Low:  TCPConfig{Workers: 30, Port: 443, Timeout: NewDurationMS(2 * time.Second), Tries: 1, OutputPrefix: "tcp_"},
		Mid:  TCPConfig{Workers: 200, Port: 443, Timeout: NewDurationMS(2 * time.Second), Tries: 1, OutputPrefix: "tcp_"},
		High: TCPConfig{Workers: 400, Port: 443, Timeout: NewDurationMS(2 * time.Second), Tries: 1, OutputPrefix: "tcp_"},
	},
	Android: {
		Low:  TCPConfig{Workers: 20, Port: 443, Timeout: NewDurationMS(3 * time.Second), Tries: 1, OutputPrefix: "tcp_"},
		Mid:  TCPConfig{Workers: 100, Port: 443, Timeout: NewDurationMS(3 * time.Second), Tries: 1, OutputPrefix: "tcp_"},
		High: TCPConfig{Workers: 150, Port: 443, Timeout: NewDurationMS(2500 * time.Millisecond), Tries: 1, OutputPrefix: "tcp_"},
	},
}
