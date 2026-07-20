package config

// WriterConfig defines configuration for the result writer subsystem.
type WriterConfig struct {
	// MergeFlushInterval controls how often buffered results
	// are flushed and merged.
	MergeFlushInterval DurationMS `toml:"merge_flush_interval"`

	// ChanSize defines the size of the internal writer channel.
	ChanSize int `toml:"chan_size"`

	// BatchSize defines how many records are processed per batch.
	BatchSize int `toml:"batch_size"`

	ResultBaseDir string `toml:"result_directory"`
}
