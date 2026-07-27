package netutil

import (
	"context"
	"fmt"
	"net/netip"
	"strings"
)

// StreamCIDR iterates through all IP addresses in a given CIDR range and sends them to the out channel.
// The iteration is efficient and thread-safe, stopping when:
//   - The context is canceled.
//   - The CIDR range is fully exhausted.
//   - The specified limit is reached (a limit of 0 indicates no limit).
//
// Note: For large IPv6 ranges, a limit should be provided to prevent near-infinite loops.
func StreamCIDR(ctx context.Context, cidr string, limit uint64, out chan<- netip.Addr) error {
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return fmt.Errorf("parse CIDR %q: %w", cidr, err)
	}

	curr := prefix.Masked().Addr()

	var count uint64
	for prefix.Contains(curr) {
		if limit > 0 && count >= limit {
			return nil
		}

		select {
		case out <- curr:
			count++
		case <-ctx.Done():
			return ctx.Err()
		}

		curr = curr.Next()

		if !curr.IsValid() {
			break
		}
	}

	return nil
}

// ParseIPOrCIDR validates and normalizes an IP or CIDR input string.
// It returns a netip.Prefix, which can represent both networks and single IPs (as /32 or /128).
func ParseIPOrCIDR(raw string) (netip.Prefix, bool) {
	raw = strings.TrimSpace(raw)

	// 1. Try CIDR first
	if prefix, err := netip.ParsePrefix(raw); err == nil {
		return prefix, true
	}

	// 2. Try plain IP and convert it to a single-host prefix
	if addr, err := netip.ParseAddr(raw); err == nil {
		bits := 128
		if addr.Is4() {
			bits = 32
		}
		return netip.PrefixFrom(addr, bits), true
	}

	return netip.Prefix{}, false
}
