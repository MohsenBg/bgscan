package resolveprobe

import (
	"fmt"
	"net/netip"
	"strconv"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
)

// Schema defines the output format and parsing logic for DNS resolver probe results.
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

// ResolverResult represents the outcome of a single DNS resolver probe.
type ResolverResult struct {
	IP         netip.Addr
	Latency    time.Duration
	RecordType string
	Tries      int
	Rcode      uint16
	DPIChecked bool
}

// Key returns the IP address string used for result deduplication.
func (r ResolverResult) Key() string {
	return r.IP.String()
}

// KeyType implements result.Result, identifying this as an IP-based key.
func (r ResolverResult) KeyType() result.KeyType {
	return result.KeyIP
}

// Equal reports whether r and other represent the same resolver IP.
func (r ResolverResult) Equal(rs result.Result) bool {
	return r.IP.String() == rs.Key()
}

// ToRecord serializes the result into a string slice for CSV-style output.
func (r ResolverResult) ToRecord() []string {
	dpi := "skipped"
	if r.DPIChecked {
		dpi = "passed"
	}

	return []string{
		r.IP.String(),
		result.FormatDuration(r.Latency),
		r.RecordType,
		strconv.Itoa(r.Tries),
		strconv.FormatUint(uint64(r.Rcode), 10),
		dpi,
	}
}

// Score calculates a performance metric where lower latency yields a higher score.
// It guards against division by zero by enforcing a minimum latency of 1ms.
func (r ResolverResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())
	if ms < 1 {
		ms = 1
	}
	return 1000.0 / ms
}

// parseResolverResult reconstructs a ResolverResult from a string slice.
// It sanitizes parsed values and falls back to legacy parsing for older 2-field records.
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
	ip, err := netip.ParseAddr(record[0])
	if err != nil {
		return nil, fmt.Errorf("parse IP: %w", err)
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
		IP:         ip,
		Latency:    max(time.Millisecond, latency.Round(time.Millisecond)),
		RecordType: record[2],
		Tries:      max(1, tries),
		Rcode:      uint16(rcode),
		DPIChecked: record[5] == "passed",
	}, nil
}

// parseResolverResultLegacy handles older 2-field result records by filling
// missing fields with safe defaults.
func parseResolverResultLegacy(record []string) (result.Result, error) {
	ip, err := netip.ParseAddr(record[0])
	if err != nil {
		return nil, fmt.Errorf("parse IP: %w", err)
	}

	latency, err := time.ParseDuration(record[1])
	if err != nil {
		return nil, fmt.Errorf("parse latency: %w", err)
	}

	return ResolverResult{
		IP:         ip,
		Latency:    max(time.Millisecond, latency.Round(time.Millisecond)),
		RecordType: "?",
		Tries:      1,
		Rcode:      0,
		DPIChecked: false,
	}, nil
}
