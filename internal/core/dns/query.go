package dns

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/miekg/dns"
)

// Msg is an alias for dns.Msg from github.com/miekg/dns.
// Provides a cleaner API surface for consumers of the dns package.
type Msg = dns.Msg

// DNSQuery defines the configuration for a single DNS query.
//
// A DNSQuery specifies the resolver address, domain, record type,
// transport protocol, EDNS settings, and timeout rules.
//
// Example:
//
//	q := DNSQuery{
//	    Resolver:   "1.1.1.1",
//	    Domain:     "example.com",
//	    RecordType: TypeA,
//	}
//	resp, err := q.Run()
type DNSQuery struct {
	// Resolver is the target DNS server (IP or hostname).
	Resolver string

	// Port is the resolver port (default: 53).
	Port uint16

	// Domain is the DNS name to query.
	Domain string

	// Transport specifies the network protocol (UDP, TCP, DOT).
	Transport Transport

	// RecordType specifies the DNS record type (A, AAAA, MX, etc.).
	RecordType RecordType

	// EDNSBufSize enables EDNS0 and announces the UDP payload size.
	// Values < 512 are ignored and reset to a safe default.
	EDNSBufSize uint16

	// RecursionDesired toggles the RD bit.
	RecursionDesired bool

	// Timeout is the maximum allowed network duration.
	Timeout time.Duration
}

// Default values used when fields are left unset.
const (
	DefaultEDNSBufSize uint16        = 1234
	DefaultTimeout     time.Duration = 2 * time.Second
	DefaultTransport   Transport     = UDP
	DefaultPort        uint16        = 53
	DefaultRecordType  RecordType    = TypeA
)

// Run executes the DNS query and returns the response.
//
// The algorithm:
//
//  1. Normalize configuration (timeouts, defaults, transport).
//  2. Construct DNS query message.
//  3. Execute query using chosen transport.
//  4. If EDNS is present and RCODE != NOERROR → retry without EDNS.
//  5. If UDP response is truncated → retry over TCP.
//
// Only UDP performs fallback to TCP.
func (q *DNSQuery) Run(ctx context.Context) (*Msg, error) {
	q.normalize()

	req := q.buildQuery()

	resp, err := q.exchange(ctx, req)
	if err != nil {
		return nil, err
	}

	// Retry without EDNS if server rejected the EDNS version.
	if q.hasEDNS(req) && resp.Rcode != dns.RcodeSuccess {
		if alt, err := q.retryWithoutEDNS(ctx, req); err == nil && alt != nil {
			resp = alt
		}
	}

	// Retry over TCP if UDP reply was truncated.
	if q.Transport == UDP && resp != nil && resp.Truncated {
		return q.exchangeTCP(ctx, req)
	}

	return resp, nil
}

// buildQuery constructs a dns.Msg based on the query configuration.
func (q *DNSQuery) buildQuery() *Msg {
	m := new(Msg)

	m.SetQuestion(
		dns.Fqdn(q.Domain),
		toMiekgDNS(q.RecordType),
	)

	m.RecursionDesired = q.RecursionDesired

	if q.EDNSBufSize > 0 {
		m.SetEdns0(q.EDNSBufSize, false)
	}

	return m
}

// exchange performs a DNS query using the selected transport.
func (q *DNSQuery) exchange(ctx context.Context, msg *Msg) (*Msg, error) {
	client := &dns.Client{
		Net:     transportNetwork(q.Transport),
		Timeout: q.Timeout,
	}

	resp, _, err := client.ExchangeContext(ctx, msg, q.address())
	return resp, err
}

// exchangeTCP reruns the query specifically over TCP.
//
// Used when a UDP response is truncated (TC bit set).
func (q *DNSQuery) exchangeTCP(ctx context.Context, msg *Msg) (*Msg, error) {
	client := &dns.Client{
		Net:     "tcp",
		Timeout: q.Timeout,
	}

	resp, _, err := client.ExchangeContext(ctx, msg, q.address())
	return resp, err
}

// retryWithoutEDNS removes the EDNS OPT record and retries the query.
// Some resolvers return FORMERR when EDNS is present.
func (q *DNSQuery) retryWithoutEDNS(ctx context.Context, msg *Msg) (*Msg, error) {
	clone := msg.Copy()
	clone.Extra = nil
	return q.exchange(ctx, clone)
}

// hasEDNS reports whether the request message includes an EDNS OPT record.
func (q *DNSQuery) hasEDNS(msg *Msg) bool {
	for _, rr := range msg.Extra {
		if rr.Header().Rrtype == dns.TypeOPT {
			return true
		}
	}
	return false
}

// address returns "ip:port" for the resolver.
func (q *DNSQuery) address() string {
	return net.JoinHostPort(q.Resolver, fmt.Sprint(q.Port))
}

// normalize applies defaults to missing configuration values.
func (q *DNSQuery) normalize() {
	if q.Timeout < 50*time.Millisecond {
		q.Timeout = DefaultTimeout
	}
	if q.EDNSBufSize > 0 && q.EDNSBufSize < 512 {
		q.EDNSBufSize = DefaultEDNSBufSize
	}
	if q.Port == 0 {
		q.Port = DefaultPort
	}
	if q.RecordType == "" {
		q.RecordType = DefaultRecordType
	}
	if q.Transport == "" {
		q.Transport = DefaultTransport
	}
}

// transportNetwork converts a Transport into the network string
// expected by github.com/miekg/dns.
//
// DOT uses "tcp-tls" as defined by the library.
func transportNetwork(t Transport) string {
	switch t {
	case TCP:
		return "tcp"
	case DOT:
		return "tcp-tls"
	case UDP:
		return "udp"
	default:
		return "udp"
	}
}
