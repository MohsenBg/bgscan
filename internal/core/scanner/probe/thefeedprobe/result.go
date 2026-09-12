package thefeedprobe

import (
	"fmt"
	"net/netip"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/result"
)

// Schema defines the database layout and parsing rules for TheFeed probe
// outcomes.
var Schema = result.ResultSchema{
	Name:      "TheFeed",
	Directory: "thefeed",

	Columns: []result.ColumnDef{
		{Name: "IP", Width: 30},
		{Name: "Latency", Width: 15},
		{Name: "Domain", Width: 35},
		{Name: "Mode", Width: 10},
	},

	Parser: parseTheFeedResult,
}

// TheFeedResult holds the outcome of a single TheFeed resolver probe.
//
// Latency measures only the DNS exchange: from the query send until the
// encrypted TXT response is decrypted successfully.
type TheFeedResult struct {
	IP        netip.Addr
	Latency   time.Duration
	Domain    string
	QueryMode string
}

func (r TheFeedResult) Key() string {
	return r.IP.String()
}

func (r TheFeedResult) KeyType() result.KeyType {
	return result.KeyIP
}

func (r TheFeedResult) Equal(rs result.Result) bool {
	return r.IP.String() == rs.Key()
}

func (r TheFeedResult) ToRecord() []string {
	return []string{
		r.IP.String(),
		result.FormatDuration(r.Latency),
		r.Domain,
		r.QueryMode,
	}
}

// Score calculates a performance rating where lower latency yields a higher
// score. Latency is clamped to 1ms to prevent division by zero.
func (r TheFeedResult) Score() float64 {
	ms := float64(r.Latency.Milliseconds())
	if ms < 1 {
		ms = 1
	}
	return 1000.0 / ms
}

func parseTheFeedResult(record []string) (result.Result, error) {
	if len(record) < 4 {
		return nil, fmt.Errorf(
			"invalid TheFeed result record: expected 4 fields, got %d",
			len(record),
		)
	}

	ip, err := netip.ParseAddr(record[0])
	if err != nil {
		return nil, fmt.Errorf("parse IP: %w", err)
	}

	latency, err := time.ParseDuration(record[1])
	if err != nil {
		return nil, fmt.Errorf("parse latency: %w", err)
	}

	return TheFeedResult{
		IP:        ip,
		Latency:   latency,
		Domain:    record[2],
		QueryMode: record[3],
	}, nil
}
