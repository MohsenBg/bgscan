package resolveprobe

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"bgscan/internal/core/dns"
)

var testIP = netip.MustParseAddr("1.2.3.4")

// baseRequest returns a standard DNSRequest configuration for testing.
func baseRequest() *DNSRequest {
	return &DNSRequest{
		Domain:         "example.com",
		Port:           53,
		CheckTypes:     []string{"A"},
		AcceptedRcodes: []uint16{0},
		Timeout:        time.Second,
		Tries:          1,
		DpiTries:       1,
	}
}

// msg creates a dns.Msg with the specified rcode.
func msg(rcode int) *dns.Msg {
	m := new(dns.Msg)
	m.Rcode = rcode
	return m
}

// stubQuery returns a queryRunner that consistently yields the given rcode and error.
func stubQuery(rcode int, err error) queryRunner {
	return func(_ context.Context, _ dns.DNSQuery) (*dns.Msg, error) {
		if err != nil {
			return nil, err
		}
		return msg(rcode), nil
	}
}

// probeWith constructs a ResolverProbe with the provided request and queryRunner,
// allowing injection of mock behaviors for testing.
func probeWith(req *DNSRequest, qr queryRunner) *ResolverProbe {
	p := NewResolverProbe(req).(*ResolverProbe)
	p.runQuery = qr
	return p
}

// TestNewResolverProbe_TriesNormalized ensures that zero or negative retry
// counts are clamped to 1.
func TestNewResolverProbe_TriesNormalized(t *testing.T) {
	for _, tries := range []int{0, -1, -100} {
		req := baseRequest()
		req.Tries = tries
		req.DpiTries = tries
		p := NewResolverProbe(req).(*ResolverProbe)
		if p.request.Tries != 1 {
			t.Errorf("Tries %d: want 1, got %d", tries, p.request.Tries)
		}
		if p.request.DpiTries != 1 {
			t.Errorf("DpiTries %d: want 1, got %d", tries, p.request.DpiTries)
		}
	}
}

func TestInit(t *testing.T) {
	p := probeWith(baseRequest(), stubQuery(0, nil))
	if err := p.Init(context.Background()); err != nil {
		t.Errorf("Init: %v", err)
	}
}

func TestClose(t *testing.T) {
	p := probeWith(baseRequest(), stubQuery(0, nil))
	if err := p.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestRun_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := probeWith(baseRequest(), stubQuery(0, nil))
	_, err := p.Run(ctx, testIP)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("want context.Canceled, got %v", err)
	}
}

