package xrayprobe

import (
	"fmt"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/speedtest"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

// Schema defines the output format and parsing logic for Xray probe results.
var Schema = result.ResultSchema{
	Name:      "Xray",
	Directory: "xray",

	Columns: []result.ColumnDef{
		{Name: "Target", Width: 40, ColumnType: result.ColumnHost},
		{Name: "Latency", Width: 20, ColumnType: result.ColumnDuration},
		{Name: "Download", Width: 20, ColumnType: result.ColumnBandwidth},
		{Name: "Upload", Width: 20, ColumnType: result.ColumnBandwidth},
	},

	Parser: parseXrayResult,
}

// XrayResult holds the outcome of a single Xray probe.
// Download and Upload are zero when the corresponding test mode is disabled.
type XrayResult struct {
	Entry    target.Entry // The row that produced this result.
	Latency  time.Duration
	Download speedtest.BitsPerSec
	Upload   speedtest.BitsPerSec
}

// Key returns the identity of the probed target with probe index prefix.
func (r XrayResult) Key() string { return r.Entry.Key() }

// KeyType returns the kind of the probed target.
func (r XrayResult) KeyType() target.Kind { return r.Entry.Kind }

// Equal reports whether r and other point at the same target.
func (r XrayResult) Equal(rs result.Result) bool { return r.Key() == rs.Key() }

// ToRecord serializes the result into a string slice for tabular output.
// The target comes first, in value,port,path,enable form; the measurements
// follow.
func (r XrayResult) ToRecord() []string {
	return append(r.Entry.EncodeCSV(),
		result.FormatDuration(r.Latency),
		r.Download.String(),
		r.Upload.String(),
	)
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
	latencyScore := 1000.0 / float64(max(1, r.Latency.Milliseconds()))
	downloadScore := float64(r.Download) / float64(speedtest.Mbps)
	uploadScore := float64(r.Upload) / float64(speedtest.Mbps)

	return latencyScore*0.1 + downloadScore*0.6 + uploadScore*0.3
}

// parseXrayResult reconstructs an XrayResult from a string slice.
// It enforces a minimum latency of 1ms and silently defaults download/upload
// speeds to zero if parsing fails.
func parseXrayResult(record []string) (result.Result, error) {
	// Record layout: the target's CSV fields first (value, port, path,
	// enable, must match len(target.Entry.EncodeCSV())), then the
	// result's own fields (latency, download, upload).
	const entryFields = 4
	const resultFields = 3

	if len(record) < entryFields+resultFields {
		return nil, fmt.Errorf(
			"invalid Xray result record: expected at least %d fields, got %d",
			entryFields+resultFields,
			len(record),
		)
	}

	entry, err := target.ParseRecord(record[:entryFields])
	if err != nil {
		return nil, fmt.Errorf("parse target: %w", err)
	}

	latencyRaw, downloadRaw, uploadRaw := record[entryFields], record[entryFields+1], record[entryFields+2]

	latency, err := time.ParseDuration(latencyRaw)
	if err != nil {
		return nil, fmt.Errorf("parse latency: %w", err)
	}

	// Speeds are optional (zero when the test mode is disabled), so a
	// value that fails to parse defaults to zero instead of failing
	// the whole record.
	download, _ := speedtest.ParseBitsPerSec(downloadRaw)
	upload, _ := speedtest.ParseBitsPerSec(uploadRaw)

	return XrayResult{
		Entry:    entry,
		Latency:  max(time.Millisecond, latency.Round(time.Millisecond)),
		Download: download,
		Upload:   upload,
	}, nil
}
