package validate

import (
	"time"

	"github.com/MohsenBg/bgscan/internal/core/config"
)

const (
	MinICMPWorkers = 1
	MaxICMPWorkers = 5000

	MinICMPTries = 1
	MaxICMPTries = 10
)

const (
	MinICMPTimeout = 100 * time.Millisecond
	MaxICMPTimeout = 30 * time.Second
)

// ValidateICMP strictly validates an ICMPConfig and returns errors by field name.
func ValidateICMP(cfg config.ICMPConfig) map[string]error {
	errs := map[string]error{}

	if err := checkInt("Workers", cfg.Workers, MinICMPWorkers, MaxICMPWorkers); err != nil {
		errs["Workers"] = err
	}

	if err := checkDuration("Timeout", cfg.Timeout.Duration(),
		MinICMPTimeout, MaxICMPTimeout); err != nil {
		errs["Timeout"] = err
	}

	if err := checkUint16("Tries", cfg.Tries, MinICMPTries, MaxICMPTries); err != nil {
		errs["Tries"] = err
	}

	if err := checkPrefix("OutputPrefix", cfg.OutputPrefix); err != nil {
		errs["OutputPrefix"] = err
	}

	return errs
}

// NormalizeICMP replaces invalid ICMPConfig fields with defaults and reports each correction.
func NormalizeICMP(cfg *config.ICMPConfig) []Warning {
	def := config.DefaultICMPConfig()
	var warns []Warning

	fixInt("Workers", &cfg.Workers,
		MinICMPWorkers, MaxICMPWorkers,
		def.Workers, &warns)

	fixDurationMS("Timeout", &cfg.Timeout,
		MinICMPTimeout, MaxICMPTimeout,
		def.Timeout, &warns)

	fixUint16("Tries", &cfg.Tries,
		MinICMPTries, MaxICMPTries,
		def.Tries, &warns)

	fixString("OutputPrefix",
		&cfg.OutputPrefix,
		def.OutputPrefix,
		&warns)

	return warns
}
