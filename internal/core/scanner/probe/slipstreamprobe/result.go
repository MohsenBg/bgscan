package slipstreamprobe

import (
	"fmt"
	"strconv"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

// Schema defines the output format and parsing logic for Slipstream probe results.
var Schema = result.ResultSchema{
	Name:      "Slipstream",
	Directory: "slipstream",

	Columns: []result.ColumnDef{
		{Name: "Target", Width: 35, ColumnType: result.ColumnHost},
		{Name: "Latency", Width: 15, ColumnType: result.ColumnDuration},
		{Name: "Port", Width: 10},
		{Name: "Auth", Width: 20},
		{Name: "Proxy", Width: 20},
	},

	Parser: parseSlipstreamResult,
}

// SlipstreamResult represents the outcome of a single Slipstream tunnel probe.
type SlipstreamResult struct {
	Entry             target.Entry          // The row that produced this result.
	Latency           time.Duration         // Measures only the proxy validation phase, reflecting tunnel quality rather than startup overhead.
	Port              uint16                // Local SOCKS5 port allocated for this run.
	AuthMethod        dns.AuthMethod        // How the tunnel authenticates.
	ResolverProxyType dns.ResolverProxyType // Proxy type used to reach the resolver.
}

// Key returns the identity of the probed target with probe index prefix.
func (r SlipstreamResult) Key() string { return r.Entry.Key() }

// KeyType implements result.Result, identifying this as an IP-based key.
func (r SlipstreamResult) KeyType() target.Kind {
	return r.Entry.Kind
}

// Equal reports whether r and other point at the same target.
func (r SlipstreamResult) Equal(rs result.Result) bool {
	return r.Key() == rs.Key()
}

// ToRecord serializes the result into a string slice for tabular output.
// The target comes first, in value,port,path,enable form; the measurements
// follow.
func (r SlipstreamResult) ToRecord() []string {
	return append(r.Entry.EncodeCSV(),
		result.FormatDuration(r.Latency),
		fmt.Sprintf("%d", r.Port),
		string(r.AuthMethod),
		string(r.ResolverProxyType),
	)
}

// Score rates lower latency higher; sub-millisecond values clamp to 1ms to
// avoid dividing by zero.
func (r SlipstreamResult) Score() float64 {
	return 1000 / float64(max(1, r.Latency.Milliseconds()))
}

// parseSlipstreamResult reconstructs a SlipstreamResult from a string slice.
func parseSlipstreamResult(record []string) (result.Result, error) {
	const entryFields = 4
	const resultFields = 4

	if len(record) < entryFields+resultFields {
		return nil, fmt.Errorf(
			"invalid Slipstream result record: expected at least %d fields, got %d",
			entryFields+resultFields,
			len(record),
		)
	}

	entry, err := target.ParseRecord(record[:entryFields])
	if err != nil {
		return nil, fmt.Errorf("parse target: %w", err)
	}

	latencyRaw, portRaw, authRaw, proxyRaw := record[entryFields], record[entryFields+1], record[entryFields+2], record[entryFields+3]

	latency, err := time.ParseDuration(latencyRaw)
	if err != nil {
		return nil, fmt.Errorf("parse latency: %w", err)
	}

	port, _ := strconv.ParseUint(portRaw, 10, 16)

	return SlipstreamResult{
		Entry:             entry,
		Latency:           latency,
		Port:              uint16(port),
		AuthMethod:        dns.ParseAuthMethod(authRaw),
		ResolverProxyType: dns.ParseResolverProxyType(proxyRaw),
	}, nil
}
