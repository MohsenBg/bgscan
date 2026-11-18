package resolveprobe

import (
	"fmt"
	"strconv"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

// Schema defines the output format and parsing logic for DNS resolver probe results.
var Schema = result.ResultSchema{
	Name:      "DNSResolver",
	Directory: "dns_resolver",

	Columns: []result.ColumnDef{
		{Name: "Target", Width: 35, ColumnType: result.ColumnHost},
		{Name: "Latency", Width: 15, ColumnType: result.ColumnDuration},
		{Name: "Record Type", Width: 12},
		{Name: "Tries", Width: 8},
		{Name: "Rcode", Width: 8},
		{Name: "DPI Check", Width: 12},
	},

	Parser: parseResolverResult,
}

// ResolverResult represents the outcome of a single DNS resolver probe.
type ResolverResult struct {
	Entry      target.Entry // The row that produced this result.
	Latency    time.Duration
	RecordType string
	Tries      int
	Rcode      uint16
	DPIChecked bool
}

// Key returns the identity of the probed target with probe index prefix.
func (r ResolverResult) Key() string { return r.Entry.Key() }

// KeyType implements result.Result, identifying this as an IP-based key.
func (r ResolverResult) KeyType() target.Kind {
	return r.Entry.Kind
}

// Equal reports whether r and other point at the same target.
func (r ResolverResult) Equal(rs result.Result) bool {
	return r.Key() == rs.Key()
}

// ToRecord serializes the result into a string slice for tabular output.
// The target comes first, in value,port,path,enable form; the measurements
// follow.
func (r ResolverResult) ToRecord() []string {
	dpi := "skipped"
	if r.DPIChecked {
		dpi = "passed"
	}

	return append(r.Entry.EncodeCSV(),
		result.FormatDuration(r.Latency),
		r.RecordType,
		strconv.Itoa(r.Tries),
		strconv.FormatUint(uint64(r.Rcode), 10),
		dpi,
	)
}

// Score rates lower latency higher; sub-millisecond values clamp to 1ms to
// avoid dividing by zero.
func (r ResolverResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())
	if ms < 1 {
		ms = 1
	}
	return 1000.0 / ms
}

func parseResolverResult(record []string) (result.Result, error) {
	const (
		entryFields  = 4
		resultFields = 5
	)

	if len(record) < entryFields+resultFields {
		return nil, fmt.Errorf(
			"invalid Resolver result record: expected at least %d fields, got %d",
			entryFields+resultFields, len(record),
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

	tries, err := strconv.Atoi(fields[2])
	if err != nil {
		return nil, fmt.Errorf("parse tries: %w", err)
	}

	rcode, err := strconv.ParseUint(fields[3], 10, 16)
	if err != nil {
		return nil, fmt.Errorf("parse rcode: %w", err)
	}

	return ResolverResult{
		Entry:      entry,
		Latency:    max(time.Millisecond, latency.Round(time.Millisecond)),
		RecordType: fields[1],
		Tries:      max(1, tries),
		Rcode:      uint16(rcode),
		DPIChecked: fields[4] == "passed",
	}, nil
}
