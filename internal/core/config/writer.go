package config

import (
	"time"

	"bgscan/internal/logger"
)

// WriterConfig defines configuration for the result writer subsystem.
type WriterConfig struct {
	MergeFlushInterval DurationMS `toml:"merge_flush_interval" comment:"Interval between result merge and flush operations, in milliseconds."`
	ChanSize           int        `toml:"chan_size" comment:"Capacity of the channel used to queue result writes."`
	BatchSize          int        `toml:"batch_size" comment:"Number of results processed in each write batch."`
	ResultBaseDir      string     `toml:"result_directory" comment:"Base directory where scan results are stored."`
}

func DefaultWriterConfig() WriterConfig {
	platform := DetectPlatform()
	tier := SelectTier(CheckResources())

	tiers, ok := writerDefaults[platform]
	if !ok {
		logger.CoreWarn("bgscan: no Writer defaults for platform %q, falling back to %q", platform, Desktop)
		platform = Desktop
		tiers = writerDefaults[platform]
	}

	cfg, ok := tiers[tier]
	if !ok {
		logger.CoreWarn("bgscan: no Writer defaults for platform %q tier %q, falling back to %q", platform, tier, Mid)
		cfg = tiers[Mid]
	}

	return cfg
}

// writerBase holds fields shared across every platform/tier.
func writerBase() WriterConfig {
	return WriterConfig{
		MergeFlushInterval: NewDurationMS(2 * time.Second),
		ResultBaseDir:      "result",
	}
}

var writerDefaults = map[Platform]map[Tier]WriterConfig{
	Server: {
		Low:  withWriter(writerBase(), 1024, 4096),
		Mid:  withWriter(writerBase(), 4096, 16384),
		High: withWriter(writerBase(), 16384, 65536),
	},
	Desktop: {
		Low:  withWriter(writerBase(), 512, 2048),
		Mid:  withWriter(writerBase(), 1024, 4096),
		High: withWriter(writerBase(), 2048, 8192),
	},
	Android: {
		Low:  withWriter(writerBase(), 128, 512),
		Mid:  withWriter(writerBase(), 256, 1024),
		High: withWriter(writerBase(), 512, 2048),
	},
}

// withWriter returns a copy of base with ChanSize and BatchSize overridden.
func withWriter(base WriterConfig, chanSize, batchSize int) WriterConfig {
	base.ChanSize = chanSize
	base.BatchSize = batchSize
	return base
}
