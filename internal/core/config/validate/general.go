package validate

import (
	"fmt"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/config"
)

var allowedPipelineModes = []string{
	"sequential",
	"simple",
	"streaming",
	"parallel",
	"batch",
	"pipeline",
}

const (
	// Status interval limits
	MinGeneralStatusInterval = 100 * time.Millisecond
	MaxGeneralStatusInterval = time.Minute
)

const (
	// Counter limits
	MinGeneralStopAfterFound = 0
	MinGeneralMaxIPsToTest   = 0
)

const (
	// Scanner stage limits
	MinGeneralMaxIPsPerStage = 1
	MaxGeneralMaxIPsPerStage = 10_000_000

	MinGeneralBatchSize = 1
	MaxGeneralBatchSize = 10_000_000
)

const (
	// Probe duration limits
	MinGeneralProbeDuration = 10 * time.Millisecond
	MaxGeneralProbeDuration = 5 * time.Second
)

const (
	// Rate limit limits
	MinGeneralProbePerSec = 1
	MaxGeneralProbePerSec = 1_000_000

	MinGeneralProbeBurst = 1
	MaxGeneralProbeBurst = 10_000
)

// ValidateGeneral strictly validates a GeneralConfig and returns errors by field name.
func ValidateGeneral(cfg config.GeneralConfig) map[string]error {
	errs := map[string]error{}

	if err := checkDuration(
		"StatusInterval",
		cfg.StatusInterval.Duration(),
		MinGeneralStatusInterval,
		MaxGeneralStatusInterval,
	); err != nil {
		errs["StatusInterval"] = err
	}

	if cfg.StopAfterFound < MinGeneralStopAfterFound {
		errs["StopAfterFound"] = fmt.Errorf("must be non-negative")
	}

	if cfg.MaxIPsToTest < MinGeneralMaxIPsToTest {
		errs["MaxIPsToTest"] = fmt.Errorf("must be non-negative")
	}

	if err := checkInt(
		"MaxIPsPerStage",
		cfg.MaxIPsPerStage,
		MinGeneralMaxIPsPerStage,
		MaxGeneralMaxIPsPerStage,
	); err != nil {
		errs["MaxIPsPerStage"] = err
	}

	if err := checkInt(
		"BatchSize",
		cfg.BatchSize,
		MinGeneralBatchSize,
		MaxGeneralBatchSize,
	); err != nil {
		errs["BatchSize"] = err
	}

	if err := checkEnum(
		"PipelineMode",
		cfg.PipelineMode,
		allowedPipelineModes,
	); err != nil {
		errs["PipelineMode"] = err
	}

	if err := checkDuration(
		"MinProbeDuration",
		cfg.MinProbeDuration.Duration(),
		MinGeneralProbeDuration,
		MaxGeneralProbeDuration,
	); err != nil {
		errs["MinProbeDuration"] = err
	}

	if err := checkInt(
		"ProbePerSec",
		cfg.ProbePerSec,
		MinGeneralProbePerSec,
		MaxGeneralProbePerSec,
	); err != nil {
		errs["ProbePerSec"] = err
	}

	if err := checkInt(
		"ProbeBurst",
		cfg.ProbeBurst,
		MinGeneralProbeBurst,
		MaxGeneralProbeBurst,
	); err != nil {
		errs["ProbeBurst"] = err
	}

	return errs
}

// NormalizeGeneral replaces invalid GeneralConfig fields with defaults and reports each correction.
func NormalizeGeneral(cfg *config.GeneralConfig) []Warning {
	def := config.DefaultGeneralConfig()
	var warns []Warning

	fixDurationMS(
		"StatusInterval",
		&cfg.StatusInterval,
		MinGeneralStatusInterval,
		MaxGeneralStatusInterval,
		def.StatusInterval,
		&warns,
	)

	if cfg.StopAfterFound < MinGeneralStopAfterFound {
		warns = append(warns, Warning{
			Field:  "StopAfterFound",
			OldVal: cfg.StopAfterFound,
			NewVal: def.StopAfterFound,
			Reason: "negative → default",
		})

		cfg.StopAfterFound = def.StopAfterFound
	}

	if cfg.MaxIPsToTest < MinGeneralMaxIPsToTest {
		warns = append(warns, Warning{
			Field:  "MaxIPsToTest",
			OldVal: cfg.MaxIPsToTest,
			NewVal: def.MaxIPsToTest,
			Reason: "negative → default",
		})

		cfg.MaxIPsToTest = def.MaxIPsToTest
	}

	fixInt(
		"MaxIPsPerStage",
		&cfg.MaxIPsPerStage,
		MinGeneralMaxIPsPerStage,
		MaxGeneralMaxIPsPerStage,
		def.MaxIPsPerStage,
		&warns,
	)

	fixInt(
		"BatchSize",
		&cfg.BatchSize,
		MinGeneralBatchSize,
		MaxGeneralBatchSize,
		def.BatchSize,
		&warns,
	)

	fixEnum(
		"PipelineMode",
		&cfg.PipelineMode,
		allowedPipelineModes,
		def.PipelineMode,
		&warns,
	)

	fixDurationMS(
		"MinProbeDuration",
		&cfg.MinProbeDuration,
		MinGeneralProbeDuration,
		MaxGeneralProbeDuration,
		def.MinProbeDuration,
		&warns,
	)

	fixInt(
		"ProbePerSec",
		&cfg.ProbePerSec,
		MinGeneralProbePerSec,
		MaxGeneralProbePerSec,
		def.ProbePerSec,
		&warns,
	)

	fixInt(
		"ProbeBurst",
		&cfg.ProbeBurst,
		MinGeneralProbeBurst,
		MaxGeneralProbeBurst,
		def.ProbeBurst,
		&warns,
	)

	return warns
}
