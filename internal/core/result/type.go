package result

import (
	"fmt"
	"time"
)

// KeyType identifies what kind of key is used to identify a result.
type KeyType uint8

const (
	KeyIP KeyType = iota
	KeyDomain
)

func (k KeyType) String() string {
	switch k {
	case KeyIP:
		return "ip"
	case KeyDomain:
		return "domain"
	default:
		return fmt.Sprintf("unknown(%d)", k)
	}
}

// Valid reports whether the KeyType is a defined constant.
func (k KeyType) Valid() bool {
	return k <= KeyDomain
}

// Result is the interface that every scan result must implement.
type Result interface {
	// Key returns the unique identifier for this result (IP or domain).
	Key() string
	// KeyType returns the type of key used to identify this result.
	KeyType() KeyType
	// ToRecord converts the result to a slice of strings for storage.
	// The order must match the ColumnDefs in the corresponding ResultSchema.
	ToRecord() []string
	// Equal reports whether the other result is equivalent.
	Equal(other Result) bool
	// Score returns a quality/confidence score between 0 and 1.
	Score() float64
}

// ResultParser converts stored record strings back into a Result.
type ResultParser func(record []string) (Result, error)

// ColumnDef describes a single column in a result table.
type ColumnDef struct {
	Name  string
	Width int // Display width for formatted output (0 = auto).
}

// ResultSchema describes the structure of a result type's storage.
type ResultSchema struct {
	Name      string
	Directory string
	Columns   []ColumnDef // Order matches ToRecord output.
	Parser    ResultParser
}

// Validate checks that the schema has all required fields.
func (s ResultSchema) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("schema name is required")
	}
	if s.Parser == nil {
		return fmt.Errorf("parser is required")
	}
	return nil
}

// Config controls asynchronous result writing behavior.
type Config struct {
	MergeFlushInterval time.Duration
	ChanSize           int
	BatchSize          int
}

const (
	DefaultChanSize       = 1024
	DefaultBatchSize      = 4096
	MinMergeFlushInterval = 120 * time.Millisecond
)

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		MergeFlushInterval: MinMergeFlushInterval,
		ChanSize:           DefaultChanSize,
		BatchSize:          DefaultBatchSize,
	}
}

// Validate checks the config and returns an error if any value is invalid.
func (c Config) Validate() error {
	if c.MergeFlushInterval < MinMergeFlushInterval {
		return fmt.Errorf("MergeFlushInterval must be >= %v, got %v",
			MinMergeFlushInterval, c.MergeFlushInterval)
	}
	if c.ChanSize <= 0 {
		return fmt.Errorf("ChanSize must be > 0, got %d", c.ChanSize)
	}
	if c.BatchSize <= 0 {
		return fmt.Errorf("BatchSize must be > 0, got %d", c.BatchSize)
	}
	return nil
}

// Normalize clamps invalid values to sensible defaults.
// Use Validate() if you want to detect errors instead of silently fixing them.
func (c *Config) Normalize() {
	if c.MergeFlushInterval < MinMergeFlushInterval {
		c.MergeFlushInterval = MinMergeFlushInterval
	}
	if c.ChanSize <= 0 {
		c.ChanSize = DefaultChanSize
	}
	if c.BatchSize <= 0 {
		c.BatchSize = DefaultBatchSize
	}
}

// ResultFile describes a stored result file on disk.
type ResultFile struct {
	Name        string
	SizeBytes   int64
	CreatedTime time.Time
	Schema      ResultSchema
	RecordCount uint64
	Path        string
}

// SizeString returns a human-readable file size (e.g., "1.5 MB").
func (f ResultFile) SizeString() string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case f.SizeBytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(f.SizeBytes)/float64(GB))
	case f.SizeBytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(f.SizeBytes)/float64(MB))
	case f.SizeBytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(f.SizeBytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", f.SizeBytes)
	}
}
