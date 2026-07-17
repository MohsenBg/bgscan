package xrayprobe

import (
	"fmt"
	"time"

	"bgscan/internal/core/result"
	"bgscan/internal/core/speedtest"
)

// Schema is the result schema emitted by the Xray probe.
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

// XrayResult holds the outcome of a single Xray probe. Download and Upload
// are zero when the corresponding test mode was not enabled (ConnectivityOnly).
type XrayResult struct {
	IP       string
	Latency  time.Duration
	Download speedtest.BitsPerSec
	Upload   speedtest.BitsPerSec
}

// Key returns the IP used to deduplicate results.
func (r XrayResult) Key() string { return r.IP }

// KeyType returns the result key type.
func (r XrayResult) KeyType() result.KeyType { return result.KeyIP }

// Equal reports whether r and rs identify the same IP.
func (r XrayResult) Equal(rs result.Result) bool { return r.IP == rs.Key() }

// ToRecord serializes the result into a CSV-style string slice.
func (r XrayResult) ToRecord() []string {
	return []string{
		r.IP,
		r.Latency.String(),
		r.Download.String(),
		r.Upload.String(),
	}
}

// Score combines latency, download, and upload into a single comparable value.
// Weights:
//   - Latency:  0.1  (lower is better → inverted to 1000/ms)
//   - Download: 0.6  (higher is better → Mbps)
//   - Upload:   0.3  (higher is better → Mbps)
//
// Zero download or upload (test mode not enabled) contributes 0 to that
// component, so ConnectivityOnly results sort purely by latency.
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

func parseXrayResult(record []string) (result.Result, error) {
	if len(record) < 4 {
		return nil, fmt.Errorf("invalid Xray result record: expected 4 fields, got %d", len(record))
	}

	latency, err := time.ParseDuration(record[1])
	if err != nil {
		return nil, fmt.Errorf("parse latency: %w", err)
	}

	download, _ := speedtest.ParseBitsPerSec(record[2])
	upload, _ := speedtest.ParseBitsPerSec(record[3])

	return XrayResult{
		IP:       record[0],
		Latency:  max(time.Millisecond, latency.Round(time.Millisecond)),
		Download: download,
		Upload:   upload,
	}, nil
}
