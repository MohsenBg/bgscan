package slipstreamprobe

import (
	"fmt"
	"net/netip"
	"time"

	"bgscan/internal/core/result"
)

// Schema defines the output format and parsing logic for Slipstream probe results.
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

// SlipstreamResult represents the outcome of a single Slipstream tunnel probe.
type SlipstreamResult struct {
	IP      netip.Addr
	Latency time.Duration // Measures only the proxy validation phase, reflecting tunnel quality rather than startup overhead.
	Port    uint16        // Local SOCKS5 port allocated for this run.
}

// Key returns the IP address string used for result deduplication.
func (r SlipstreamResult) Key() string {
	return r.IP.String()
}

// KeyType implements result.Result, identifying this as an IP-based key.
func (r SlipstreamResult) KeyType() result.KeyType {
	return result.KeyIP
}

// Equal reports whether r and other represent the same target IP.
func (r SlipstreamResult) Equal(rs result.Result) bool {
	return r.IP.String() == rs.Key()
}

// ToRecord serializes the result into a string slice for tabular output.
func (r SlipstreamResult) ToRecord() []string {
	return []string{
		r.IP.String(),
		r.Latency.String(),
		fmt.Sprintf("%d", r.Port),
	}
}

// Score calculates a performance metric where lower latency yields a higher score.
// It guards against division by zero by enforcing a minimum latency of 1ms.
func (r SlipstreamResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())
	if ms < 1 {
		ms = 1
	}
	return 1000.0 / ms
}

// parseSlipstreamResult reconstructs a SlipstreamResult from a string slice.
// It provides backward compatibility for older records that only contain IP and latency.
func parseSlipstreamResult(record []string) (result.Result, error) {
	if len(record) < 2 {
		return nil, fmt.Errorf(
			"invalid Slipstream result record: expected at least 2 fields, got %d",
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

	// Backward compatibility: old records had only IP + Latency.
	var port uint16
	if len(record) >= 3 {
		if _, err := fmt.Sscanf(record[2], "%d", &port); err != nil {
			return nil, fmt.Errorf("parse port: %w", err)
		}
	}

	return SlipstreamResult{
		IP:      ip,
		Latency: latency,
		Port:    port,
	}, nil
}
