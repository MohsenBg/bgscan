package dnsttprobe

import (
	"fmt"
	"time"

	"bgscan/internal/core/dns"
	"bgscan/internal/core/result"
)

// Schema is the result schema for DNSTT probes, defining columns and the parsing function.
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
// Latency measures only the proxy validation phase — the time from first
// byte through the tunnel to a confirmed response. Tunnel startup is
// excluded so the value reflects tunnel quality, not startup overhead.
type DNSTTResult struct {
	IP        string
	Latency   time.Duration
	Transport dns.Transport // transport used for the tunnel (UDP, DoH, DoT, …)
	Port      uint16        // local SOCKS5 port allocated for this run
}

// Key returns the IP address as the unique identifier for this result.
func (r DNSTTResult) Key() string {
	return r.IP
}

// KeyType returns the type of key used for this result, which is KeyIP.
func (r DNSTTResult) KeyType() result.KeyType {
	return result.KeyIP
}

// Equal checks if the given result represents the same probe target by comparing IP addresses.
func (r DNSTTResult) Equal(rs result.Result) bool {
	return r.IP == rs.Key()
}

// ToRecord converts the result fields into a slice of strings for tabular output.
func (r DNSTTResult) ToRecord() []string {
	return []string{
		r.IP,
		r.Latency.String(),
		string(r.Transport),
		fmt.Sprintf("%d", r.Port),
	}
}

// Score returns a latency-based score where lower latency yields a higher score.
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

	latency, err := time.ParseDuration(record[1])
	if err != nil {
		return nil, fmt.Errorf("parse latency: %w", err)
	}

	// Backward compatibility: old records had only IP + Latency.
	var transport dns.Transport
	var port uint16
	if len(record) >= 4 {
		transport = dns.ParseTransport(record[2])
		if _, err := fmt.Sscanf(record[3], "%d", &port); err != nil {
			return nil, fmt.Errorf("parse port: %w", err)
		}
	}

	return DNSTTResult{
		IP:        record[0],
		Latency:   latency,
		Transport: transport,
		Port:      port,
	}, nil
}