func TestRun_Success(t *testing.T) {
	p := probeWith(baseRequest(), stubQuery(0, nil))

	res, err := p.Run(context.Background(), testIP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := res.(ResolverResult)
	if r.IP != testIP {
		t.Errorf("IP = %v, want %v", r.IP, testIP)
	}
	if r.RecordType != "A" {
		t.Errorf("RecordType = %q, want A", r.RecordType)
	}
	if r.Rcode != 0 {
		t.Errorf("Rcode = %d, want 0", r.Rcode)
	}
	if r.Tries != 1 {
		t.Errorf("Tries = %d, want 1", r.Tries)
	}
	if r.DPIChecked {
		t.Error("DPIChecked should be false when DpiCheck is disabled")
	}
}

func TestRun_NoAcceptedResponse(t *testing.T) {
	req := baseRequest()
	req.AcceptedRcodes = []uint16{0}

	// rcode 3 (NXDOMAIN) is not in AcceptedRcodes.
	p := probeWith(req, stubQuery(3, nil))
	_, err := p.Run(context.Background(), testIP)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestRun_QueryError_RetriesExhausted ensures that network errors trigger
// retries up to the configured Tries limit.
func TestRun_QueryError_RetriesExhausted(t *testing.T) {
	req := baseRequest()
	req.Tries = 3

	calls := 0
	p := probeWith(req, func(_ context.Context, _ dns.DNSQuery) (*dns.Msg, error) {
		calls++
		return nil, errors.New("network error")
	})

	_, err := p.Run(context.Background(), testIP)
	if err == nil {
		t.Fatal("expected error after exhausted retries")
	}
	if calls != 3 {
		t.Errorf("want 3 query attempts, got %d", calls)
	}
}

// TestRun_SucceedsOnRetry ensures that transient network errors are retried
// and the probe eventually succeeds.
func TestRun_SucceedsOnRetry(t *testing.T) {
	req := baseRequest()
	req.Tries = 3

	calls := 0
	p := probeWith(req, func(_ context.Context, _ dns.DNSQuery) (*dns.Msg, error) {
		calls++
		if calls < 3 {
			return nil, errors.New("transient error")
		}
		return msg(0), nil
	})

	res, err := p.Run(context.Background(), testIP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(ResolverResult).Tries != 3 {
		t.Errorf("Tries = %d, want 3", res.(ResolverResult).Tries)
	}
}

// TestRun_UnacceptedRcodeStopsRetries ensures that a rejected rcode stops
// retries immediately, as retries are reserved for network errors.
func TestRun_UnacceptedRcodeStopsRetries(t *testing.T) {
	req := baseRequest()
	req.Tries = 3
	req.AcceptedRcodes = []uint16{0}

	calls := 0
	p := probeWith(req, func(_ context.Context, _ dns.DNSQuery) (*dns.Msg, error) {
		calls++
		return msg(3), nil // NXDOMAIN
	})

	p.Run(context.Background(), testIP) //nolint:errcheck
	if calls != 1 {
		t.Errorf("want 1 call (no retry on unaccepted rcode), got %d", calls)
	}
}

// TestRun_FirstAcceptedCheckTypeWins ensures the probe stops querying as
// soon as a CheckType yields an accepted rcode.
func TestRun_FirstAcceptedCheckTypeWins(t *testing.T) {
	req := baseRequest()
	req.CheckTypes = []string{"AAAA", "A"}
	req.AcceptedRcodes = []uint16{0}

	var queriedTypes []dns.RecordType
	p := probeWith(req, func(_ context.Context, q dns.DNSQuery) (*dns.Msg, error) {
		queriedTypes = append(queriedTypes, q.RecordType)
		return msg(0), nil
	})

	res, err := p.Run(context.Background(), testIP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(ResolverResult).RecordType != "AAAA" {
		t.Errorf("RecordType = %q, want AAAA", res.(ResolverResult).RecordType)
	}
	if len(queriedTypes) != 1 {
		t.Errorf("want 1 query, got %d", len(queriedTypes))
	}
}

// TestRun_FallsBackToNextCheckType ensures the probe moves to the next
// CheckType if the current one is rejected.
func TestRun_FallsBackToNextCheckType(t *testing.T) {
	req := baseRequest()
	req.CheckTypes = []string{"AAAA", "A"}
	req.AcceptedRcodes = []uint16{0}

	p := probeWith(req, func(_ context.Context, q dns.DNSQuery) (*dns.Msg, error) {
		if q.RecordType == dns.TypeAAAA {
			return msg(3), nil // rejected
		}
		return msg(0), nil
	})

	res, err := p.Run(context.Background(), testIP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(ResolverResult).RecordType != "A" {
		t.Errorf("RecordType = %q, want A", res.(ResolverResult).RecordType)
	}
}

func TestRun_MultipleAcceptedRcodes(t *testing.T) {
	req := baseRequest()
	req.AcceptedRcodes = []uint16{0, 3} // NOERROR and NXDOMAIN both ok

	p := probeWith(req, stubQuery(3, nil))
	_, err := p.Run(context.Background(), testIP)
	if err != nil {
		t.Errorf("rcode 3 should be accepted, got error: %v", err)
	}
}

// TestRun_ContextCanceledDuringRetry ensures that context cancellation
// halts the retry loop immediately.
func TestRun_ContextCanceledDuringRetry(t *testing.T) {
	req := baseRequest()
	req.Tries = 10

	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	p := probeWith(req, func(_ context.Context, _ dns.DNSQuery) (*dns.Msg, error) {
		calls++
		if calls == 2 {
			cancel()
		}
		return nil, errors.New("error")
	})

	_, err := p.Run(ctx, testIP)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("want context.Canceled, got %v", err)
	}
}

// TestRun_DpiCheck_HonestResolver ensures that a resolver returning a
// non-zero rcode for a .invalid domain passes the honesty check.
func TestRun_DpiCheck_HonestResolver(t *testing.T) {
	req := baseRequest()
	req.DpiCheck = true
	req.DpiTries = 1

	// First call = DPI check (rcode 3 = honest), second = normal probe (rcode 0 = success).
	calls := 0
	p := probeWith(req, func(_ context.Context, _ dns.DNSQuery) (*dns.Msg, error) {
		calls++
		if calls == 1 {
			return msg(3), nil
		}
		return msg(0), nil
	})

	res, err := p.Run(context.Background(), testIP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.(ResolverResult).DPIChecked {
		t.Error("DPIChecked should be true")
	}
}

// TestRun_DpiCheck_DetectsHijacking ensures that a resolver returning rcode 0
// for a .invalid domain is flagged as hijacked.
func TestRun_DpiCheck_DetectsHijacking(t *testing.T) {
	req := baseRequest()
	req.DpiCheck = true
	req.DpiTries = 1

	// rcode 0 for .invalid domain = DPI/hijacking detected
	p := probeWith(req, stubQuery(0, nil))
	_, err := p.Run(context.Background(), testIP)
	if err == nil {
		t.Fatal("expected DPI detection error")
	}
}

// TestRun_DpiCheck_RetriesOnError ensures that the DPI check retries on
// network errors up to the DpiTries limit.
func TestRun_DpiCheck_RetriesOnError(t *testing.T) {
	req := baseRequest()
	req.DpiCheck = true
	req.DpiTries = 3

	calls := 0
	p := probeWith(req, func(_ context.Context, _ dns.DNSQuery) (*dns.Msg, error) {
		calls++
		return nil, errors.New("timeout")
	})

	_, err := p.Run(context.Background(), testIP)
	if err == nil {
		t.Fatal("expected error after DPI retries exhausted")
	}
	// Only DPI queries; normal probe is never reached.
	if calls != 3 {
		t.Errorf("want 3 DPI attempts, got %d", calls)
	}
}

// TestRun_DpiCheck_SucceedsOnRetry ensures that the DPI check succeeds if
// it eventually returns an honest response after transient errors.
func TestRun_DpiCheck_SucceedsOnRetry(t *testing.T) {
	req := baseRequest()
	req.DpiCheck = true
	req.DpiTries = 3

	dpiCalls := 0
	p := probeWith(req, func(_ context.Context, q dns.DNSQuery) (*dns.Msg, error) {
		// Distinguish DPI queries by .invalid suffix.
		if len(q.Domain) > 8 && q.Domain[len(q.Domain)-8:] == ".invalid" {
			dpiCalls++
			if dpiCalls < 3 {
				return nil, errors.New("timeout")
			}
			return msg(3), nil // honest on 3rd try
		}
		return msg(0), nil
	})

	_, err := p.Run(context.Background(), testIP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRun_DpiCheck_DefaultTimeout ensures DpiTimeout defaults to 500ms
// when explicitly set to 0.
func TestRun_DpiCheck_DefaultTimeout(t *testing.T) {
	req := baseRequest()
	req.DpiCheck = true
	req.DpiTimeout = 0 // should default to 500ms

	var capturedTimeout time.Duration
	calls := 0
	p := probeWith(req, func(_ context.Context, q dns.DNSQuery) (*dns.Msg, error) {
		calls++
		if calls == 1 {
			capturedTimeout = q.Timeout
			return msg(3), nil
		}
		return msg(0), nil
	})

	p.Run(context.Background(), testIP) //nolint:errcheck
	if capturedTimeout != 500*time.Millisecond {
		t.Errorf("DPI timeout = %v, want 500ms", capturedTimeout)
	}
}

// TestRun_DpiCheck_ContextCanceledDuringDpi ensures that context cancellation
// halts the DPI check retry loop.
func TestRun_DpiCheck_ContextCanceledDuringDpi(t *testing.T) {
	req := baseRequest()
	req.DpiCheck = true
	req.DpiTries = 10

	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	p := probeWith(req, func(_ context.Context, _ dns.DNSQuery) (*dns.Msg, error) {
		calls++
		if calls == 2 {
			cancel()
		}
		return nil, errors.New("error")
	})

	_, err := p.Run(ctx, testIP)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("want context.Canceled, got %v", err)
	}
}

// TestParseRecordType verifies that parseRecordType correctly maps strings to
// dns.RecordType, handling case insensitivity, whitespace, and defaults.
func TestParseRecordType(t *testing.T) {
	tests := []struct {
		input string
		want  dns.RecordType
	}{
		{"A", dns.TypeA},
		{"a", dns.TypeA},
		{"AAAA", dns.TypeAAAA},
		{"aaaa", dns.TypeAAAA},
		{"TXT", dns.TypeTXT},
		{"NS", dns.TypeNS},
		{"CNAME", dns.TypeCNAME},
		{"MX", dns.TypeMX},
		{"unknown", dns.TypeA}, // default
		{"", dns.TypeA},        // default
		{"  A  ", dns.TypeA},   // whitespace trimmed
	}
	for _, tt := range tests {
		if got := parseRecordType(tt.input); got != tt.want {
			t.Errorf("parseRecordType(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// TestRun_QueryReceivesCorrectResolver ensures the underlying DNS query is
// constructed with the correct resolver IP and port.
func TestRun_QueryReceivesCorrectResolver(t *testing.T) {
	req := baseRequest()
	req.Port = 5353

	var capturedQuery dns.DNSQuery
	p := probeWith(req, func(_ context.Context, q dns.DNSQuery) (*dns.Msg, error) {
		capturedQuery = q
		return msg(0), nil
	})

	p.Run(context.Background(), testIP) //nolint:errcheck
	if capturedQuery.Resolver != testIP.String() {
		t.Errorf("Resolver = %q, want %q", capturedQuery.Resolver, testIP.String())
	}
	if capturedQuery.Port != 5353 {
		t.Errorf("Port = %d, want 5353", capturedQuery.Port)
	}
}
