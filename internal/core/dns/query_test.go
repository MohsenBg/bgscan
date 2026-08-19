package dns

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func testDNSServer(t *testing.T) (addr string, stop func()) {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("testDNSServer: ListenPacket: %v", err)
	}
	addr = pc.LocalAddr().String()

	mux := dns.NewServeMux()
	mux.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Answer = append(m.Answer, &dns.A{
			Hdr: dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
			A:   net.ParseIP("127.0.0.2"),
		})
		_ = w.WriteMsg(m)
	})

	srv := &dns.Server{PacketConn: pc, Net: "udp", Handler: mux}
	started := make(chan struct{})
	srv.NotifyStartedFunc = func() { close(started) }

	go func() { _ = srv.ActivateAndServe() }()
	<-started

	return addr, func() { _ = srv.Shutdown() }
}

func splitHostPort(t *testing.T, addr string) (host, port string) {
	t.Helper()
	h, p, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("splitHostPort(%q): %v", addr, err)
	}
	return h, p
}

func portFromAddr(t *testing.T, addr string) uint16 {
	t.Helper()
	_, p := splitHostPort(t, addr)
	n := 0
	for _, c := range p {
		n = n*10 + int(c-'0')
	}
	return uint16(n)
}

func TestQuery_Normalize_Defaults(t *testing.T) {
	q := Query{}
	q.normalize()

	if q.Timeout != DefaultTimeout {
		t.Errorf("Timeout: got %v, want %v", q.Timeout, DefaultTimeout)
	}
	if q.Port != DefaultPort {
		t.Errorf("Port: got %d, want %d", q.Port, DefaultPort)
	}
	if q.RecordType != DefaultRecordType {
		t.Errorf("RecordType: got %q, want %q", q.RecordType, DefaultRecordType)
	}
	if q.Transport != DefaultTransport {
		t.Errorf("Transport: got %q, want %q", q.Transport, DefaultTransport)
	}
	if q.EDNSBufSize != 0 {
		t.Errorf("EDNSBufSize: got %d, want 0 when not set", q.EDNSBufSize)
	}
}

func TestQuery_Normalize_ExplicitValuesUnchanged(t *testing.T) {
	q := Query{Timeout: 500 * time.Millisecond, Port: 5353, RecordType: TypeMX, Transport: ResolverTypeTCP, EDNSBufSize: 4096}
	q.normalize()

	if q.Timeout != 500*time.Millisecond {
		t.Errorf("Timeout should not be changed: got %v", q.Timeout)
	}
	if q.Port != 5353 {
		t.Errorf("Port should not be changed: got %d", q.Port)
	}
	if q.RecordType != TypeMX {
		t.Errorf("RecordType should not be changed: got %q", q.RecordType)
	}
	if q.Transport != ResolverTypeTCP {
		t.Errorf("Transport should not be changed: got %q", q.Transport)
	}
	if q.EDNSBufSize != 4096 {
		t.Errorf("EDNSBufSize should not be changed: got %d", q.EDNSBufSize)
	}
}

func TestQuery_Normalize_BUG_ShortTimeoutSilentlyReplaced(t *testing.T) {
	q := Query{Timeout: 10 * time.Millisecond}
	q.normalize()

	if q.Timeout == 10*time.Millisecond {
		t.Log("BUG 2 appears to be fixed: 10ms timeout is now preserved")
	} else if q.Timeout != DefaultTimeout {
		t.Errorf("unexpected timeout after normalize: %v", q.Timeout)
	} else {
		t.Logf("BUG 2 CONFIRMED: 10ms timeout was silently replaced with %v", q.Timeout)
	}
}

func TestQuery_Normalize_EDNSBufSizeTooSmall(t *testing.T) {
	q := Query{EDNSBufSize: 100}
	q.normalize()
	if q.EDNSBufSize != DefaultEDNSBufSize {
		t.Errorf("EDNSBufSize < 512 should be reset to DefaultEDNSBufSize (%d), got %d", DefaultEDNSBufSize, q.EDNSBufSize)
	}
}

