package httpprobe

import (
	"fmt"
	"net/netip"
	"strconv"
	"time"

	"bgscan/internal/core/result"
)

// Schema defines the output format and parsing logic for HTTP probe results.
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

// HTTPResult represents the outcome of a single HTTP probe.
type HTTPResult struct {
	IP          netip.Addr
	StatusCode  int    // 0 if the request failed before a response was received.
	HTTPVersion string // e.g., "HTTP/1.1" or "HTTP/2.0".
	UseTLS      bool   // True if the connection was established over TLS.
	Latency     time.Duration
}

// Key returns the IP address string used for result deduplication.
func (r HTTPResult) Key() string {
	return r.IP.String()
}

// KeyType implements result.Result, identifying this as an IP-based key.
func (r HTTPResult) KeyType() result.KeyType {
	return result.KeyIP
}

// Equal reports whether r and other represent the same target IP.
func (r HTTPResult) Equal(rs result.Result) bool {
	return r.IP.String() == rs.Key()
}

// ToRecord serializes the result into a string slice for tabular output.
// Failed requests (StatusCode 0) render as "-" for status and TLS.
func (r HTTPResult) ToRecord() []string {
	status := "-"
	useTLS := "-"
	if r.StatusCode != 0 {
		status = strconv.Itoa(r.StatusCode)
		useTLS = strconv.FormatBool(r.UseTLS)
	}

	return []string{
		r.IP.String(),
		result.FormatDuration(r.Latency),
		status,
		r.HTTPVersion,
		useTLS,
	}
}

// Score calculates a performance metric where lower latency yields a higher score.
// It guards against division by zero by enforcing a minimum latency of 1ms.
func (r HTTPResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())

	if ms < 1 {
		ms = 1
	}

	score := 1000.0 / ms

	return score
}

// parseHTTPResult reconstructs an HTTPResult from a string slice.
// It enforces a minimum latency of 1ms and supports backward compatibility
// for legacy records containing only IP and latency.
func parseHTTPResult(record []string) (result.Result, error) {
	if len(record) < 2 {
		return nil, fmt.Errorf("invalid HTTP result record: expected 6 fields, got %d", len(record))
	}

	latency, err := time.ParseDuration(record[1])
	if err != nil {
		return nil, fmt.Errorf("parse latency: %w", err)
	}

	ip, err := netip.ParseAddr(record[0])
	if err != nil {
		return nil, fmt.Errorf("parse IP: %w", err)
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
		IP:      ip,
		Latency: max(time.Millisecond, latency.Round(time.Millisecond)),

		StatusCode:  status,
		HTTPVersion: httpVersion,
		UseTLS:      tls,
	}, nil
}
