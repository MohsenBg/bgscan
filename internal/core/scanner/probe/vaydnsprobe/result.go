package vaydnsprobe

import (
	"fmt"
	"net/netip"
	"time"

	"bgscan/internal/core/dns"
	"bgscan/internal/core/result"
)

// Schema defines the database layout and parsing rules for VayDNS probe outcomes.
var Schema = result.ResultSchema{
	Name:      "VayDNS",
	Directory: "vaydns",

	Columns: []result.ColumnDef{
		{Name: "IP", Width: 30},
		{Name: "Latency", Width: 15},
		{Name: "Transport", Width: 15},
		{Name: "Port", Width: 10},
		{Name: "Auth", Width: 15},
		{Name: "Proxy", Width: 15},
	},

	Parser: parseVayDNSResult,
}

// VayDNSResult holds the outcome of a single VayDNS tunnel probe.
//
// Latency measures only the proxy validation phase—from the first byte sent
// through the tunnel until a valid response is received. It excludes the initial
// tunnel startup overhead to reflect sustained tunnel performance.
type VayDNSResult struct {
	IP                netip.Addr
	Latency           time.Duration
	Transport         dns.ResolverType      // Underlying DNS transport used for the tunnel (e.g., UDP, DoH, DoT).
	Port              uint16                // Local SOCKS5 port allocated for validation.
	AuthMethod        dns.AuthMethod        // How the tunnel authenticates.
	ResolverProxyType dns.ResolverProxyType // Proxy type used to reach the resolver.
}

func (r VayDNSResult) Key() string {
	return r.IP.String()
}

func (r VayDNSResult) KeyType() result.KeyType {
	return result.KeyIP
}

func (r VayDNSResult) Equal(rs result.Result) bool {
	return r.IP.String() == rs.Key()
}

func (r VayDNSResult) ToRecord() []string {
	return []string{
		r.IP.String(),
		result.FormatDuration(r.Latency),
		string(r.Transport),
		fmt.Sprintf("%d", r.Port),
		string(r.AuthMethod),
		string(r.ResolverProxyType),
	}
}

// Score calculates a performance rating where lower latency yields a higher score.
// A latency of 0 or negative values are clamped to 1ms to prevent division by zero.
func (r VayDNSResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())
	if ms < 1 {
		ms = 1
	}
	return 1000.0 / ms
}

func parseVayDNSResult(record []string) (result.Result, error) {
	if len(record) < 2 {
		return nil, fmt.Errorf(
			"invalid VayDNS result record: expected at least 2 fields, got %d",
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
	var authMethod dns.AuthMethod
	var proxyType dns.ResolverProxyType

	if len(record) >= 4 {
		transport = dns.ParseResolverType(record[2])
		if _, err := fmt.Sscanf(record[3], "%d", &port); err != nil {
			return nil, fmt.Errorf("parse port: %w", err)
		}
	}

	if len(record) >= 6 {
		authMethod = dns.ParseAuthMethod(record[4])
		proxyType = dns.ParseResolverProxyType(record[5])
	} else {
		authMethod = dns.AuthNone
		proxyType = dns.ResolverProxySOCKS
	}

	return VayDNSResult{
		IP:                ip,
		Latency:           latency,
		Transport:         transport,
		Port:              port,
		AuthMethod:        authMethod,
		ResolverProxyType: proxyType,
	}, nil
}