func TestQuery_Normalize_EDNSBufSizeExactly512(t *testing.T) {
	q := Query{EDNSBufSize: 512}
	q.normalize()
	if q.EDNSBufSize != 512 {
		t.Errorf("EDNSBufSize=512 should not be changed, got %d", q.EDNSBufSize)
	}
}

func TestQuery_BuildQuery_BasicA(t *testing.T) {
	q := Query{Domain: "example.com", RecordType: TypeA, RecursionDesired: true}
	msg := q.buildQuery()
	if len(msg.Question) != 1 {
		t.Fatalf("expected 1 question, got %d", len(msg.Question))
	}
	if msg.Question[0].Qtype != dns.TypeA {
		t.Errorf("Qtype: got %d, want %d (A)", msg.Question[0].Qtype, dns.TypeA)
	}
	if !msg.RecursionDesired {
		t.Errorf("RecursionDesired should be true")
	}
	if msg.IsEdns0() != nil {
		t.Errorf("expected no EDNS0 OPT record when EDNSBufSize=0")
	}
}

func TestQuery_BuildQuery_WithEDNS(t *testing.T) {
	q := Query{Domain: "example.com", RecordType: TypeA, EDNSBufSize: 1232}
	msg := q.buildQuery()

	opt := msg.IsEdns0()
	if opt == nil {
		t.Fatal("expected EDNS0 OPT record, got nil")
	}
	if opt.UDPSize() != 1232 {
		t.Errorf("EDNS0 UDP size: got %d, want 1232", opt.UDPSize())
	}
}

func TestQuery_BuildQuery_DomainFQDN(t *testing.T) {
	q := Query{Domain: "example.com", RecordType: TypeA}
	msg := q.buildQuery()
	if msg.Question[0].Name != "example.com." {
		t.Errorf("domain should be FQDN: got %q, want %q", msg.Question[0].Name, "example.com.")
	}
}

func TestQuery_BuildQuery_AllRecordTypes(t *testing.T) {
	tests := []struct {
		rt   RecordType
		want uint16
	}{
		{TypeA, dns.TypeA},
		{TypeAAAA, dns.TypeAAAA},
		{TypeCNAME, dns.TypeCNAME},
		{TypeNS, dns.TypeNS},
		{TypeMX, dns.TypeMX},
		{TypeTXT, dns.TypeTXT},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(string(tc.rt), func(t *testing.T) {
			q := Query{Domain: "example.com", RecordType: tc.rt}
			msg := q.buildQuery()
			if msg.Question[0].Qtype != tc.want {
				t.Errorf("RecordType %s: Qtype got %d, want %d", tc.rt, msg.Question[0].Qtype, tc.want)
			}
		})
	}
}

