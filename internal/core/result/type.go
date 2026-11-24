// Package result defines the interfaces, schemas, and persistence layer for scan results.
package result

import (
	"fmt"
	"math"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/target"
)

// Result is the interface that every scan result must implement.
type Result interface {
	// Key returns the unique identifier for this result (IP or domain).
	Key() string
	// KeyType returns the type of key used to identify this result.
	KeyType() target.Kind
	// ToRecord converts the result to storable strings, in ColumnDefs order.
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
	Name       string
	Width      int        // Display width for formatted output.
	ColumnType ColumnType // Semantic content type. The zero value is plain text.
}

// ColumnType classifies the semantic content of a column.
// The zero value is plain text.
type ColumnType uint8

const (
	// ColumnText marks plain, unstyled columns.
	ColumnText ColumnType = iota
	// ColumnHost marks the results' key column; copyable, never styled.
	ColumnHost
	// ColumnDuration marks latency measurements; styled by quality tier.
	ColumnDuration
	// ColumnBandwidth marks throughput measurements; reserved for tiering.
	ColumnBandwidth
)

// ResultSchema describes the structure of a result type's storage.
type ResultSchema struct {
	Name      string
	Directory string
	Columns   []ColumnDef
	Parser    ResultParser
}

// Validate checks that the schema has all required fields.
// Validate returns an error when Name is empty or Parser is nil.
func (s ResultSchema) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("schema name is required")
	}
	if s.Parser == nil {
		return fmt.Errorf("parser is required")
	}
	return nil
}

// ResultFile describes a stored result file on disk.
type ResultFile struct {
	Name        string
	SizeBytes   int64
	CreatedTime time.Time
	Schema      ResultSchema
	RecordCount uint64
	Path        string
	Kind        target.Kind
}

// SizeString returns a human-readable file size (e.g., "1.5 MB").
func (f ResultFile) SizeString() string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	round := func(v float64) float64 {
		return math.Round(v*100) / 100
	}

	switch {
	case f.SizeBytes >= GB:
		return fmt.Sprintf("%.2f GB", round(float64(f.SizeBytes)/float64(GB)))
	case f.SizeBytes >= MB:
		return fmt.Sprintf("%.2f MB", round(float64(f.SizeBytes)/float64(MB)))
	case f.SizeBytes >= KB:
		return fmt.Sprintf("%.2f KB", round(float64(f.SizeBytes)/float64(KB)))
	default:
		return fmt.Sprintf("%d B", f.SizeBytes)
	}
}
