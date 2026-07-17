package slipstreamprobe

import (
	"fmt"
	"time"

	"bgscan/internal/core/result"
)

// Schema is the result schema for Slipstream probes, defining columns and the parsing function.
var Schema = result.ResultSchema{
	Name:      "Slipstream",
	Directory: "slipstream",

	Columns: []result.ColumnDef{
		{
			Name:  "IP",
			Width: 45,
		},
		{
			Name:  "Latency",
			Width: 35,
		},
		{
			Name:  "Port",
			Width: 20,
		},
	},

	Parser: parseSlipstreamResult,
}

// SlipstreamResult holds the outcome of a single Slipstream tunnel probe.
//
// Latency measures only the proxy validation phase — after the tunnel is up —
// so it reflects tunnel quality rather than startup overhead.
type SlipstreamResult struct {
	IP      string
	Latency time.Duration
	Port    uint16 // local SOCKS5 port allocated for this run
}

// Key returns the IP address as the unique identifier for this result.
func (r SlipstreamResult) Key() string {
	return r.IP
}

// KeyType returns the type of key used for this result, which is KeyIP.
func (r SlipstreamResult) KeyType() result.KeyType {
	return result.KeyIP
}

// Equal checks if the given result represents the same probe target by comparing IP addresses.
func (r SlipstreamResult) Equal(rs result.Result) bool {
	return r.IP == rs.Key()
}

// ToRecord converts the result fields into a slice of strings for tabular output.
func (r SlipstreamResult) ToRecord() []string {
	return []string{
		r.IP,
		r.Latency.String(),
		fmt.Sprintf("%d", r.Port),
	}
}

// Score returns a latency-based score (lower latency → higher score).
func (r SlipstreamResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())
	if ms < 1 {
		ms = 1
	}
	return 1000.0 / ms
}

func parseSlipstreamResult(record []string) (result.Result, error) {
	if len(record) < 2 {
		return nil, fmt.Errorf(
			"invalid Slipstream result record: expected at least 2 fields, got %d",
			len(record),
		)
	}

	latency, err := time.ParseDuration(record[1])
	if err != nil {
		return nil, fmt.Errorf("parse latency: %w", err)
	}

	// Backward compatibility: old records had only IP + Latency.
	var port uint16
	if len(record) >= 3 {
		if _, err := fmt.Sscanf(record[2], "%d", &port); err != nil {
			return nil, fmt.Errorf("parse port: %w", err)
		}
	}

	return SlipstreamResult{
		IP:      record[0],
		Latency: latency,
		Port:    port,
	}, nil
}
