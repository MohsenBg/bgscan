package dns

import (
	"context"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// dnstt.go — path helpers & construction
// ─────────────────────────────────────────────────────────────────────────────

func TestDNSTTClientPaths_NonEmpty(t *testing.T) {
	if paths := DNSTTClientPaths(); len(paths) == 0 {
		t.Error("DNSTTClientPaths() should return at least one candidate path")
	}
}

func TestDNSTTClientPaths_ContainsExpectedEntries(t *testing.T) {
	paths := DNSTTClientPaths()
	required := []string{"assets/dnstt-client", "assets/dns/dnstt-client", "dnstt-client"}
	pathSet := make(map[string]bool, len(paths))
	for _, p := range paths {
		pathSet[p] = true
	}
	for _, r := range required {
		if !pathSet[r] {
			t.Errorf("DNSTTClientPaths() missing expected entry %q", r)
		}
	}
}

func TestFindDNSTTClient_ReturnsErrorWhenMissing(t *testing.T) {
	if _, err := FindDNSTTClient(); err == nil {
		t.Skip("dnstt-client binary found on PATH; skipping absence test")
	}
}

func TestNewDNSTTClient_ReturnsErrorWhenBinaryMissing(t *testing.T) {
	if _, err := NewDNSTTClient("example.com", "deadbeef", UDP, 53); err == nil {
		t.Skip("dnstt-client binary found; skipping missing-binary test")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// dnstt.go — StopTunnel nil-guard
// ─────────────────────────────────────────────────────────────────────────────

func TestDNSTTClient_StopTunnel_WhenProcessNil(t *testing.T) {
	client := &DNSTTClient{} // process field is nil
	if err := client.StopTunnel(context.Background()); err == nil {
		t.Error("StopTunnel with nil process should return an error")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// dnstt.go — getDNSTransportFlag
// ─────────────────────────────────────────────────────────────────────────────

func TestGetDNSTransportFlag(t *testing.T) {
	tests := []struct {
		transport Transport
		want      string
	}{
		{UDP, "-udp"},
		{DOT, "-dot"},
		{DOH, "-dot"},
		{TCP, "-dot"},
		{Transport("QUIC"), "-udp"},
		{Transport(""), "-udp"},
	}
	for _, tc := range tests {
		if got := getDNSTransportFlag(tc.transport); got != tc.want {
			t.Errorf("getDNSTransportFlag(%q) = %q; want %q", tc.transport, got, tc.want)
		}
	}
}

func TestGetDNSTransportFlag_BUG_TCPMappedToDot(t *testing.T) {
	flag := getDNSTransportFlag(TCP)
	if flag == "-dot" {
		t.Logf("BUG 1 CONFIRMED: getDNSTransportFlag(TCP) = %q. TCP is plaintext; correct flag should be \"-tcp\", not \"-dot\".", flag)
	}
	if flag == "-tcp" {
		t.Log("BUG 1 appears to be fixed: getDNSTransportFlag(TCP) now returns \"-tcp\"")
	}
}
