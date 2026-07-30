package xrayprobe

import (
	"fmt"
	"net/netip"
	"time"

	"bgscan/internal/core/result"
	"bgscan/internal/core/speedtest"
)

// Schema defines the output format and parsing logic for Xray probe results.
var Schema = result.ResultSchema{
	Name:      "Xray",
	Directory: "xray",

	Columns: []result.ColumnDef{
		{Name: "IP", Width: 40},
		{Name: "Latency", Width: 20},
		{Name: "Download", Width: 20},
		{Name: "Upload", Width: 20},
	},

	Parser: parseXrayResult,
}

// XrayResult represents the outcome of a single Xray probe.
// Download and Upload are zero when the corresponding test mode is disabled.
type XrayResult struct {
	IP       netip.Addr
	Latency  time.Duration
	Download speedtest.BitsPerSec
	Upload   speedtest.BitsPerSec
}

// Key returns the IP address string used for result deduplication.
func (r XrayResult) Key() string { return r.IP.String() }

// KeyType implements result.Result, identifying this as an IP-based key.
func (r XrayResult) KeyType() result.KeyType { return result.KeyIP }

// Equal reports whether r and other represent the same target IP.
func (r XrayResult) Equal(rs result.Result) bool { return r.IP.String() == rs.Key() }

// ToRecord serializes the result into a string slice for tabular output.
func (r XrayResult) ToRecord() []string {
	return []string{
		r.IP.String(),
		r.Latency.String(),
		r.Download.String(),
		r.Upload.String(),
	}
}

// Score calculates a weighted performance metric where lower latency and
// higher throughput yield a higher score.
//
// Weights:
//   - Latency:  0.1 (inverted as 1000/ms, minimum 1ms)
//   - Download: 0.6 (in Mbps)
//   - Upload:   0.3 (in Mbps)
//
// If a test mode is disabled, its speed is zero and contributes nothing,
// causing ConnectivityOnly results to sort purely by latency.
func (r XrayResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())
	if ms < 1 {
		ms = 1
	}

	latencyScore := 1000.0 / ms
	downloadScore := float64(r.Download) / float64(speedtest.Mbps)
	uploadScore := float64(r.Upload) / float64(speedtest.Mbps)

	return latencyScore*0.1 + downloadScore*0.6 + uploadScore*0.3
}

// parseXrayResult reconstructs an XrayResult from a string slice.
// It enforces a minimum latency of 1ms and silently defaults download/upload
// speeds to zero if parsing fails.
func parseXrayResult(record []string) (result.Result, error) {
	if len(record) < 4 {
		return nil, fmt.Errorf("invalid Xray result record: expected 4 fields, got %d", len(record))
	}
	ip, err := netip.ParseAddr(record[0])
	if err != nil {
		return nil, fmt.Errorf("parse IP: %w", err)
	}

	latency, err := time.ParseDuration(record[1])
	if err != nil {
		return nil, fmt.Errorf("parse latency: %w", err)
	}

	download, _ := speedtest.ParseBitsPerSec(record[2])
	upload, _ := speedtest.ParseBitsPerSec(record[3])

	return XrayResult{
		IP:       ip,
		Latency:  max(time.Millisecond, latency.Round(time.Millisecond)),
		Download: download,
		Upload:   upload,
	}, nil
}
