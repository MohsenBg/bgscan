package masterdnsprobe

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

func joinConfigErrors(errs map[dns.Field]error) error {
	var joined error
	for field, err := range errs {
		joined = errors.Join(joined, fmt.Errorf("%s: %w", field, err))
	}
	return joined
}

// Schema defines the database layout and parsing rules for MasterDNS
// probe outcomes.
var Schema = result.ResultSchema{
	Name:      "MasterDNS",
	Directory: "masterdns",

	Columns: []result.ColumnDef{
		{Name: "Target", Width: 35, ColumnType: result.ColumnHost},
		{Name: "Latency", Width: 15, ColumnType: result.ColumnDuration},
		{Name: "Port", Width: 10},
		{Name: "Enc", Width: 15},
	},

	Parser: parseMasterDNSResult,
}

// MasterDNSResult holds the outcome of a single MasterDNS tunnel probe.
//
// Latency measures only the proxy validation phase, excluding tunnel
// startup overhead, to reflect sustained tunnel performance.
type MasterDNSResult struct {
	Entry   target.Entry // The row that produced this result.
	Latency time.Duration
	Port    uint16        // Local SOCKS5 port allocated for this run.
	Enc     dns.EncMethod // Data encryption method of the tunnel.
}

// Key returns the identity of the probed target with probe index prefix.
func (r MasterDNSResult) Key() string { return r.Entry.Key() }

func (r MasterDNSResult) KeyType() target.Kind {
	return r.Entry.Kind
}

func (r MasterDNSResult) Equal(rs result.Result) bool {
	return r.Key() == rs.Key()
}

func (r MasterDNSResult) ToRecord() []string {
	return append(r.Entry.EncodeCSV(),
		result.FormatDuration(r.Latency),
		fmt.Sprintf("%d", r.Port),
		r.Enc.String(),
	)
}

// Score rates lower latency higher; sub-millisecond values clamp to 1ms to
// avoid dividing by zero.
func (r MasterDNSResult) Score() float64 {
	return 1000 / float64(max(1, r.Latency.Milliseconds()))
}

func parseMasterDNSResult(record []string) (result.Result, error) {
	const entryFields = 4
	const resultFields = 3

	if len(record) < entryFields+resultFields {
		return nil, fmt.Errorf(
			"invalid MasterDNS result record: expected at least %d fields, got %d",
			entryFields+resultFields,
			len(record),
		)
	}

	entry, err := target.ParseRecord(record[:entryFields])
	if err != nil {
		return nil, fmt.Errorf("parse target: %w", err)
	}

	latencyRaw, portRaw, encRaw := record[entryFields], record[entryFields+1], record[entryFields+2]

	latency, err := time.ParseDuration(latencyRaw)
	if err != nil {
		return nil, fmt.Errorf("parse latency: %w", err)
	}

	port, err := strconv.ParseUint(portRaw, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("parse port: %w", err)
	}

	return MasterDNSResult{
		Entry:   entry,
		Latency: latency,
		Port:    uint16(port),
		Enc:     dns.ParseEncMethod(encRaw),
	}, nil
}
