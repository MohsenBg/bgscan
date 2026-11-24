package thefeedprobe

import (
	"fmt"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

// Schema defines the database layout and parsing rules for TheFeed probe
// outcomes.
var Schema = result.ResultSchema{
	Name:      "TheFeed",
	Directory: "thefeed",

	Columns: []result.ColumnDef{
		{Name: "Target", Width: 30, ColumnType: result.ColumnHost},
		{Name: "Latency", Width: 15, ColumnType: result.ColumnDuration},
		{Name: "Domain", Width: 35},
		{Name: "Mode", Width: 10},
	},

	Parser: parseTheFeedResult,
}

// TheFeedResult holds the outcome of a single TheFeed resolver probe.
//
// Latency measures only the DNS exchange: from the query send until the
// encrypted TXT response is decrypted successfully.
type TheFeedResult struct {
	Entry     target.Entry // The row that produced this result.
	Latency   time.Duration
	Domain    string
	QueryMode string
}

// Key returns the identity of the probed target with probe index prefix.
func (r TheFeedResult) Key() string { return r.Entry.Key() }

// KeyType returns the kind of the probed target.
func (r TheFeedResult) KeyType() target.Kind {
	return r.Entry.Kind
}

// Equal reports whether rs has the same key as r.
func (r TheFeedResult) Equal(rs result.Result) bool {
	return r.Key() == rs.Key()
}

// ToRecord converts the TheFeedResult into a slice of strings for serialization.
// The target comes first, in value,port,path,enable form; the measurements
// follow.
func (r TheFeedResult) ToRecord() []string {
	return append(r.Entry.EncodeCSV(),
		result.FormatDuration(r.Latency),
		r.Domain,
		r.QueryMode,
	)
}

// Score rates lower latency higher; sub-millisecond values clamp to 1ms to
// avoid dividing by zero.
func (r TheFeedResult) Score() float64 {
	return 1000 / float64(max(1, r.Latency.Milliseconds()))
}

func parseTheFeedResult(record []string) (result.Result, error) {
	const entryFields = 4
	const resultFields = 3

	if len(record) < entryFields+resultFields {
		return nil, fmt.Errorf(
			"invalid TheFeed result record: expected at least %d fields, got %d",
			entryFields+resultFields,
			len(record),
		)
	}

	entry, err := target.ParseRecord(record[:entryFields])
	if err != nil {
		return nil, fmt.Errorf("parse target: %w", err)
	}

	latencyRaw, domain, mode := record[entryFields], record[entryFields+1], record[entryFields+2]

	latency, err := time.ParseDuration(latencyRaw)
	if err != nil {
		return nil, fmt.Errorf("parse latency: %w", err)
	}

	return TheFeedResult{
		Entry:     entry,
		Latency:   latency,
		Domain:    domain,
		QueryMode: mode,
	}, nil
}
