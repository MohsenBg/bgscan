package iplist

import (
	"context"
	"io"
	"math"
	"net/netip"
	"strconv"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
)

// DefaultCSVConfig defines the canonical format for CSV inputs.
var DefaultCSVConfig = fileutil.CSVConfig{
	Comma:            ',',
	HasHeader:        false,
	FieldsPerRecord:  -1,
	LazyQuotes:       true,
	TrimLeadingSpace: true,
}

// StreamCIDR now accepts a pre-parsed netip.Prefix for performance.
// It updates the count pointer to allow the parent caller to track total progress against limits.
func StreamCIDR(ctx context.Context, prefix netip.Prefix, limit uint64, count *uint64, out chan<- netip.Addr) error {
	prefix = prefix.Masked()
	curr := prefix.Addr()

	for prefix.Contains(curr) {
		if limit > 0 && *count >= limit {
			return nil
		}

		select {
		case out <- curr:
			*count++
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

// ParseRecord converts a raw CSV row into an IPList structure.
func ParseRecord(rec []string) (IPList, bool) {
	if len(rec) == 0 {
		return IPList{}, false
	}

	raw := strings.TrimSpace(rec[0])
	ip, ok := ParseIPOrCIDR(raw)
	if !ok {
		return IPList{}, false
	}

	enable := 1
	if len(rec) > 1 {
		if v, err := strconv.Atoi(strings.TrimSpace(rec[1])); err == nil {
			enable = v
		}
	}

	return New(ip.Masked(), enable), true
}

// ParseIPOrCIDR leverages netip.Prefix for all cases (single IPs become /32 or /128).
func ParseIPOrCIDR(raw string) (netip.Prefix, bool) {
	raw = strings.TrimSpace(raw)

	// 1. Try CIDR first
	if prefix, err := netip.ParsePrefix(raw); err == nil {
		return prefix.Masked(), true
	}

	// 2. Try plain IP and convert it to a single-host prefix (/32 or /128 automatically)
	if addr, err := netip.ParseAddr(raw); err == nil {
		return netip.PrefixFrom(addr, addr.BitLen()), true
	}

	return netip.Prefix{}, false
}

// StreamActiveIPs now outputs netip.Addr to avoid string allocations in the scan loop.
func StreamActiveIPs(ctx context.Context, path string, limit uint64, shuffled bool, out chan<- netip.Addr) error {
	if limit == 0 {
		limit = math.MaxUint64 - 1
	}

	if shuffled {
		// Ensure streamActiveIPsShuffled also uses chan<- netip.Addr
		return streamActiveIPsShuffled(ctx, path, limit, out)
	}
	return streamActiveIPsSequential(ctx, path, limit, out)
}

// streamActiveIPsSequential coordinates the CSV reading and CIDR expansion.
func streamActiveIPsSequential(ctx context.Context, path string, limit uint64, out chan<- netip.Addr) error {
	var count uint64 = 0

	return ReadCSV(path, func(row IPList, _ int64) error {
		if !row.Enable {
			return nil
		}

		if limit > 0 && count >= limit {
			return io.EOF // Stop reading CSV if limit reached
		}

		// Check if it's a network (e.g. /24) or a single host (/32 or /128)
		// Assuming row.IP is a netip.Prefix
		if !row.IsSingle() {
			return StreamCIDR(ctx, row.IP, limit, &count, out)
		}

		// Single IP case
		select {
		case out <- row.IP.Addr():
			count++
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
}
