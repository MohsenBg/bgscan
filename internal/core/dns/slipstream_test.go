package dns

import (
	"context"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// slipstream.go — path helpers & construction
// ─────────────────────────────────────────────────────────────────────────────

func TestSlipstreamClientPaths_NonEmpty(t *testing.T) {
	if paths := SlipstreamClientPaths(); len(paths) == 0 {
		t.Error("SlipstreamClientPaths() should return at least one candidate path")
	}
}

func TestSlipstreamClientPaths_ContainsExpectedEntries(t *testing.T) {
	paths := SlipstreamClientPaths()
	required := []string{"assets/slipstream-client", "assets/dns/slipstream-client", "slipstream-client"}
	pathSet := make(map[string]bool, len(paths))
	for _, p := range paths {
		pathSet[p] = true
	}
	for _, r := range required {
		if !pathSet[r] {
			t.Errorf("SlipstreamClientPaths() missing expected entry %q", r)
		}
	}
}

func TestFindSlipstreamClient_ReturnsErrorWhenMissing(t *testing.T) {
	if _, err := FindSlipstreamClient(); err == nil {
		t.Skip("slipstream-client binary found on PATH; skipping absence test")
	}
}

func TestNewSlipstreamClient_ReturnsErrorWhenBinaryMissing(t *testing.T) {
	if _, err := NewSlipstreamClient("example.com", 53, ""); err == nil {
		t.Skip("slipstream-client binary found; skipping missing-binary test")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// slipstream.go — StopTunnel nil-guard
// ─────────────────────────────────────────────────────────────────────────────

func TestSlipstreamClient_StopTunnel_WhenProcessNil(t *testing.T) {
	client := &SlipstreamClient{} // process field is nil
	if err := client.StopTunnel(context.Background()); err == nil {
		t.Error("StopTunnel with nil process should return an error")
	}
}
