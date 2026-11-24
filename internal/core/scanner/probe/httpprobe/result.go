package httpprobe

import (
	"fmt"
	"strconv"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

// Schema defines the output format and parsing logic for HTTP probe results.
var Schema = result.ResultSchema{
	Name:      "HTTP",
	Directory: "http",

	Columns: []result.ColumnDef{
		{
			Name:       "Target",
			Width:      39,
			ColumnType: result.ColumnHost,
		},
		{
			Name:       "Latency",
			Width:      19,
			ColumnType: result.ColumnDuration,
		},
		{
			Name:  "Status",
			Width: 12,
		},
		{
			Name:  "Version",
			Width: 20,
		},
		{
			Name:  "TLS",
			Width: 10,
		},
	},

	Parser: parseHTTPResult,
}

// HTTPResult represents the outcome of a single HTTP probe.
type HTTPResult struct {
	Entry       target.Entry // The row that produced this result.
	StatusCode  int          // 0 if the request failed before a response was received.
	HTTPVersion string       // e.g., "HTTP/1.1" or "HTTP/2.0".
	UseTLS      bool         // True if the connection was established over TLS.
	Latency     time.Duration
}

// Key returns the identity of the probed target with probe index prefix.
func (r HTTPResult) Key() string { return r.Entry.Key() }

// KeyType returns the kind of the entry that produced this result.
func (r HTTPResult) KeyType() target.Kind {
	return r.Entry.Kind
}

// Equal reports whether r and other point at the same target.
func (r HTTPResult) Equal(rs result.Result) bool {
	return r.Key() == rs.Key()
}

// ToRecord serializes the result into a string slice for tabular output.
// The target comes first, in value,port,path,enable form; the measurements
// follow. Failed requests (StatusCode 0) render as "-" for status and TLS.
func (r HTTPResult) ToRecord() []string {
	status := "-"
	useTLS := "-"
	if r.StatusCode != 0 {
		status = strconv.Itoa(r.StatusCode)
		useTLS = strconv.FormatBool(r.UseTLS)
	}

	return append(r.Entry.EncodeCSV(),
		result.FormatDuration(r.Latency),
		status,
		r.HTTPVersion,
		useTLS,
	)
}

// Score rates lower latency higher; sub-millisecond values clamp to 1ms to
// avoid dividing by zero.
func (r HTTPResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())

	if ms < 1 {
		ms = 1
	}

	score := 1000.0 / ms

	return score
}

// parseHTTPResult reconstructs an HTTPResult from a string slice.
func parseHTTPResult(record []string) (result.Result, error) {
	const (
		entryFields  = 4
		resultFields = 4
		minFields    = entryFields + resultFields
	)

	if len(record) < minFields {
		return nil, fmt.Errorf(
			"invalid HTTP result record: expected at least %d fields, got %d",
			minFields,
			len(record),
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

	// Failed requests are written as "-" for both status and TLS.
	statusCode := 0
	if fields[1] != "-" {
		statusCode, err = strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("parse status code: %w", err)
		}
	}

	useTLS := false
	if fields[3] != "-" {
		useTLS, err = strconv.ParseBool(fields[3])
		if err != nil {
			return nil, fmt.Errorf("parse TLS: %w", err)
		}
	}

	return HTTPResult{
		Entry:       entry,
		Latency:     max(time.Millisecond, latency.Round(time.Millisecond)),
		StatusCode:  statusCode,
		HTTPVersion: fields[2],
		UseTLS:      useTLS,
	}, nil
}
