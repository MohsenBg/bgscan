package netutil

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// StreamCIDR iterates all IP addresses in a CIDR range and sends them to out.
// Iteration stops when:
//   - The context is canceled
//   - The CIDR range is exhausted
//   - The limit is reached (limit <= 0 means no limit)
//
// For large IPv6 ranges, use limit to avoid extremely long iterations.
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

// ParseIP parses a single IP address (no CIDR allowed).
// It explicitly rejects inputs that contain a slash (/), e.g. "192.168.0.0/24".
// Returns nil if the input is invalid.
func ParseIP(input string) net.IP {
	input = strings.TrimSpace(input)

	// Quick reject CIDR-form input
	if strings.Contains(input, "/") {
		return nil
	}

	ip := net.ParseIP(input)
	return ip
}

// ParseCIDR parses a CIDR string and returns its *net.IPNet.
// It returns a detailed error if parsing fails.
func ParseCIDR(input string) (*net.IPNet, error) {
	_, netw, err := net.ParseCIDR(strings.TrimSpace(input))
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR %q: %w", input, err)
	}
	return netw, nil
}

// NormalizeIPOrCIDR validates and normalizes an IP or CIDR input string.
// Returns the normalized string (canonical form) and a boolean indicating validity.
//
// Examples:
//
//	NormalizeIPOrCIDR("192.168.001.001") → "192.168.1.1", true
//	NormalizeIPOrCIDR("10.0.0.0/8") → "10.0.0.0/8", true
//	NormalizeIPOrCIDR("invalid") → "", false
func NormalizeIPOrCIDR(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)

	// Try CIDR first
	if _, _, err := net.ParseCIDR(raw); err == nil {
		return raw, true
	}

	// Clean leading zeros from IPv4 addresses before parsing
	cleanedRaw := removeIPv4LeadingZeros(raw)

	// Try plain IP
	ip := net.ParseIP(cleanedRaw)
	if ip == nil {
		return "", false
	}

	return ip.String(), true
}

// removeIPv4LeadingZeros strips leading zeros from IPv4 address octets.
// If the string is not a 4-part dotted decimal, it returns the original string.
func removeIPv4LeadingZeros(ipStr string) string {
	parts := strings.Split(ipStr, ".")
	if len(parts) != 4 {
		return ipStr
	}

	for i, part := range parts {
		if num, err := strconv.Atoi(part); err == nil {
			parts[i] = strconv.Itoa(num)
		}
	}
	return strings.Join(parts, ".")
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
