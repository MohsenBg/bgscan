package icmpprobe

import (
	"fmt"
	"strconv"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

// Schema defines the result schema for ICMP probes.
var Schema = result.ResultSchema{
	Name:      "ICMP",
	Directory: "icmp",
	Columns: []result.ColumnDef{
		{Name: "Target", Width: 60, ColumnType: result.ColumnHost},
		{Name: "Latency", Width: 20, ColumnType: result.ColumnDuration},
		{Name: "Tries", Width: 10},
		{Name: "Mode", Width: 10},
	},
	Parser: parseICMPResult,
}

// ICMPResult holds the outcome of a single ICMP echo probe.
type ICMPResult struct {
	Entry   target.Entry
	Latency time.Duration
	Tries   int
	Mode    string // "raw" or "udp"
}

// Key returns the identity of the probed target with probe index prefix.
func (r ICMPResult) Key() string { return r.Entry.Key() }

// KeyType returns the key type used to identify ICMP results.
func (r ICMPResult) KeyType() target.Kind {
	return r.Entry.Kind
}

// Equal reports whether both results point at the same target.
func (r ICMPResult) Equal(rs result.Result) bool {
	return r.Key() == rs.Key()
}

// ToRecord converts the result into a slice of strings for serialization.
func (r ICMPResult) ToRecord() []string {
	return append(r.Entry.EncodeCSV(),
		result.FormatDuration(r.Latency),
		strconv.Itoa(max(1, r.Tries)),
		r.Mode,
	)
}

// Score returns a latency-based score where lower latency yields a higher score.
func (r ICMPResult) Score() float64 {
	ms := max(1, r.Latency.Milliseconds())
	return 1000 / float64(ms)
}

func parseICMPResult(record []string) (result.Result, error) {
	const entryFields = 4
	const resultFields = 3

	if len(record) < entryFields+resultFields {
		return nil, fmt.Errorf(
			"invalid ICMP result record: expected at least %d fields, got %d",
			entryFields+resultFields,
			len(record),
		)
	}

	entry, err := target.ParseRecord(record[:entryFields])
	if err != nil {
		return nil, err
	}

	fields := record[entryFields:]

	latency, err := time.ParseDuration(fields[0])
	if err != nil {
		return nil, fmt.Errorf("parse latency: %w", err)
	}

	tries, err := strconv.Atoi(fields[1])
	if err != nil {
		return nil, fmt.Errorf("parse tries: %w", err)
	}

	return ICMPResult{
		Entry:   entry,
		Latency: max(time.Millisecond, latency.Round(time.Millisecond)),
		Tries:   max(1, tries),
		Mode:    fields[2],
	}, nil
}
