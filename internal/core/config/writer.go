package config

import "time"

// WriterConfig defines configuration for the result writer subsystem.
type WriterConfig struct {
	MergeFlushInterval DurationMS `toml:"merge_flush_interval"`
	ChanSize           int        `toml:"chan_size"`
	BatchSize          int        `toml:"batch_size"`
	ResultBaseDir      string     `toml:"result_directory"`
}

// DefaultWriterConfig returns the default configuration for the result writer.
func DefaultWriterConfig() WriterConfig {
	return WriterConfig{
		MergeFlushInterval: NewDurationMS(2 * time.Second),
		ChanSize:           1024,
		BatchSize:          4096,
		ResultBaseDir:      "result",
	}
}
