package dnsttprobe

import (
	"fmt"
	"net/netip"
	"time"

	"bgscan/internal/core/dns"
	"bgscan/internal/core/result"
)

// Schema defines the database layout and parsing rules for DNSTT probe outcomes.
var Schema = result.ResultSchema{
	Name:      "DNSTT",
	Directory: "dnstt",

	Columns: []result.ColumnDef{
		{
			Name:  "IP",
			Width: 40,
		},
		{
			Name:  "Latency",
			Width: 20,
		},
		{
			Name:  "Transport",
			Width: 20,
		},
		{
			Name:  "Port",
			Width: 20,
		},
	},

	Parser: parseDNSTTResult,
}

// DNSTTResult holds the outcome of a single DNSTT tunnel probe.
//
// Latency measures only the proxy validation phase—from the first byte sent
// through the tunnel until a valid response is received. It excludes the initial
// tunnel startup overhead to reflect sustained tunnel performance.
type DNSTTResult struct {
	IP        netip.Addr
	Latency   time.Duration
	Transport dns.ResolverType // Underlying DNS transport used for the tunnel (e.g., UDP, DoH, DoT).
	Port      uint16           // Local SOCKS5 port allocated for validation.
}

func (r DNSTTResult) Key() string {
	return r.IP.String()
}

func (r DNSTTResult) KeyType() result.KeyType {
	return result.KeyIP
}

func (r DNSTTResult) Equal(rs result.Result) bool {
	return r.IP.String() == rs.Key()
}

func (r DNSTTResult) ToRecord() []string {
	return []string{
		r.IP.String(),
		r.Latency.String(),
		string(r.Transport),
		fmt.Sprintf("%d", r.Port),
	}
}

// Score calculates a performance rating where lower latency yields a higher score.
// A latency of 0 or negative values are clamped to 1ms to prevent division by zero.
func (r DNSTTResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())
	if ms < 1 {
		ms = 1
	}
	return 1000.0 / ms
}

func parseDNSTTResult(record []string) (result.Result, error) {
	if len(record) < 2 {
		return nil, fmt.Errorf(
			"invalid DNSTT result record: expected at least 2 fields, got %d",
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
	var transport dns.ResolverType
	var port uint16
	if len(record) >= 4 {
		transport = dns.ParseResolverType(record[2])
		if _, err := fmt.Sscanf(record[3], "%d", &port); err != nil {
			return nil, fmt.Errorf("parse port: %w", err)
		}
	}

	return DNSTTResult{
		IP:        ip,
		Latency:   latency,
		Transport: transport,
		Port:      port,
	}, nil
}
