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

// StreamCIDR expands row's prefix into single-address entries, stopping at
// limit or ctx cancel. Port, path, and enable are copied onto every address.
// count is shared with the caller so several prefixes can share one limit.
func StreamCIDR(ctx context.Context, row Entry, limit uint64, count *uint64, out chan<- Entry) error {
	prefix := row.IP.Masked()
	curr := prefix.Addr()

	for prefix.Contains(curr) {
		if limit > 0 && *count >= limit {
			return nil
		}
		row.IP = netip.PrefixFrom(curr, curr.BitLen())

		select {
		case out <- row:
			(*count)++
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

func ParseRecord(rec []string) (Entry, error) {
	if len(rec) < 4 {
		return Entry{}, fmt.Errorf("invalid record length: expected at least 4 fields, got %d", len(rec))
	}

	raw := strings.TrimSpace(rec[0])
	if raw == "" {
		return Entry{}, fmt.Errorf("empty target")
	}

	var entry Entry
	if ip, ok := netutil.ParseIPOrCIDR(raw); ok {
		entry = NewEnabled(ip)
	} else {
		if err := netutil.ValidateDomain(raw); err != nil {
			return Entry{}, fmt.Errorf("invalid domain %q: %w", raw, err)
		}

		domain, err := netutil.NormalizeHostWithSuffix(raw)
		if err != nil {
			return Entry{}, fmt.Errorf("normalize domain %q: %w", raw, err)
		}

		entry = NewDomain(strings.ToLower(domain), 1)
	}

	port, ok := netutil.ParsePort(strings.TrimSpace(rec[1]))
	if !ok {
		return Entry{}, fmt.Errorf("invalid port %q", rec[1])
	}
	entry.Port = port

	entry.Path = netutil.NormalizePath(rec[2])
	if err := netutil.ValidatePath(entry.Path); err != nil {
		return Entry{}, fmt.Errorf("invalid path %q: %w", rec[2], err)
	}

	enable, ok := parseEnable(strings.TrimSpace(rec[3]))
	if !ok {
		return Entry{}, fmt.Errorf("invalid enable value %q", rec[3])
	}
	entry.Enable = enable == 1

	return entry, nil
}

// ParseInputRecord parses a user-supplied row.
//
// Stored target lists use ParseRecord and always carry four columns. Import is
// more forgiving because people paste plain IPs and older exports:
//
//	value               -> value,0,,1
//	value,enable        -> value,0,,enable
//	value,port,enable   -> value,port,,enable
//	value,port,path,enable
func ParseInputRecord(rec []string) (Entry, error) {
	switch len(rec) {
	case 1:
		return ParseRecord([]string{rec[0], "", "", "1"})
	case 2:
		return ParseRecord([]string{rec[0], "", "", rec[1]})
	case 3:
		return ParseRecord([]string{rec[0], rec[1], "", rec[2]})
	case 4:
		return ParseRecord(rec)
	default:
		return Entry{}, fmt.Errorf("invalid record length: expected 1 to 4 fields, got %d", len(rec))
	}
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

// StreamActive sends one Entry per enabled target to out, keeping each row's
// port and path. A domain and a single address are one target; a CIDR is
// expanded into one target per address. Resolving a domain is the caller's
// job: the stream only says what the user asked for. limit 0 means unlimited.
func StreamActive(log *logger.Logger, ctx context.Context, path string, limit uint64, shuffled bool, out chan<- Entry) error {
	if limit == 0 {
		limit = math.MaxUint64
	}
	if shuffled {
		return streamActiveShuffled(log, ctx, path, limit, out)
	}
	return streamActiveSequential(ctx, path, limit, out)
}

func streamActiveSequential(ctx context.Context, path string, limit uint64, out chan<- Entry) error {
	var count uint64
	return ReadCSV(path, func(row Entry, _ int64) error {
		if !row.Enable {
			return nil
		}
		if limit > 0 && count >= limit {
			return io.EOF
		}
		if row.Kind == KindIP && !row.IsSingle() {
			return StreamCIDR(ctx, row, limit, &count, out)
		}
		select {
		case out <- row:
			count++
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
}
