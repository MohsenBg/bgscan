package tcpprobe

import (
	"fmt"
	"net/netip"
	"strconv"
	"time"

	"bgscan/internal/core/result"
)

// Schema defines the output format and parsing logic for TCP probe results.
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

// TCPResult represents the outcome of a single TCP handshake probe.
type TCPResult struct {
	IP      netip.Addr
	Port    uint16 // 0 if the connection failed before a port could be established.
	Latency time.Duration
	Tries   int // Number of connection attempts made.
}

// Key returns the IP address string used for result deduplication.
func (r TCPResult) Key() string {
	return r.IP.String()
}

// KeyType implements result.Result, identifying this as an IP-based key.
func (r TCPResult) KeyType() result.KeyType {
	return result.KeyIP
}

// Equal reports whether r and other represent the same target IP.
func (r TCPResult) Equal(rs result.Result) bool {
	return r.IP.String() == rs.Key()
}

// ToRecord serializes the result into a string slice for tabular output.
// It renders port as "-" if 0, and defaults tries to 1 if 0.
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
		r.IP.String(),
		result.FormatDuration(r.Latency),
		port,
		strconv.Itoa(tries),
	}
}

// Score calculates a performance metric where lower latency yields a higher score.
// It guards against division by zero by enforcing a minimum latency of 1ms.
func (r TCPResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())
	if ms < 1 {
		ms = 1
	}
	return 1000.0 / ms
}

// parseTCPResult reconstructs a TCPResult from a string slice.
// It enforces a minimum latency of 1ms and a minimum of 1 try,
// and provides backward compatibility for legacy records missing the Tries field.
func parseTCPResult(record []string) (result.Result, error) {
	if len(record) < 3 {
		return nil, fmt.Errorf(
			"invalid TCP result record: expected at least 3 fields, got %d",
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
		IP:      ip,
		Latency: max(time.Millisecond, latency.Round(time.Millisecond)),
		Port:    uint16(port),
		Tries:   max(1, tries),
	}, nil
}
