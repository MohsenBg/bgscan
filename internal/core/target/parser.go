package target

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/netip"
	"strconv"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
	"github.com/MohsenBg/bgscan/internal/core/netutil"
	"github.com/MohsenBg/bgscan/internal/logger"
)

// DefaultCSVConfig is the format every target list is read and written with.
var DefaultCSVConfig = fileutil.CSVConfig{
	Comma:            ',',
	HasHeader:        false,
	FieldsPerRecord:  -1,
	LazyQuotes:       true,
	TrimLeadingSpace: true,
}

// StreamCIDR expands a prefix into addresses, stopping at limit or ctx cancel.
// count is shared with the caller so several prefixes can share one limit.
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

// ParseRecord turns a raw CSV row into an Entry.
//
// Accepted layouts:
//
//	value,enable               legacy
//	value,port,enable          transitional
//	value,port,path,enable     canonical
func ParseRecord(rec []string) (Entry, bool) {
	if len(rec) < 1 || len(rec) > 4 {
		return Entry{}, false
	}
	raw := strings.TrimSpace(rec[0])
	if raw == "" {
		return Entry{}, false
	}

	var entry Entry
	if ip, ok := ParseIPOrCIDR(raw); ok {
		entry = NewEnabled(ip)
	} else {
		if netutil.ValidateDomain(raw) != nil {
			return Entry{}, false
		}
		domain, err := netutil.NormalizeHostWithSuffix(raw)
		if err != nil {
			return Entry{}, false
		}
		entry = NewDomain(strings.ToLower(domain), 1)
	}

	// Column positions depend on which layout the row uses.
	portCol, pathCol, enableCol := -1, -1, -1
	switch len(rec) {
	case 2:
		enableCol = 1
	case 3:
		portCol, enableCol = 1, 2
	case 4:
		portCol, pathCol, enableCol = 1, 2, 3
	}

	if portCol >= 0 {
		port, ok := parsePort(rec[portCol])
		if !ok {
			return Entry{}, false
		}
		entry.Port = port
	}
	if pathCol >= 0 {
		entry.Path = strings.TrimSpace(rec[pathCol])
		if entry.Path != "" && !strings.HasPrefix(entry.Path, "/") {
			return Entry{}, false
		}
	}
	enable := 1
	if enableCol >= 0 {
		v, ok := parseEnable(rec[enableCol])
		if !ok {
			return Entry{}, false
		}
		enable = v
	}
	entry.Enable = enable == 1
	return entry, true
}

// ParseIPOrCIDR accepts a CIDR or a plain IP; plain IPs become /32 or /128.
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

// parsePort treats an empty field as "use the scan settings" (0).
func parsePort(raw string) (uint16, bool) {
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

// parseEnable requires an integer; anything but 1 means disabled.
func parseEnable(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}
	enable, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return enable, true
}

// StreamActive sends every enabled address in the list to out.
// limit 0 means unlimited. Domains are rejected: they need DNS first.
func StreamActive(log *logger.Logger, ctx context.Context, path string, limit uint64, shuffled bool, out chan<- netip.Addr) error {
	if limit == 0 {
		limit = math.MaxUint64
	}
	if shuffled {
		return streamActiveShuffled(log, ctx, path, limit, out)
	}
	return streamActiveSequential(log, ctx, path, limit, out)
}

func streamActiveSequential(log *logger.Logger, ctx context.Context, path string, limit uint64, out chan<- netip.Addr) error {
	var count uint64
	warnedPort := false
	return ReadCSV(path, func(row Entry, _ int64) error {
		if !row.Enable {
			return nil
		}
		if row.Kind != KindIP {
			return fmt.Errorf("domain target %q requires DNS resolution, which is not implemented yet", row.Domain)
		}
		if row.Path != "" {
			return fmt.Errorf("path %q is not supported by scans yet", row.Path)
		}
		if row.Port != 0 && !warnedPort {
			warnedPort = true
			if log != nil {
				log.Warn(
					"list %q sets per-target port %d; per-target ports are not used yet, scan settings are used",
					path, row.Port,
				)
			}
		}
		if limit > 0 && count >= limit {
			return io.EOF // tells ReadCSV to stop early
		}
		if !row.IsSingle() {
			return StreamCIDR(ctx, row.IP, limit, &count, out)
		}
		select {
		case out <- row.IP.Addr():
			count++
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
}
