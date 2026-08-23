package config

import (
	"time"

	"bgscan/internal/logger"
)

// TCPConfig defines configuration for TCP probing.
type TCPConfig struct {
	Workers      int        `toml:"workers" comment:"Concurrent TCP workers. Range: 1-5000. Higher = faster but more CPU/network."`
	Port         int        `toml:"port" comment:"TCP port to probe. Range: 1-65535. Common: 80 (HTTP), 443 (HTTPS), 22 (SSH)."`
	Timeout      DurationMS `toml:"timeout" comment:"Max wait for TCP handshake, in ms. Range: 100-30000. Lower = faster but may miss slow targets."`
	Tries        uint16     `toml:"tries" comment:"Retry attempts per target. Range: 1-10. Only retries on timeout, not on connection refused."`
	OutputPrefix string     `toml:"output_prefix" comment:"Filename prefix for result files."`
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
