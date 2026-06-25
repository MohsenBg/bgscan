package config

import (
	"strings"
	"time"
)

// GeneralConfig defines global scanner behavior and execution settings.
type GeneralConfig struct {
	StatusInterval DurationMS `toml:"status_interval"`
	StopAfterFound int        `toml:"stop_after_found"`
	MaxIPsToTest   int        `toml:"max_ips_to_test"`
	PipelineMode   string     `toml:"pipeline_mode"`
	MaxIPsPerStage int        `toml:"max_ips_per_stage"`
	BatchSize      int        `toml:"batch_size"`
	Shuffled       bool       `toml:"shuffled"`
}

// Normalize validates general configuration values and adjusts them to
// allowed ranges when necessary. All corrections are recorded in the
// provided ValidationReport.
func (g *GeneralConfig) Normalize(rep *ValidationReport) {
	def := DefaultGeneralConfig()

	// StatusInterval must be between 100ms and 1 minute.
	normalizeDuration(
		"General.StatusInterval",
		&g.StatusInterval,
		NewDurationMS(100*time.Millisecond),
		NewDurationMS(time.Minute),
		def.StatusInterval,
		rep,
	)

	// StopAfterFound must be non-negative.
	if g.StopAfterFound < 0 {
		old := g.StopAfterFound
		g.StopAfterFound = def.StopAfterFound
		rep.AddChange("General.StopAfterFound", old, g.StopAfterFound, "negative → default")
	}

	// MaxIPsToTest must be non-negative.
	if g.MaxIPsToTest < 0 {
		old := g.MaxIPsToTest
		g.MaxIPsToTest = def.MaxIPsToTest
		rep.AddChange("General.MaxIPsToTest", old, g.MaxIPsToTest, "negative → default")
	}

	// MaxIPsPerStage must be within a safe range.
	normalizeInt(
		"General.MaxIPsPerStage",
		&g.MaxIPsPerStage,
		1,
		10_000_000,
		def.MaxIPsPerStage,
		rep,
	)

	// BatchSize must be within a safe range.
	normalizeInt(
		"General.BatchSize",
		&g.BatchSize,
		1,
		10_000_000,
		def.BatchSize,
		rep,
	)

	// Validate PiplineMode.
	mode := strings.ToLower(strings.TrimSpace(g.PipelineMode))
	switch mode {
	case "sequential", "simple", "streaming", "parallel", "batch", "pipeline":
		g.PipelineMode = mode
	default:
		old := g.PipelineMode
		g.PipelineMode = def.PipelineMode
		rep.AddChange(
			"General.PiplineMode",
			old,
			g.PipelineMode,
			"invalid → default (allowed: sequential, streaming,  batch)",
		)
	}
}
