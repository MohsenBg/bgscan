package httpprobe

import (
	"fmt"
	"strconv"
	"time"

	"bgscan/internal/core/result"
)

// Schema defines the result schema for HTTP probes.
var Schema = result.ResultSchema{
	Name:      "HTTP",
	Directory: "http",

	Columns: []result.ColumnDef{
		{
			Name:  "IP",
			Width: 39,
		},
		{
			Name:  "Latency",
			Width: 19,
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

// HTTPResult holds the outcome of a single HTTP probe.
type HTTPResult struct {
	IP string

	StatusCode  int
	HTTPVersion string

	UseTLS bool

	Latency time.Duration
}

// Equal checks if the given result matches this result's IP address.
func (r HTTPResult) Equal(rs result.Result) bool {
	return r.IP == rs.Key()
}

// Key returns the IP address as the unique identifier for the result.
func (r HTTPResult) Key() string {
	return r.IP
}

// KeyType returns the type of key used for this result.
func (r HTTPResult) KeyType() result.KeyType {
	return result.KeyIP
}

// ToRecord converts the HTTPResult into a slice of strings for serialization.
func (r HTTPResult) ToRecord() []string {
	status := "-"
	useTLS := "-"
	if r.StatusCode != 0 {
		status = strconv.Itoa(r.StatusCode)
		useTLS = strconv.FormatBool(r.UseTLS)
	}

	return []string{
		r.IP,
		r.Latency.String(),
		status,
		r.HTTPVersion,
		useTLS,
	}
}

// Score returns a latency-based score where lower latency yields a higher score.
func (r HTTPResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())

	if ms < 1 {
		ms = 1
	}

	score := 1000.0 / ms

	return score
}

func parseHTTPResult(record []string) (result.Result, error) {
	if len(record) < 2 {
		return nil, fmt.Errorf("invalid HTTP result record: expected 6 fields, got %d", len(record))
	}

	latency, err := time.ParseDuration(record[1])
	if err != nil {
		return nil, fmt.Errorf("parse latency: %w", err)
	}

	status := 0
	tls := false
	httpVersion := "-"

	if len(record) > 4 {
		var err error

		status, err = strconv.Atoi(record[2])
		if err != nil {
			return nil, fmt.Errorf("parse status code: %w", err)
		}

		httpVersion = record[3]

		tls, err = strconv.ParseBool(record[4])
		if err != nil {
			return nil, fmt.Errorf("parse tls: %w", err)
		}

	}
	return HTTPResult{
		IP:      record[0],
		Latency: max(time.Millisecond, latency.Round(time.Millisecond)),

		StatusCode:  status,
		HTTPVersion: httpVersion,
		UseTLS:      tls,
	}, nil
}
