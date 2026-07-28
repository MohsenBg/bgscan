package validate

import (
	"time"

	"bgscan/internal/core/config"
)

// ValidateWriter strictly validates a WriterConfig and returns errors by field name.
func ValidateWriter(cfg config.WriterConfig) map[string]error {
	errs := map[string]error{}

	if err := checkDuration("MergeFlushInterval", cfg.MergeFlushInterval.Duration(),
		100*time.Millisecond, 5*time.Minute); err != nil {
		errs["MergeFlushInterval"] = err
	}

	if err := checkInt("ChanSize", cfg.ChanSize, 1, 1_000_000); err != nil {
		errs["ChanSize"] = err
	}

	if err := checkInt("BatchSize", cfg.BatchSize, 1, 1_000_000); err != nil {
		errs["BatchSize"] = err
	}

	if err := checkDirectoryName("ResultDirectory", cfg.ResultBaseDir); err != nil {
		errs["ResultBaseDir"] = err
	}

	return errs
}

// NormalizeWriter replaces invalid WriterConfig fields with defaults and reports each correction.
func NormalizeWriter(cfg *config.WriterConfig) []Warning {
	def := config.DefaultWriterConfig()
	var warns []Warning

	fixDurationMS("MergeFlushInterval", &cfg.MergeFlushInterval,
		100*time.Millisecond, 5*time.Minute, def.MergeFlushInterval, &warns)

	fixInt("ChanSize", &cfg.ChanSize, 1, 1_000_000, def.ChanSize, &warns)
	fixInt("BatchSize", &cfg.BatchSize, 1, 1_000_000, def.BatchSize, &warns)
	fixDirectoryName("ResultBaseDir", &cfg.ResultBaseDir, def.ResultBaseDir, &warns)

	return warns
}
