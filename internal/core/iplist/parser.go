package iplist

import (
	"context"
	"io"
	"strconv"
	"strings"

	"bgscan/internal/core/fileutil"
	"bgscan/internal/core/ip"
)

// DefaultCSVConfig defines the canonical format for CSV inputs.
var DefaultCSVConfig = fileutil.CSVConfig{
	Comma:            ',',
	HasHeader:        false,
	FieldsPerRecord:  -1,
	LazyQuotes:       true,
	TrimLeadingSpace: true,
}

// ParseRecord converts a raw CSV row into an IPList structure.
func ParseRecord(rec []string) (IPList, bool) {
	if len(rec) == 0 {
		return IPList{}, false
	}

	raw := strings.TrimSpace(rec[0])
	ipStr, ok := ip.NormalizeIPOrCIDR(raw)
	if !ok {
		return IPList{}, false
	}

	enable := 1
	if len(rec) > 1 {
		if v, err := strconv.Atoi(strings.TrimSpace(rec[1])); err == nil {
			enable = v
		}
	}

	return New(ipStr, enable), true
}

// StreamActiveIPs provides the main entry point, branching into sequential or shuffled logic.
func StreamActiveIPs(ctx context.Context, path string, limit int, shuffled bool, out chan<- string) error {
	if shuffled {
		return streamActiveIPsShuffled(ctx, path, limit, out)
	}
	return streamActiveIPsSequential(ctx, path, limit, out)
}

// streamActiveIPsSequential loops sequentially, relying directly on step iterations.
func streamActiveIPsSequential(ctx context.Context, path string, limit int, out chan<- string) error {
	count := 0
	return ReadCSV(path, func(row IPList, _ int64) error {
		if !row.Enable {
			return nil
		}
		if limit > 0 && count >= limit {
			return io.EOF
		}
		if row.IsCIDR() {
			return ip.StreamCIDR(ctx, row.IP, limit-count, out)
		}

		select {
		case out <- row.IP:
			count++
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
}
