package tcpprobe

import (
	"fmt"
	"strconv"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

// Schema defines the output format and parsing logic for TCP probe results.
var Schema = result.ResultSchema{
	Name:      "TCP",
	Directory: "tcp",

	Columns: []result.ColumnDef{
		{
			Name:       "Target",
			Width:      50,
			ColumnType: result.ColumnHost,
		},
		{
			Name:       "Latency",
			Width:      10,
			ColumnType: result.ColumnDuration,
		},
		{
			Name:  "Port",
			Width: 10,
		},
		{
			Name:  "Tries",
			Width: 10,
		},
		{
			Name:       "ResolvedFrom",
			Width:      20,
			ColumnType: result.ColumnText,
		},
	},

	Parser: parseTCPResult,
}

// TCPResult represents the outcome of a single TCP handshake probe.
type TCPResult struct {
	Entry        target.Entry // The row that produced this result.
	Port         uint16       // 0 if the connection failed before a port could be established.
	Latency      time.Duration
	ResolvedFrom string
	Tries        int // Number of connection attempts made.
}

// Key returns the identity of the probed target with probe index prefix.
func (r TCPResult) Key() string { return r.Entry.Key() }

// KeyType implements result.Result, identifying this as an IP-based key.
func (r TCPResult) KeyType() target.Kind {
	return r.Entry.Kind
}

// Equal reports whether r and other point at the same target.
func (r TCPResult) Equal(rs result.Result) bool {
	return r.Key() == rs.Key()
}

// ToRecord serializes the result into a string slice for tabular output.
// The target comes first, in value,port,path,enable form; the measurements
// follow. Port renders as "-" if 0, and tries defaults to 1 if 0.
func (r TCPResult) ToRecord() []string {
	port := "-"
	if r.Port != 0 {
		port = strconv.FormatUint(uint64(r.Port), 10)
	}

	tries := 1
	if r.Tries != 0 {
		tries = r.Tries
	}

	return append(r.Entry.EncodeCSV(),
		result.FormatDuration(r.Latency),
		port,
		strconv.Itoa(tries),
	)
}

// Score rates lower latency higher; sub-millisecond values clamp to 1ms to
// avoid dividing by zero.
func (r TCPResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())
	if ms < 1 {
		ms = 1
	}
	return 1000.0 / ms
}

func parseTCPResult(record []string) (result.Result, error) {
	const entryFields = 4
	const resultFields = 3

	if len(record) < entryFields+resultFields {
		return nil, fmt.Errorf(
			"invalid TCP result record: expected at least %d fields, got %d",
			entryFields+resultFields,
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

	var port uint16
	if fields[1] != "-" && fields[1] != "" {
		p, err := strconv.ParseUint(fields[1], 10, 16)
		if err != nil {
			return nil, fmt.Errorf("parse port: %w", err)
		}
		port = uint16(p)
	}

	tries, err := strconv.Atoi(fields[2])
	if err != nil {
		return nil, fmt.Errorf("parse tries: %w", err)
	}

	// ResolvedFrom is optional: older records end after tries.
	resolvedFrom := ""
	if len(fields) > 3 {
		resolvedFrom = fields[3]
	}

	return TCPResult{
		Entry:        entry,
		Port:         port,
		Latency:      max(time.Millisecond, latency.Round(time.Millisecond)),
		ResolvedFrom: resolvedFrom,
		Tries:        max(1, tries),
	}, nil
}
