package config

import (
	"time"

	"bgscan/internal/logger"
)

// GeneralConfig defines global scanner behavior and execution settings.
type GeneralConfig struct {
	StatusInterval   DurationMS `toml:"status_interval" comment:"Interval between status updates, in milliseconds."`
	StopAfterFound   int        `toml:"stop_after_found" comment:"Stop after this many successful results. 0 disables the limit."`
	MaxIPsToTest     int        `toml:"max_ips_to_test" comment:"Maximum number of IPs to test. 0 disables the limit."`
	PipelineMode     string     `toml:"pipeline_mode" comment:"Pipeline execution mode: sequential, streaming, or batch."`
	MaxIPsPerStage   int        `toml:"max_ips_per_stage" comment:"Maximum number of IPs buffered by each pipeline stage."`
	BatchSize        int        `toml:"batch_size" comment:"Number of IPs processed in each batch."`
	Shuffled         bool       `toml:"shuffled" comment:"Randomize IP order before scanning."`
	MinProbeDuration DurationMS `toml:"min_probe_duration" comment:"Minimum duration allocated to each probe, in milliseconds."`
	ProbePerSec      int        `toml:"probe_per_sec" comment:"Maximum number of probes issued per second."`
	ProbeBurst       int        `toml:"probe_burst" comment:"Maximum number of probes allowed in a burst."`
}

func DefaultGeneralConfig() GeneralConfig {
	platform := DetectPlatform()
	tier := SelectTier(CheckResources())

	tiers, ok := generalDefaults[platform]
	if !ok {
		logger.CoreWarn("bgscan: no General defaults for platform %q, falling back to %q", platform, Desktop)
		platform = Desktop
		tiers = generalDefaults[platform]
	}

	cfg, ok := tiers[tier]
	if !ok {
		logger.CoreWarn("bgscan: no General defaults for platform %q tier %q, falling back to %q", platform, tier, Mid)
		cfg = tiers[Mid]
	}

	return cfg
}

var generalDefaults = map[Platform]map[Tier]GeneralConfig{
	Server: {
		Low: GeneralConfig{
			StatusInterval: NewDurationMS(1 * time.Second), StopAfterFound: 0, MaxIPsToTest: 0,
			PipelineMode: "streaming", MaxIPsPerStage: 50_000, BatchSize: 2_000, Shuffled: true,
			MinProbeDuration: NewDurationMS(20 * time.Millisecond), ProbePerSec: 400, ProbeBurst: 100,
		},
		Mid: GeneralConfig{
			StatusInterval: NewDurationMS(1 * time.Second), StopAfterFound: 0, MaxIPsToTest: 0,
			PipelineMode: "streaming", MaxIPsPerStage: 200_000, BatchSize: 10_000, Shuffled: true,
			MinProbeDuration: NewDurationMS(10 * time.Millisecond), ProbePerSec: 800, ProbeBurst: 200,
		},
		High: GeneralConfig{
			StatusInterval: NewDurationMS(1 * time.Second), StopAfterFound: 0, MaxIPsToTest: 0,
			PipelineMode: "streaming", MaxIPsPerStage: 500_000, BatchSize: 20_000, Shuffled: true,
			MinProbeDuration: NewDurationMS(5 * time.Millisecond), ProbePerSec: 2000, ProbeBurst: 300,
		},
	},
	Desktop: {
		Low: GeneralConfig{
			StatusInterval: NewDurationMS(1 * time.Second), StopAfterFound: 0, MaxIPsToTest: 0,
			PipelineMode: "streaming", MaxIPsPerStage: 20_000, BatchSize: 1_000, Shuffled: true,
			MinProbeDuration: NewDurationMS(50 * time.Millisecond), ProbePerSec: 200, ProbeBurst: 50,
		},
		Mid: GeneralConfig{
			StatusInterval: NewDurationMS(1 * time.Second), StopAfterFound: 0, MaxIPsToTest: 0,
			PipelineMode: "streaming", MaxIPsPerStage: 100_000, BatchSize: 5_000, Shuffled: true,
			MinProbeDuration: NewDurationMS(50 * time.Millisecond), ProbePerSec: 500, ProbeBurst: 100,
		},
		High: GeneralConfig{
			StatusInterval: NewDurationMS(1 * time.Second), StopAfterFound: 0, MaxIPsToTest: 0,
			PipelineMode: "streaming", MaxIPsPerStage: 100_000, BatchSize: 5_000, Shuffled: true,
			MinProbeDuration: NewDurationMS(50 * time.Millisecond), ProbePerSec: 800, ProbeBurst: 100,
		},
	},
	Android: {
		Low: GeneralConfig{
			StatusInterval: NewDurationMS(1 * time.Second), StopAfterFound: 0, MaxIPsToTest: 0,
			PipelineMode: "streaming", MaxIPsPerStage: 10_000, BatchSize: 500, Shuffled: true,
			MinProbeDuration: NewDurationMS(100 * time.Millisecond), ProbePerSec: 120, ProbeBurst: 30,
		},
		Mid: GeneralConfig{
			StatusInterval: NewDurationMS(1 * time.Second), StopAfterFound: 0, MaxIPsToTest: 0,
			PipelineMode: "streaming", MaxIPsPerStage: 50_000, BatchSize: 2_000, Shuffled: true,
			MinProbeDuration: NewDurationMS(80 * time.Millisecond), ProbePerSec: 160, ProbeBurst: 50,
		},
		High: GeneralConfig{
			StatusInterval: NewDurationMS(1 * time.Second), StopAfterFound: 0, MaxIPsToTest: 0,
			PipelineMode: "streaming", MaxIPsPerStage: 100_000, BatchSize: 5_000, Shuffled: true,
			MinProbeDuration: NewDurationMS(50 * time.Millisecond), ProbePerSec: 220, ProbeBurst: 100,
		},
	},
}
