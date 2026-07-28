package validate

import (
	"fmt"
	"time"

	"bgscan/internal/core/config"
)

var allowedPipelineModes = []string{
	"sequential", "simple", "streaming", "parallel", "batch", "pipeline",
}

// ValidateGeneral strictly validates a GeneralConfig and returns errors by field name.
func ValidateGeneral(cfg config.GeneralConfig) map[string]error {
	errs := map[string]error{}

	if err := checkDuration("StatusInterval", cfg.StatusInterval.Duration(),
		100*time.Millisecond, time.Minute); err != nil {
		errs["StatusInterval"] = err
	}

	if cfg.StopAfterFound < 0 {
		errs["StopAfterFound"] = fmt.Errorf("must be non-negative")
	}

	if cfg.MaxIPsToTest < 0 {
		errs["MaxIPsToTest"] = fmt.Errorf("must be non-negative")
	}

	if err := checkInt("MaxIPsPerStage", cfg.MaxIPsPerStage, 1, 10_000_000); err != nil {
		errs["MaxIPsPerStage"] = err
	}

	if err := checkInt("BatchSize", cfg.BatchSize, 1, 10_000_000); err != nil {
		errs["BatchSize"] = err
	}

	if err := checkEnum("PipelineMode", cfg.PipelineMode, allowedPipelineModes); err != nil {
		errs["PipelineMode"] = err
	}

	return errs
}

// NormalizeGeneral replaces invalid GeneralConfig fields with defaults and reports each correction.
func NormalizeGeneral(cfg *config.GeneralConfig) []Warning {
	def := config.DefaultGeneralConfig()
	var warns []Warning

	fixDurationMS("StatusInterval", &cfg.StatusInterval,
		100*time.Millisecond, time.Minute, def.StatusInterval, &warns)

	if cfg.StopAfterFound < 0 {
		warns = append(warns, Warning{"StopAfterFound", cfg.StopAfterFound, def.StopAfterFound, "negative → default"})
		cfg.StopAfterFound = def.StopAfterFound
	}

	if cfg.MaxIPsToTest < 0 {
		warns = append(warns, Warning{"MaxIPsToTest", cfg.MaxIPsToTest, def.MaxIPsToTest, "negative → default"})
		cfg.MaxIPsToTest = def.MaxIPsToTest
	}

	fixInt("MaxIPsPerStage", &cfg.MaxIPsPerStage, 1, 10_000_000, def.MaxIPsPerStage, &warns)
	fixInt("BatchSize", &cfg.BatchSize, 1, 10_000_000, def.BatchSize, &warns)
	fixEnum("PipelineMode", &cfg.PipelineMode, allowedPipelineModes, def.PipelineMode, &warns)

	return warns
}
