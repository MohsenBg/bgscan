package config

import (
	"time"

	"bgscan/internal/logger"
)

// ICMPConfig defines configuration for ICMP probing.
type ICMPConfig struct {
	Workers      int        `toml:"workers" comment:"Number of concurrent ICMP echo requests in flight."`
	Timeout      DurationMS `toml:"timeout" comment:"Maximum time to wait for an ICMP echo reply, in milliseconds."`
	Tries        uint16     `toml:"tries" comment:"Maximum number of echo request attempts per target."`
	OutputPrefix string     `toml:"output_prefix" comment:"Filename prefix for ICMP scan results."`
}

func DefaultICMPConfig() ICMPConfig {
	platform := DetectPlatform()
	tier := SelectTier(CheckResources())

	tiers, ok := icmpDefaults[platform]
	if !ok {
		logger.CoreWarn("bgscan: no ICMP defaults for platform %q, falling back to %q", platform, Desktop)
		platform = Desktop
		tiers = icmpDefaults[platform]
	}

	cfg, ok := tiers[tier]
	if !ok {
		logger.CoreWarn("bgscan: no ICMP defaults for platform %q tier %q, falling back to %q", platform, tier, Mid)
		cfg = tiers[Mid]
	}

	return cfg
}

var icmpDefaults = map[Platform]map[Tier]ICMPConfig{
	Server: {
		Low:  ICMPConfig{Workers: 100, Timeout: NewDurationMS(2 * time.Second), Tries: 1, OutputPrefix: "icmp_"},
		Mid:  ICMPConfig{Workers: 400, Timeout: NewDurationMS(2 * time.Second), Tries: 1, OutputPrefix: "icmp_"},
		High: ICMPConfig{Workers: 1000, Timeout: NewDurationMS(1500 * time.Millisecond), Tries: 1, OutputPrefix: "icmp_"},
	},
	Desktop: {
		Low:  ICMPConfig{Workers: 30, Timeout: NewDurationMS(2 * time.Second), Tries: 1, OutputPrefix: "icmp_"},
		Mid:  ICMPConfig{Workers: 150, Timeout: NewDurationMS(2 * time.Second), Tries: 1, OutputPrefix: "icmp_"},
		High: ICMPConfig{Workers: 300, Timeout: NewDurationMS(2 * time.Second), Tries: 1, OutputPrefix: "icmp_"},
	},
	Android: {
		Low:  ICMPConfig{Workers: 15, Timeout: NewDurationMS(3 * time.Second), Tries: 1, OutputPrefix: "icmp_"},
		Mid:  ICMPConfig{Workers: 60, Timeout: NewDurationMS(3 * time.Second), Tries: 1, OutputPrefix: "icmp_"},
		High: ICMPConfig{Workers: 100, Timeout: NewDurationMS(2500 * time.Millisecond), Tries: 1, OutputPrefix: "icmp_"},
	},
}
