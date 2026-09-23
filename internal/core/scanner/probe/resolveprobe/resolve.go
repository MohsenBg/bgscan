package resolveprobe

import (
	"context"
	"fmt"
	"math/rand"
	"net/netip"
	"slices"
	"strings"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
)

// queryExecutor abstracts DNS query execution, primarily to allow
// mocking in tests.
type queryExecutor func(ctx context.Context, q dns.Query) (*dns.Msg, error)

// DNSRequest configures the parameters for testing a DNS resolver.
type DNSRequest struct {
	// Domain is the base domain to query.
	// If RandomSubdomain is true, a random prefix is added to bypass cache.
	Domain string

	// Port is the resolver port (e.g., 53).
	Port uint16

	// RandomSubdomain prevents resolver caching by appending a random
	// label to the Domain for each probe.
	RandomSubdomain bool

	// DpiCheck enables an honesty check using a guaranteed NXDOMAIN
	// (.invalid) prior to normal queries to detect hijacking or DPI.
	DpiCheck bool

	// DpiTimeout is the per-request timeout for DPI verification queries.
	// Defaults to 500ms if zero.
	DpiTimeout time.Duration

	// DpiTries is the max number of DPI verification attempts.
	// Defaults to 1 if zero or negative.
	DpiTries int

	// Edns0Size sets the advertised EDNS0 UDP buffer size.
	Edns0Size uint16

	// CheckTypes is the ordered list of DNS record types to test
	// (e.g., "A", "AAAA"). The first acceptable response stops the probe.
	CheckTypes []string

	// AcceptedRcodes lists the DNS response codes considered successful.
	// Responses outside this list are treated as failures for that type.
	AcceptedRcodes []uint16

	// Timeout is the per-query timeout for normal resolver tests.
	Timeout time.Duration

	// Transport is the underlying mechanism (UDP, TCP, DoT, DoH, etc.).
	Transport dns.ResolverType

	// Tries is the max number of retries per record type during normal probing.
	// Defaults to 1 if zero or negative.
	Tries int
}

// ResolverProbe validates a single DNS resolver IP address.
// It optionally performs a DPI/hijacking honesty check before executing
// standard record queries based on the DNSRequest configuration.
type ResolverProbe struct {
	request  *DNSRequest
	query    queryExecutor
	resolver dns.Resolver
}

// NewResolverProbe returns a probe that mutates req to ensure Tries and
// DpiTries are at least 1.
func NewResolverProbe(req *DNSRequest) probe.Probe {
	if req.Tries <= 0 {
		req.Tries = 1
	}

	if req.DpiTries <= 0 {
		req.DpiTries = 1
	}

	resolver := dns.NewResolver()

	return &ResolverProbe{
		request:  req,
		resolver: resolver,
		query: func(ctx context.Context, q dns.Query) (*dns.Msg, error) {
			return resolver.Query(ctx, q)
		},
	}
}

// Schema returns the result schema for this probe.
func (r *ResolverProbe) Schema() result.ResultSchema {
	return Schema
}

// Init is a no-op: the probe is stateless.
func (r *ResolverProbe) Init(_ context.Context) error {
	return nil
}

func (r *ResolverProbe) Close() error {
	return nil
}

// Run validates the resolver at ip, with an optional DPI honesty check
// first.
func (r *ResolverProbe) Run(ctx context.Context, ip netip.Addr) (result.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if r.request.DpiCheck {
		if err := r.verifyResolverHonesty(ctx, ip); err != nil {
			return nil, err
		}
	}

	return r.executeNormalProbe(ctx, ip)
}

// verifyResolverHonesty queries a guaranteed-invalid .invalid domain; rcode 0
// means the resolver answers for names that cannot exist (hijacking/DPI).
func (r *ResolverProbe) verifyResolverHonesty(ctx context.Context, ip netip.Addr) error {
	fakeDomain := generateRandomString(16) + ".invalid"

	timeout := r.request.DpiTimeout
	if timeout == 0 {
		timeout = 500 * time.Millisecond
	}

	query := dns.Query{
		Resolver:         ip.String(),
		Port:             r.request.Port,
		Domain:           fakeDomain,
		RecordType:       dns.TypeA,
		Transport:        r.request.Transport,
		EDNSBufSize:      r.request.Edns0Size,
		RecursionDesired: true,
		Timeout:          timeout,
	}

	var lastErr error

	for i := 0; i < r.request.DpiTries; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		resp, err := r.query(ctx, query)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.Rcode == 0 {
			return fmt.Errorf("dpi detected: resolver returned rcode 0 for %s", fakeDomain)
		}

		return nil
	}

	return fmt.Errorf("dpi verification failed after %d tries: %w", r.request.DpiTries, lastErr)
}

// executeNormalProbe returns the first CheckType yielding an AcceptedRcode.
func (r *ResolverProbe) executeNormalProbe(ctx context.Context, ip netip.Addr) (result.Result, error) {
	query := dns.Query{
		Resolver:         ip.String(),
		Port:             r.request.Port,
		Transport:        r.request.Transport,
		EDNSBufSize:      r.request.Edns0Size,
		RecursionDesired: true,
		Timeout:          r.request.Timeout,
	}

	target := r.request.Domain
	if r.request.RandomSubdomain {
		target = generateRandomString(10) + "." + target
	}
	query.Domain = target

	for _, typeStr := range r.request.CheckTypes {
		query.RecordType = parseRecordType(typeStr)

		var lastErr error

		for i := 0; i < r.request.Tries; i++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			start := time.Now()

			resp, err := r.query(ctx, query)
			if err != nil {
				lastErr = err
				continue
			}

			latency := time.Since(start)

			if r.isRcodeAccepted(uint16(resp.Rcode)) {
				return ResolverResult{
					IP:         ip,
					Latency:    latency,
					RecordType: strings.ToUpper(typeStr),
					Tries:      i + 1,
					Rcode:      uint16(resp.Rcode),
					DPIChecked: r.request.DpiCheck,
				}, nil
			}

			break
		}

		_ = lastErr
	}

	return nil, fmt.Errorf("no accepted response for %s", target)
}

func (r *ResolverProbe) isRcodeAccepted(code uint16) bool {
	return slices.Contains(r.request.AcceptedRcodes, code)
}

// parseRecordType maps a string to a dns.RecordType, defaulting to TypeA.
func parseRecordType(s string) dns.RecordType {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "A":
		return dns.TypeA
	case "AAAA":
		return dns.TypeAAAA
	case "TXT":
		return dns.TypeTXT
	case "NS":
		return dns.TypeNS
	case "CNAME":
		return dns.TypeCNAME
	case "MX":
		return dns.TypeMX
	default:
		return dns.TypeA
	}
}

// generateRandomString returns n random lowercase alphanumerics. math/rand is
// fine here: cache-busting needs no cryptographic strength.
func generateRandomString(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}

	return string(b)
}
