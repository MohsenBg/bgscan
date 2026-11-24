package dnsttprobe

import (
	"fmt"
	"strconv"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/target"
)

// Schema defines the database layout and parsing rules for DNSTT probe outcomes.
var Schema = result.ResultSchema{
	Name:      "DNSTT",
	Directory: "dnstt",

	Columns: []result.ColumnDef{
		{Name: "Target", Width: 40, ColumnType: result.ColumnHost},
		{Name: "Latency", Width: 15, ColumnType: result.ColumnDuration},
		{Name: "Transport", Width: 15},
		{Name: "Port", Width: 10},
		{Name: "Auth", Width: 15},
		{Name: "Proxy", Width: 15},
	},

	Parser: parseDNSTTResult,
}

// DNSTTResult holds the outcome of a single DNSTT tunnel probe.
//
// Latency measures only the proxy validation phase, from the first
// byte sent through the tunnel until a valid response is received.
// It excludes the initial tunnel startup overhead.
type DNSTTResult struct {
	Entry             target.Entry // The row that produced this result.
	Latency           time.Duration
	Transport         dns.ResolverType // Underlying DNS transport used for the tunnel (e.g., UDP, DoH, DoT).
	Port              uint16
	AuthMethod        dns.AuthMethod        // How the tunnel authenticates.
	ResolverProxyType dns.ResolverProxyType // Proxy type used to reach the resolver.
}

// Key returns the identity of the probed target with probe index prefix.
func (r DNSTTResult) Key() string { return "0:" + r.Entry.Key() }

// KeyType returns the kind of the entry that produced this result.
func (r DNSTTResult) KeyType() target.Kind {
	return r.Entry.Kind
}

// Equal reports whether rs has the same key as r.
func (r DNSTTResult) Equal(rs result.Result) bool {
	return r.Key() == rs.Key()
}

// ToRecord returns the result as a record row. The target comes
// first, in value, port, path, enable form. The measurements follow.
func (r DNSTTResult) ToRecord() []string {
	return append(r.Entry.EncodeCSV(),
		result.FormatDuration(r.Latency),
		string(r.Transport),
		fmt.Sprintf("%d", r.Port),
		string(r.AuthMethod),
		string(r.ResolverProxyType),
	)
}

// Score rates lower latency higher; sub-millisecond values clamp to 1ms to
// avoid dividing by zero.
func (r DNSTTResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())
	if ms < 1 {
		ms = 1
	}
	return 1000.0 / ms
}

func parseDNSTTResult(record []string) (result.Result, error) {
	const (
		entryFields  = 4
		resultFields = 5
	)

	if len(record) < entryFields+resultFields {
		return nil, fmt.Errorf(
			"invalid DNSTT result record: expected at least %d fields, got %d",
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

	port, _ := strconv.ParseUint(fields[2], 10, 16)

	return DNSTTResult{
		Entry:             entry,
		Latency:           latency,
		Transport:         dns.ParseResolverType(fields[1]),
		Port:              uint16(port),
		AuthMethod:        dns.ParseAuthMethod(fields[3]),
		ResolverProxyType: dns.ParseResolverProxyType(fields[4]),
	}, nil
}
