package stormdnsprobe

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

// Schema defines the database layout and parsing rules for StormDNS
// probe outcomes.
var Schema = result.ResultSchema{
	Name:      "StormDNS",
	Directory: "stormdns",

	Columns: []result.ColumnDef{
		{Name: "Target", Width: 35, ColumnType: result.ColumnHost},
		{Name: "Latency", Width: 20, ColumnType: result.ColumnDuration},
		{Name: "Port", Width: 10},
		{Name: "Query", Width: 15},
		{Name: "Enc", Width: 15},
	},

	Parser: parseStormDNSResult,
}

// StormDNSResult holds the outcome of a single StormDNS tunnel probe.
//
// Latency measures only the proxy validation phase, excluding tunnel
// startup overhead, to reflect sustained tunnel performance.
type StormDNSResult struct {
	Entry     target.Entry // The row that produced this result.
	Latency   time.Duration
	Port      uint16
	QueryType string        // Tunnel DNS query type: TXT, NS, CNAME or ROTATE.
	Enc       dns.EncMethod // Data encryption method of the tunnel.
}

// Key returns the identity of the probed target with probe index prefix.
func (r StormDNSResult) Key() string { return r.Entry.Key() }

func (r StormDNSResult) KeyType() target.Kind {
	return target.KindIP
}

func (r StormDNSResult) Equal(rs result.Result) bool {
	return r.Key() == rs.Key()
}

func (r StormDNSResult) ToRecord() []string {
	return append(r.Entry.EncodeCSV(),
		result.FormatDuration(r.Latency),
		fmt.Sprintf("%d", r.Port),
		normalizeQueryType(r.QueryType),
		r.Enc.String(),
	)
}

// Score rates lower latency higher; sub-millisecond values clamp to 1ms to
// avoid dividing by zero.
func (r StormDNSResult) Score() float64 {
	return 1000 / float64(max(1, r.Latency.Milliseconds()))
}

func parseStormDNSResult(record []string) (result.Result, error) {
	const entryFields = 4
	const resultFields = 4

	if len(record) < entryFields+resultFields {
		return nil, fmt.Errorf(
			"invalid StormDNS result record: expected at least %d fields, got %d",
			entryFields+resultFields,
			len(record),
		)
	}

	entry, err := target.ParseRecord(record[:entryFields])
	if err != nil {
		return nil, fmt.Errorf("parse target: %w", err)
	}

	latencyRaw, portRaw, queryRaw, encRaw := record[entryFields], record[entryFields+1], record[entryFields+2], record[entryFields+3]

	latency, err := time.ParseDuration(latencyRaw)
	if err != nil {
		return nil, fmt.Errorf("parse latency: %w", err)
	}

	port, _ := strconv.ParseUint(portRaw, 10, 16)

	return StormDNSResult{
		Entry:     entry,
		Latency:   latency,
		Port:      uint16(port),
		QueryType: normalizeQueryType(queryRaw),
		Enc:       dns.ParseEncMethod(encRaw),
	}, nil
}
