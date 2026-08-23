package validate

import (
	"time"

	"bgscan/internal/core/config"
)

const (
	// Merge flush interval limits
	MinWriterMergeFlushInterval = 100 * time.Millisecond
	MaxWriterMergeFlushInterval = 5 * time.Minute
)

const (
	// Channel size limits
	MinWriterChanSize = 1
	MaxWriterChanSize = 1_000_000

	// Batch size limits
	MinWriterBatchSize = 1
	MaxWriterBatchSize = 102_400
)

// ValidateWriter strictly validates a WriterConfig and returns errors by field name.
func ValidateWriter(cfg config.WriterConfig) map[string]error {
	errs := map[string]error{}

	if err := checkDuration(
		"MergeFlushInterval",
		cfg.MergeFlushInterval.Duration(),
		MinWriterMergeFlushInterval,
		MaxWriterMergeFlushInterval,
	); err != nil {
		errs["MergeFlushInterval"] = err
	}

	if err := checkInt(
		"ChanSize",
		cfg.ChanSize,
		MinWriterChanSize,
		MaxWriterChanSize,
	); err != nil {
		errs["ChanSize"] = err
	}

	if err := checkInt(
		"BatchSize",
		cfg.BatchSize,
		MinWriterBatchSize,
		MaxWriterBatchSize,
	); err != nil {
		errs["BatchSize"] = err
	}

	if err := checkDirectoryName(
		"ResultDirectory",
		cfg.ResultBaseDir,
	); err != nil {
		errs["ResultBaseDir"] = err
	}

	return errs
}

// NormalizeWriter replaces invalid WriterConfig fields with defaults and reports each correction.
func NormalizeWriter(cfg *config.WriterConfig) []Warning {
	def := config.DefaultWriterConfig()
	var warns []Warning

	fixDurationMS(
		"MergeFlushInterval",
		&cfg.MergeFlushInterval,
		MinWriterMergeFlushInterval,
		MaxWriterMergeFlushInterval,
		def.MergeFlushInterval,
		&warns,
	)

	fixInt(
		"ChanSize",
		&cfg.ChanSize,
		MinWriterChanSize,
		MaxWriterChanSize,
		def.ChanSize,
		&warns,
	)

	fixInt(
		"BatchSize",
		&cfg.BatchSize,
		MinWriterBatchSize,
		MaxWriterBatchSize,
		def.BatchSize,
		&warns,
	)

	fixDirectoryName(
		"ResultBaseDir",
		&cfg.ResultBaseDir,
		def.ResultBaseDir,
		&warns,
	)

	return warns
}
