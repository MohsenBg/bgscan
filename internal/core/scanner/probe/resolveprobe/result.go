package resolveprobe

import (
	"fmt"
	"strconv"
	"time"

	"bgscan/internal/core/result"
)

// Schema is the result schema emitted by the DNS resolver probe.
var Schema = result.ResultSchema{
	Name:      "DNSResolver",
	Directory: "dns_resolver",

	Columns: []result.ColumnDef{
		{Name: "IP", Width: 35},
		{Name: "Latency", Width: 15},
		{Name: "Record Type", Width: 12},
		{Name: "Tries", Width: 8},
		{Name: "Rcode", Width: 8},
		{Name: "DPI Check", Width: 12},
	},

	Parser: parseResolverResult,
}

// ResolverResult holds the outcome of a single DNS resolver probe.
type ResolverResult struct {
	IP         string
	Latency    time.Duration
	RecordType string
	Tries      int
	Rcode      uint16
	DPIChecked bool
}

// Key returns the IP used to deduplicate results.
func (r ResolverResult) Key() string {
	return r.IP
}

// KeyType returns the result key type.
func (r ResolverResult) KeyType() result.KeyType {
	return result.KeyIP
}

// Equal reports whether r and rs identify the same IP.
func (r ResolverResult) Equal(rs result.Result) bool {
	return r.IP == rs.Key()
}

// ToRecord serializes the result into a CSV-style string slice.
func (r ResolverResult) ToRecord() []string {
	dpi := "skipped"
	if r.DPIChecked {
		dpi = "passed"
	}

	return []string{
		r.IP,
		r.Latency.String(),
		r.RecordType,
		strconv.Itoa(r.Tries),
		strconv.FormatUint(uint64(r.Rcode), 10),
		dpi,
	}
}

// Score returns a latency-based score where lower latency yields a higher score.
func (r ResolverResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())
	if ms < 1 {
		ms = 1
	}
	return 1000.0 / ms
}

func parseResolverResult(record []string) (result.Result, error) {
	// Backward compatibility: old records only had IP + Latency.
	if len(record) == 2 {
		return parseResolverResultLegacy(record)
	}

	if len(record) < 6 {
		return nil, fmt.Errorf(
			"invalid DNS resolver result record: expected 6 fields, got %d",
			len(record),
		)
	}

	latency, err := time.ParseDuration(record[1])
	if err != nil {
		return nil, fmt.Errorf("parse latency: %w", err)
	}

	tries, err := strconv.Atoi(record[3])
	if err != nil {
		return nil, fmt.Errorf("parse tries: %w", err)
	}

	rcode, err := strconv.ParseUint(record[4], 10, 16)
	if err != nil {
		return nil, fmt.Errorf("parse rcode: %w", err)
	}

	return ResolverResult{
		IP:         record[0],
		Latency:    max(time.Millisecond, latency.Round(time.Millisecond)),
		RecordType: record[2],
		Tries:      max(1, tries),
		Rcode:      uint16(rcode),
		DPIChecked: record[5] == "passed",
	}, nil
}

func parseResolverResultLegacy(record []string) (result.Result, error) {
	latency, err := time.ParseDuration(record[1])
	if err != nil {
		return nil, fmt.Errorf("parse latency: %w", err)
	}

	return ResolverResult{
		IP:         record[0],
		Latency:    max(time.Millisecond, latency.Round(time.Millisecond)),
		RecordType: "?",
		Tries:      1,
		Rcode:      0,
		DPIChecked: false,
	}, nil
}
