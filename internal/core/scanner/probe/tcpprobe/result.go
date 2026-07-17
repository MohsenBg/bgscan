package tcpprobe

import (
	"fmt"
	"strconv"
	"time"

	"bgscan/internal/core/result"
)

// Schema is the result schema emitted by the TCP probe.
var Schema = result.ResultSchema{
	Name:      "TCP",
	Directory: "tcp",

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
			Name:  "Port",
			Width: 10,
		},
		{
			Name:  "Tries",
			Width: 10,
		},
	},

	Parser: parseTCPResult,
}

// TCPResult holds the outcome of a single TCP handshake probe.
type TCPResult struct {
	IP      string
	Port    uint16
	Latency time.Duration
	Tries   int
}

// Key returns the IP used to deduplicate results.
func (r TCPResult) Key() string {
	return r.IP
}

// KeyType returns the result key type.
func (r TCPResult) KeyType() result.KeyType {
	return result.KeyIP
}

// Equal reports whether r and rs identify the same IP.
func (r TCPResult) Equal(rs result.Result) bool {
	return r.IP == rs.Key()
}

// ToRecord serializes the result into a CSV-style string slice.
func (r TCPResult) ToRecord() []string {
	port := "-"
	if r.Port != 0 {
		port = strconv.FormatUint(uint64(r.Port), 10)
	}

	tries := 1
	if r.Tries != 0 {
		tries = r.Tries
	}

	return []string{
		r.IP,
		r.Latency.String(),
		port,
		strconv.Itoa(tries),
	}
}

// Score returns a latency-based score where lower latency yields a higher score.
func (r TCPResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())
	if ms < 1 {
		ms = 1
	}
	return 1000.0 / ms
}

func parseTCPResult(record []string) (result.Result, error) {
	if len(record) < 3 {
		return nil, fmt.Errorf(
			"invalid TCP result record: expected at least 3 fields, got %d",
			len(record),
		)
	}

	latency, err := time.ParseDuration(record[1])
	if err != nil {
		return nil, fmt.Errorf("parse latency: %w", err)
	}

	port, err := strconv.ParseUint(record[2], 10, 16)
	if err != nil {
		port = 0
	}

	// Backward compatibility: old records had no Tries field.
	var tries int
	if len(record) >= 4 {
		if _, err := fmt.Sscanf(record[3], "%d", &tries); err != nil {
			return nil, fmt.Errorf("parse tries: %w", err)
		}
	}

	return TCPResult{
		IP:      record[0],
		Latency: max(time.Millisecond, latency.Round(time.Millisecond)),
		Port:    uint16(port),
		Tries:   max(1, tries),
	}, nil
}
