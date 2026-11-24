package netutil

import (
	"context"
	"fmt"
	"net/netip"
	"strconv"
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
// ParseIPOrCIDR returns a netip.Prefix and true on success. A single IP
// comes back as a /32 or /128 prefix. Bad input returns false.
func ParseIPOrCIDR(raw string) (netip.Prefix, bool) {
	raw = strings.TrimSpace(raw)
	if prefix, err := netip.ParsePrefix(raw); err == nil {
		return prefix.Masked(), true
	}
	if addr, err := netip.ParseAddr(raw); err == nil {
		return netip.PrefixFrom(addr, addr.BitLen()), true
	}
	return netip.Prefix{}, false
}

// Loopback returns the loopback address matching the family of addr:
// ::1 for IPv6 (including IPv4-mapped IPv6), 127.0.0.1 otherwise.
func Loopback(addr netip.Addr) netip.Addr {
	if addr.Is6() {
		return netip.IPv6Loopback()
	}
	return netip.AddrFrom4([4]byte{127, 0, 0, 1})
}

// ParsePort parses a port field and reports whether the value is valid.
// An empty field is valid and returns 0, which means the scan settings
// choose the port.
func ParsePort(raw string) (uint16, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, true
	}
	port, err := strconv.ParseUint(raw, 10, 16)
	if err != nil {
		return 0, false
	}
	return uint16(port), true
}
