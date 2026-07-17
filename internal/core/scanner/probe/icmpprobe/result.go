package icmpprobe

import (
	"fmt"
	"strconv"
	"time"

	"bgscan/internal/core/result"
)

// Schema defines the result schema for ICMP probes.
var Schema = result.ResultSchema{
	Name:      "ICMP",
	Directory: "icmp",

	Columns: []result.ColumnDef{
		{
			Name:  "IP",
			Width: 60,
		},
		{
			Name:  "Latency",
			Width: 20,
		},
		{
			Name:  "Tries",
			Width: 10,
		},
		{
			Name:  "Mode",
			Width: 10,
		},
	},

	Parser: parseICMPResult,
}

// ICMPResult holds the outcome of a single ICMP echo probe.
type ICMPResult struct {
	IP      string
	Latency time.Duration
	Tries   int    // number of attempts before success
	Mode    string // "raw" or "udp"
}

// Key returns the IP address as the unique identifier for the result.
func (r ICMPResult) Key() string {
	return r.IP
}

// KeyType returns the type of key used for this result.
func (r ICMPResult) KeyType() result.KeyType {
	return result.KeyIP
}

// Equal checks if the given result matches this result's IP address.
func (r ICMPResult) Equal(rs result.Result) bool {
	return r.IP == rs.Key()
}

// ToRecord converts the ICMPResult into a slice of strings for serialization.
func (r ICMPResult) ToRecord() []string {
	tries := 1
	if r.Tries != 0 {
		tries = r.Tries
	}
	return []string{
		r.IP,
		r.Latency.String(),
		strconv.Itoa(tries),
		r.Mode,
	}
}

// Score returns a latency-based score where lower latency yields a higher score.
func (r ICMPResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())
	if ms < 1 {
		ms = 1
	}
	return 1000.0 / ms
}

func parseICMPResult(record []string) (result.Result, error) {
	if len(record) < 2 {
		return nil, fmt.Errorf(
			"invalid ICMP result record: expected at least 2 fields, got %d",
			len(record),
		)
	}

	latency, err := time.ParseDuration(record[1])
	if err != nil {
		return nil, fmt.Errorf("parse latency: %w", err)
	}

	// Backward compatibility: old records had only IP + Latency.
	var tries int
	var mode string
	if len(record) >= 4 {
		if _, err := fmt.Sscanf(record[2], "%d", &tries); err != nil {
			return nil, fmt.Errorf("parse tries: %w", err)
		}
		mode = record[3]
	}

	return ICMPResult{
		IP:      record[0],
		Latency: max(time.Millisecond, latency.Round(time.Millisecond)),
		Tries:   max(1, tries),
		Mode:    mode,
	}, nil
}
