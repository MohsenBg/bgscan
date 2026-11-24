package vaydnsprobe

import (
	"fmt"
	"strconv"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

// Schema defines the database layout and parsing rules for VayDNS probe outcomes.
var Schema = result.ResultSchema{
	Name:      "VayDNS",
	Directory: "vaydns",

	Columns: []result.ColumnDef{
		{Name: "Target", Width: 40, ColumnType: result.ColumnHost},
		{Name: "Latency", Width: 15, ColumnType: result.ColumnDuration},
		{Name: "Port", Width: 10},
		{Name: "Transport", Width: 15},
		{Name: "Auth", Width: 15},
		{Name: "Proxy", Width: 15},
	},

	Parser: parseVayDNSResult,
}

// VayDNSResult holds the outcome of a single VayDNS tunnel probe.
//
// Latency measures only the proxy validation phase, from the first byte sent
// through the tunnel until a valid response is received. It excludes the initial
// tunnel startup overhead to reflect sustained tunnel performance.
type VayDNSResult struct {
	Entry             target.Entry // The row that produced this result.
	Latency           time.Duration
	Port              uint16                // Local SOCKS5 port allocated for this run.
	Transport         dns.ResolverType      // Underlying DNS transport used for the tunnel (e.g., UDP, DoH, DoT).
	AuthMethod        dns.AuthMethod        // How the tunnel authenticates.
	ResolverProxyType dns.ResolverProxyType // Proxy type used to reach the resolver.
}

// Key returns the identity of the probed target.
func (r VayDNSResult) Key() string { return r.Entry.Key() }

// KeyType returns the kind of the probed target.
func (r VayDNSResult) KeyType() target.Kind {
	return r.Entry.Kind
}

// Equal reports whether rs has the same key as r.
func (r VayDNSResult) Equal(rs result.Result) bool {
	return r.Key() == rs.Key()
}

// ToRecord converts the VayDNSResult into a slice of strings for storage.
// The target comes first, in value,port,path,enable form; the measurements
// follow.
func (r VayDNSResult) ToRecord() []string {
	return append(r.Entry.EncodeCSV(),
		result.FormatDuration(r.Latency),
		fmt.Sprintf("%d", r.Port),
		string(r.Transport),
		string(r.AuthMethod),
		string(r.ResolverProxyType),
	)
}

// Score rates lower latency higher; sub-millisecond values clamp to 1ms to
// avoid dividing by zero.
func (r VayDNSResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())
	if ms < 1 {
		ms = 1
	}
	return 1000.0 / ms
}

func parseVayDNSResult(record []string) (result.Result, error) {
	const (
		entryFields  = 4
		resultFields = 5
	)

	if len(record) < entryFields+resultFields {
		return nil, fmt.Errorf(
			"invalid VayDNS result record: expected at least %d fields, got %d",
			entryFields+resultFields, len(record),
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

	port, _ := strconv.ParseUint(fields[1], 10, 16)

	return VayDNSResult{
		Entry:             entry,
		Latency:           latency,
		Port:              uint16(port),
		Transport:         dns.ParseResolverType(fields[2]),
		AuthMethod:        dns.ParseAuthMethod(fields[3]),
		ResolverProxyType: dns.ParseResolverProxyType(fields[4]),
	}, nil
}
