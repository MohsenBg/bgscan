package masterdnsprobe

import (
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/result"
)

func joinConfigErrors(errs map[string]error) error {
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
		{Name: "IP", Width: 35},
		{Name: "Latency", Width: 15},
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
	IP      netip.Addr
	Latency time.Duration
	Port    uint16        // Local SOCKS5 port allocated for this run.
	Enc     dns.EncMethod // Data encryption method of the tunnel.
}

func (r MasterDNSResult) Key() string {
	return r.IP.String()
}

func (r MasterDNSResult) KeyType() result.KeyType {
	return result.KeyIP
}

func (r MasterDNSResult) Equal(rs result.Result) bool {
	return r.IP.String() == rs.Key()
}

func (r MasterDNSResult) ToRecord() []string {
	return []string{
		r.IP.String(),
		result.FormatDuration(r.Latency),
		fmt.Sprintf("%d", r.Port),
		r.Enc.String(),
	}
}

// Score calculates a performance rating where lower latency yields a
// higher score. Latency is clamped to 1ms to prevent division by zero.
func (r MasterDNSResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())
	if ms < 1 {
		ms = 1
	}
	return 1000.0 / ms
}

func parseMasterDNSResult(record []string) (result.Result, error) {
	if len(record) < 2 {
		return nil, fmt.Errorf(
			"invalid MasterDNS result record: expected at least 2 fields, got %d",
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

	// Legacy records contain only IP and Latency.
	var port uint16
	enc := dns.EncXOR

	if len(record) >= 3 {
		if _, err := fmt.Sscanf(record[2], "%d", &port); err != nil {
			return nil, fmt.Errorf("parse port: %w", err)
		}
	}

	if len(record) >= 4 {
		enc = dns.ParseEncMethod(record[3])
	}

	return MasterDNSResult{
		IP:      ip,
		Latency: latency,
		Port:    port,
		Enc:     enc,
	}, nil
}
