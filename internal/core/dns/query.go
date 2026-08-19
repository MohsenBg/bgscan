package dns

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/miekg/dns"
)

// Msg is an alias for dns.Msg.
type Msg = dns.Msg

// Query contains the parameters for a DNS query.
type Query struct {
	Resolver         string
	Port             uint16
	Domain           string
	Transport        ResolverType
	RecordType       RecordType
	EDNSBufSize      uint16
	RecursionDesired bool
	Timeout          time.Duration
}

const (
	DefaultEDNSBufSize uint16        = 1234
	DefaultTimeout     time.Duration = 2 * time.Second
	DefaultTransport   ResolverType  = ResolverTypeUDP
	DefaultPort        uint16        = 53
	DefaultRecordType  RecordType    = TypeA
)

// Resolver executes DNS queries.
type Resolver interface {
	Query(ctx context.Context, q Query) (*Msg, error)
}

type defaultResolver struct{}

// NewResolver returns a DNS resolver using the configured query transport.
func NewResolver() Resolver {
	return &defaultResolver{}
}

func (r *defaultResolver) Query(ctx context.Context, q Query) (*Msg, error) {
	q.normalize()

	req := q.buildQuery()

	resp, err := q.exchange(ctx, req)
	if err != nil {
		return nil, err
	}

	// Some resolvers reject requests containing EDNS.
	if q.hasEDNS(req) && resp.Rcode != dns.RcodeSuccess {
		if retry, err := q.retryWithoutEDNS(ctx, req); err == nil && retry != nil {
			resp = retry
		}
	}

	// Retry truncated UDP responses over TCP.
	if q.Transport == ResolverTypeUDP && resp != nil && resp.Truncated {
		return q.exchangeTCP(ctx, req)
	}

	return resp, nil
}

func (q Query) buildQuery() *Msg {
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

func (q Query) exchange(ctx context.Context, msg *Msg) (*Msg, error) {
	client := &dns.Client{
		Net:     transportNetwork(q.Transport),
		Timeout: q.Timeout,
	}

	resp, _, err := client.ExchangeContext(ctx, msg, q.address())
	return resp, err
}

func (q Query) exchangeTCP(ctx context.Context, msg *Msg) (*Msg, error) {
	client := &dns.Client{
		Net:     "tcp",
		Timeout: q.Timeout,
	}

	resp, _, err := client.ExchangeContext(ctx, msg, q.address())
	return resp, err
}

func (q Query) retryWithoutEDNS(ctx context.Context, msg *Msg) (*Msg, error) {
	clone := msg.Copy()
	clone.Extra = nil

	return q.exchange(ctx, clone)
}

func (q Query) hasEDNS(msg *Msg) bool {
	for _, rr := range msg.Extra {
		if rr.Header().Rrtype == dns.TypeOPT {
			return true
		}
	}

	return false
}

func (q Query) address() string {
	return net.JoinHostPort(q.Resolver, fmt.Sprint(q.Port))
}

func (q *Query) normalize() {
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

func transportNetwork(t ResolverType) string {
	switch t {
	case ResolverTypeTCP:
		return "tcp"
	case ResolverTypeDOT:
		return "tcp-tls"
	case ResolverTypeUDP:
		return "udp"
	default:
		return "udp"
	}
}