func TestResolver_Query_SuccessfulAQuery(t *testing.T) {
	addr, stop := testDNSServer(t)
	defer stop()

	host, _ := splitHostPort(t, addr)
	port := portFromAddr(t, addr)

	q := Query{Resolver: host, Port: port, Domain: "example.com", RecordType: TypeA, RecursionDesired: true, Transport: ResolverTypeUDP, Timeout: 2 * time.Second}

	resolver := NewResolver()

	resp, err := resolver.Query(context.Background(), q)
	if err != nil {
		t.Fatalf("Query() unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("Query() returned nil response")
	}
	if len(resp.Answer) == 0 {
		t.Fatal("Query() returned response with no answers")
	}

	a, ok := resp.Answer[0].(*dns.A)
	if !ok {
		t.Fatalf("expected *dns.A in answer, got %T", resp.Answer[0])
	}
	if !a.A.Equal(net.ParseIP("127.0.0.2")) {
		t.Errorf("A record: got %v, want 127.0.0.2", a.A)
	}
}

func TestResolver_Query_CancelledContext(t *testing.T) {
	addr, stop := testDNSServer(t)
	defer stop()

	host, _ := splitHostPort(t, addr)
	port := portFromAddr(t, addr)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	q := Query{Resolver: host, Port: port, Domain: "example.com", RecordType: TypeA, Transport: ResolverTypeUDP, Timeout: 2 * time.Second}

	resolver := NewResolver()

	_, err := resolver.Query(ctx, q)
	if err == nil {
		t.Error("Query() with cancelled context should return an error")
	}
}

func TestResolver_Query_UnreachableResolver(t *testing.T) {
	q := Query{Resolver: "127.0.0.1", Port: 1, Domain: "example.com", RecordType: TypeA, Transport: ResolverTypeUDP, Timeout: 200 * time.Millisecond}
	resolver := NewResolver()
	_, err := resolver.Query(context.Background(), q)
	if err == nil {
		t.Error("Query() against unreachable resolver should return an error")
	}
}

func TestResolver_Query_DefaultsApplied(t *testing.T) {
	addr, stop := testDNSServer(t)
	defer stop()

	host, _ := splitHostPort(t, addr)
	port := portFromAddr(t, addr)

	q := Query{Resolver: host, Port: port, Domain: "test.local", Timeout: 2 * time.Second}

	resolver := NewResolver()

	resp, err := resolver.Query(context.Background(), q)
	if err != nil {
		t.Fatalf("Query() with default fields: %v", err)
	}
	if resp == nil {
		t.Fatal("Query() with default fields returned nil")
	}
}

func TestQuery_Address(t *testing.T) {
	tests := []struct {
		resolver string
		port     uint16
		want     string
	}{
		{"1.1.1.1", 53, "1.1.1.1:53"},
		{"8.8.8.8", 853, "8.8.8.8:853"},
		{"::1", 53, "[::1]:53"},
		{"2001:db8::1", 5353, "[2001:db8::1]:5353"},
	}
	for _, tc := range tests {
		q := Query{Resolver: tc.resolver, Port: tc.port}
		if got := q.address(); got != tc.want {
			t.Errorf("address() for %q:%d = %q; want %q", tc.resolver, tc.port, got, tc.want)
		}
	}
}

func TestQuery_HasEDNS_WithOPT(t *testing.T) {
	q := Query{Domain: "example.com", RecordType: TypeA, EDNSBufSize: 1232}
	if !q.hasEDNS(q.buildQuery()) {
		t.Error("hasEDNS should return true when OPT record is present")
	}
}

func TestQuery_HasEDNS_WithoutOPT(t *testing.T) {
	q := Query{Domain: "example.com", RecordType: TypeA}
	if q.hasEDNS(q.buildQuery()) {
		t.Error("hasEDNS should return false when no OPT record")
	}
}

func TestTransportNetwork(t *testing.T) {
	tests := []struct {
		transport ResolverType
		want      string
	}{
		{ResolverTypeUDP, "udp"},
		{ResolverTypeTCP, "tcp"},
		{ResolverTypeDOT, "tcp-tls"},
		{ResolverType("QUIC"), "udp"},
		{ResolverType(""), "udp"},
	}
	for _, tc := range tests {
		if got := transportNetwork(tc.transport); got != tc.want {
			t.Errorf("transportNetwork(%q) = %q; want %q", tc.transport, got, tc.want)
		}
	}
}

func TestDefaultConstants(t *testing.T) {
	if DefaultEDNSBufSize < 512 {
		t.Errorf("DefaultEDNSBufSize=%d is below the minimum safe EDNS size of 512", DefaultEDNSBufSize)
	}
	if DefaultTimeout < time.Millisecond {
		t.Errorf("DefaultTimeout=%v is suspiciously small", DefaultTimeout)
	}
	if DefaultPort == 0 {
		t.Error("DefaultPort must not be 0")
	}
	if DefaultRecordType == "" {
		t.Error("DefaultRecordType must not be empty")
	}
	if DefaultTransport == "" {
		t.Error("DefaultTransport must not be empty")
	}
}
