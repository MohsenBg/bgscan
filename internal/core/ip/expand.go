package ip

import (
	"context"
	"fmt"
	"net"
)

// StreamCIDR sequentially iterates all IP addresses within the given CIDR
// and sends each address (as a string) into the provided output channel.
// Iteration stops when:
//
//   - The context is canceled
//   - The CIDR range is fully exhausted
//   - The optional limit is reached (limit <= 0 means “no limit”)
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	ch := make(chan string)
//	go func() {
//	    defer close(ch)
//	    StreamCIDR(ctx, "192.168.1.0/30", 0, ch)
//	}()
//
//	for ip := range ch {
//	    fmt.Println(ip)
//	}
//
// Note: IPv6 CIDRs broader than /64 can represent astronomically large
// spaces. Use caller‑side limiting (limit parameter or RangeLimits) to
// avoid memory exhaustion or runaway loops.
func StreamCIDR(ctx context.Context, cidr string, limit uint64, out chan<- string) error {
	ip, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("parse CIDR %q: %w", cidr, err)
	}

	start := masked(ip, network.Mask)

	var count uint64 = 0
	for curr := start; network.Contains(curr); curr = increment(curr) {

		// Limit reached (if enabled).
		if limit > 0 && count >= limit {
			return nil
		}

		// Context cancellation takes priority.
		select {
		case out <- curr.String():
			count++
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// increment returns the next IP address after the given one.
// Works for both IPv4 and IPv6 by incrementing the byte slice directly.
func increment(ip net.IP) net.IP {
	next := make(net.IP, len(ip))
	copy(next, ip)

	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] != 0 {
			break
		}
	}
	return next
}

// masked returns a new IP with the given subnet mask applied.
// The returned slice is always a fresh copy.
func masked(ip net.IP, mask net.IPMask) net.IP {
	m := ip.Mask(mask)
	out := make(net.IP, len(m))
	copy(out, m)
	return out
}
