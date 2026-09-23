package engine

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
)

func TestStripTargetIP_RemovesBothForms(t *testing.T) {
	t.Parallel()

	ip := netip.MustParseAddr("1.2.3.4")
	got := stripTargetIP(ip, "dial tcp 1.2.3.4:80: connect: refused")
	want := "dial tcp <ip>:80: connect: refused"
	if got != want {
		t.Fatalf("stripTargetIP = %q, want %q", got, want)
	}

	v6 := netip.MustParseAddr("::ffff:1.2.3.4")
	got = stripTargetIP(v6, "dial tcp 1.2.3.4:443: i/o timeout")
	want = "dial tcp <ip>:443: i/o timeout"
	if got != want {
		t.Fatalf("stripTargetIP (v4-in-v6) = %q, want %q", got, want)
	}
}

func TestFirstSightingMessage_HidesIPStatesCauseAndHintsDebug(t *testing.T) {
	t.Parallel()

	ip := netip.MustParseAddr("1.2.3.4")
	err := fmt.Errorf("dial tcp %s:80: connect: network is unreachable", ip)

	msg := firstSightingMessage(ip, err)

	if strings.Contains(msg, "1.2.3.4") {
		t.Fatalf("default-level message must not contain target IP: %q", msg)
	}
	if !strings.Contains(msg, "network is unreachable") {
		t.Fatalf("message must state the cause: %q", msg)
	}
	if !strings.Contains(msg, "set logger level to debug") {
		t.Fatalf("message must hint at the debug level: %q", msg)
	}
}

func TestFailureSignature_StripTargetIP(t *testing.T) {
	t.Parallel()

	ip := netip.MustParseAddr("1.2.3.4")
	other := netip.MustParseAddr("5.6.7.8")

	got := failureSignature(ip, fmt.Errorf("dial tcp %s:80: connect: refused", ip))
	want := "dial tcp <ip>:80: connect: refused"
	if got != want {
		t.Fatalf("failureSignature = %q, want %q", got, want)
	}

	// The same failure mode against a different IP must collapse to the
	// same signature, otherwise every error looks distinct.
	gotOther := failureSignature(other, fmt.Errorf("dial tcp %s:80: connect: refused", other))
	if gotOther != got {
		t.Fatalf("signature not stable across targets: %q vs %q", got, gotOther)
	}
}

func TestFailureSignature_UnmappedV4InV6(t *testing.T) {
	t.Parallel()

	ip := netip.MustParseAddr("::ffff:1.2.3.4")
	got := failureSignature(ip, fmt.Errorf("dial tcp %s:443: i/o timeout", "1.2.3.4"))
	want := "dial tcp <ip>:443: i/o timeout"
	if got != want {
		t.Fatalf("failureSignature = %q, want %q", got, want)
	}
}

func TestErrTripwire_FirstSightingOnly(t *testing.T) {
	t.Parallel()

	tr := newErrTripwire()
	ip := netip.MustParseAddr("1.2.3.4")
	other := netip.MustParseAddr("5.6.7.8")

	if !tr.note(ip, fmt.Errorf("dial tcp %s:80: boom", ip)) {
		t.Fatal("first sighting should be new")
	}
	if tr.note(other, fmt.Errorf("dial tcp %s:80: boom", other)) {
		t.Fatal("same failure on another IP should be a repeat")
	}
	if !tr.note(ip, errors.New("a different failure")) {
		t.Fatal("a distinct failure should be new")
	}
}

func TestProbeFailures_EnvironmentThresholdAbortsOnce(t *testing.T) {
	t.Parallel()

	var aborts, onErrors atomic.Int32
	fails := newProbeFailures(nil, func() { aborts.Add(1) }, ScanHooks{
		OnError: func(error) { onErrors.Add(1) },
	})

	ip := netip.MustParseAddr("1.2.3.4")
	envErr := fmt.Errorf("%w: local port pool exhausted", probe.ErrEnvironment)

	for range envFailThreshold {
		fails.note(ip, envErr)
	}

	if aborts.Load() != 1 {
		t.Fatalf("abort calls = %d, want 1", aborts.Load())
	}
	if onErrors.Load() != 1 {
		t.Fatalf("OnError calls = %d, want 1", onErrors.Load())
	}

	// Further environment failures must not re-abort or re-fire OnError.
	for range 5 {
		fails.note(ip, envErr)
	}
	if aborts.Load() != 1 || onErrors.Load() != 1 {
		t.Fatalf("abort/OnError = %d/%d, want 1/1", aborts.Load(), onErrors.Load())
	}
}

func TestProbeFailures_ResetOnSuccessAndNonEnv(t *testing.T) {
	t.Parallel()

	var aborts atomic.Int32
	fails := newProbeFailures(nil, func() { aborts.Add(1) }, ScanHooks{})

	ip := netip.MustParseAddr("1.2.3.4")
	envErr := fmt.Errorf("%w: no route", probe.ErrEnvironment)

	// A success resets the streak: threshold-1 env fails, reset, then
	// threshold-1 more must not abort.
	for range envFailThreshold - 1 {
		fails.note(ip, envErr)
	}
	fails.resetEnv()
	for range envFailThreshold - 1 {
		fails.note(ip, envErr)
	}
	if aborts.Load() != 0 {
		t.Fatalf("abort fired after reset, calls = %d", aborts.Load())
	}

	// A non-environment failure also resets the streak.
	fails.note(ip, errors.New("target refused"))
	for range envFailThreshold - 1 {
		fails.note(ip, envErr)
	}
	if aborts.Load() != 0 {
		t.Fatalf("abort fired after non-env reset, calls = %d", aborts.Load())
	}

	// Crossing the threshold after the resets aborts exactly once.
	fails.note(ip, envErr)
	if aborts.Load() != 1 {
		t.Fatalf("abort calls = %d, want 1", aborts.Load())
	}
}
