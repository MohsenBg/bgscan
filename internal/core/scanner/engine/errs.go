package engine

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/MohsenBg/bgscan/internal/core/scanner/probe"
	"github.com/MohsenBg/bgscan/internal/logger"
)

// envFailThreshold is the number of consecutive probe.ErrEnvironment
// failures tolerated before the scan is aborted. Local-environment
// failures (no IPv6 route, exhausted port pool, broken local Xray) cannot
// be fixed by probing more IPs, so the scan stops instead of grinding
// through the rest of the list.
const envFailThreshold = 20

// firstSightingHint tells the user why only one Error line appeared and
// how to see the rest: the IP-bearing detail for every occurrence lives
// at Debug, and repeats are not echoed at Error to avoid log spam.
const firstSightingHint = "identical errors on other IPs go to debug level — set logger level to debug to see them; repeats are ignored here to avoid log spam"

// errTripwire tracks distinct per-IP probe failure signatures so each one
// is announced once at Error (cause only, no target IP) while every
// occurrence — first included — is available at Debug with full detail.
// One systematic cause leaves a single visible breadcrumb at the default
// level instead of flooding the log.
type errTripwire struct {
	seen map[string]struct{}
	mu   sync.Mutex
}

func newErrTripwire() *errTripwire {
	return &errTripwire{seen: make(map[string]struct{})}
}

// note records the failure signature and reports whether it is the first
// sighting.
func (t *errTripwire) note(ip netip.Addr, err error) bool {
	sig := failureSignature(ip, err)

	t.mu.Lock()
	defer t.mu.Unlock()

	if _, dup := t.seen[sig]; dup {
		return false
	}
	t.seen[sig] = struct{}{}
	return true
}

// stripTargetIP replaces every occurrence of the target IP (both its
// netip.String and unmapped form) with <ip>, so default-level lines can
// describe the cause without revealing which target failed, and so
// signatures collapse across IPs — dial errors embed "1.2.3.4:80" /
// "[v6]:80", and without stripping every error would look unique.
func stripTargetIP(ip netip.Addr, msg string) string {
	if !ip.IsValid() {
		return msg
	}

	msg = strings.ReplaceAll(msg, ip.String(), "<ip>")
	if un := ip.Unmap().String(); un != ip.String() {
		msg = strings.ReplaceAll(msg, un, "<ip>")
	}
	return msg
}

// failureSignature is the dedupe key for a per-IP probe error: the error
// text with the target IP stripped.
func failureSignature(ip netip.Addr, err error) string {
	return stripTargetIP(ip, err.Error())
}

// firstSightingMessage renders the default-level (Error) line: the cause
// with the target IP removed, plus the hint about where the per-IP detail
// went. Pre-rendered on purpose — callers pass it to the logger with no
// format args so '%' inside error text is never reinterpreted.
func firstSightingMessage(ip netip.Addr, err error) string {
	return fmt.Sprintf("probe failed: %s (%s)", stripTargetIP(ip, err.Error()), firstSightingHint)
}

// probeFailures tracks probe errors for one scan: a tripwire that levels
// the per-IP log lines, and a consecutive ErrEnvironment counter that
// aborts the scan when the local environment cannot make progress.
type probeFailures struct {
	log      *logger.Logger
	tripwire *errTripwire
	hooks    ScanHooks

	envFails  atomic.Uint64
	abortOnce sync.Once
	abort     func()
}

// newProbeFailures creates the shared failure tracker for a scan. abort
// cancels the scan ctx; hooks receives at most one OnError when the
// environment threshold is hit. Both may be nil.
func newProbeFailures(log *logger.Logger, abort func(), hooks ScanHooks) *probeFailures {
	return &probeFailures{
		log:      log,
		tripwire: newErrTripwire(),
		hooks:    hooks,
		abort:    abort,
	}
}

// note records a failed probe against ip. The first sighting of a
// distinct failure logs one Error line stating the cause (target IP
// stripped) plus the debug-level hint; every occurrence, first included,
// also logs at Debug with the full "probe failed for <ip>: <err>" text,
// so enabling the debug level reveals IPs and all repeats. Non-environment
// failures reset the consecutive environment streak; once envFailThreshold
// environment failures occur in a row, the scan is cancelled and OnError
// fires once.
func (f *probeFailures) note(ip netip.Addr, err error) {
	if f.tripwire.note(ip, err) {
		f.log.Error("%s", firstSightingMessage(ip, err))
	}
	f.log.Debug("probe failed for %s: %v", ip, err)

	if !errors.Is(err, probe.ErrEnvironment) {
		f.envFails.Store(0)
		return
	}

	n := f.envFails.Add(1)
	if n < envFailThreshold {
		return
	}

	f.abortOnce.Do(func() {
		f.log.Error("aborting scan after %d consecutive probe environment failures: %s",
			n, stripTargetIP(ip, err.Error()))
		if f.abort != nil {
			f.abort()
		}
		f.hooks.callOnError(fmt.Errorf(
			"scan aborted after %d consecutive probe environment failures: %w",
			n, err,
		))
	})
}

// resetEnv clears the consecutive environment-failure streak after a
// successful probe.
func (f *probeFailures) resetEnv() {
	f.envFails.Store(0)
}
