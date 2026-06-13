package result

import "time"

// IPScanResult represents the result of probing a single IP address.
//
// Latency, Download, and Upload may be zero when the corresponding
// measurement is not applicable for the probe type (for example,
// ICMP or TCP-only scans).
type IPScanResult struct {
	IP       string        // Target IP address.
	Latency  time.Duration // Round‑trip or connection latency.
	Download time.Duration // Optional download measurement.
	Upload   time.Duration // Optional upload measurement.
}

// Config controls the behavior of the asynchronous result writer,
// including batching, channel buffering, and flush frequency.
type Config struct {
	MergeFlushInterval time.Duration // How often accumulated results are merged.
	ChanSize           int           // Capacity of the internal result channel.
	BatchSize          int           // Max results buffered before merge.
}

// Recommended defaults and safety bounds.
const (
	FallbackLatency       = 999 * time.Second      // Used when latency cannot be determined.
	DefaultChanSize       = 1024                   // Default result channel capacity.
	DefaultBatchSize      = 4096                   // Default batch capacity.
	MinMergeFlushInterval = 120 * time.Millisecond // Minimum allowed flush interval.
)

// DefaultConfig returns a Config populated with safe operational defaults.
func DefaultConfig() Config {
	return Config{
		MergeFlushInterval: MinMergeFlushInterval,
		ChanSize:           DefaultChanSize,
		BatchSize:          DefaultBatchSize,
	}
}

// Normalize ensures configuration values fall within safe bounds and
// applies defaults when fields are unset.
func (c *Config) Normalize() {
	if c.MergeFlushInterval < MinMergeFlushInterval {
		c.MergeFlushInterval = MinMergeFlushInterval
	}
	if c.ChanSize <= 0 {
		c.ChanSize = DefaultChanSize
	}
	if c.BatchSize <= 0 {
		c.BatchSize = DefaultBatchSize
	}
}

// ResultType identifies the type of probe that produced a result file.
type ResultType int

const (
	ResultAll        ResultType = iota // Combined or unspecified results.
	ResultICMP                         // ICMP probe results.
	ResultTCP                          // TCP probe results.
	ResultHTTP                         // HTTP probe results.
	ResultXRAY                         // Xray probe results.
	ResultDNSTT                        // DNSTT probe results.
	ResultSLIPSTREAM                   // Slipstream probe results.
	ResultRESOLVE                      // Resolver probe results.
)

// String returns the textual representation of the ResultType.
// Unknown values return "unknown".
func (t ResultType) String() string {
	switch t {
	case ResultAll:
		return "all"
	case ResultICMP:
		return "icmp"
	case ResultTCP:
		return "tcp"
	case ResultHTTP:
		return "http"
	case ResultXRAY:
		return "xray"
	case ResultDNSTT:
		return "dnstt"
	case ResultSLIPSTREAM:
		return "slipstream"
	case ResultRESOLVE:
		return "resolve"
	default:
		return "unknown"
	}
}

// ResultFile describes metadata about a stored result file.
//
// IPCount indicates the number of IP entries contained in the file.
// A value of -1 means the count has not been computed yet.
type ResultFile struct {
	Name        string     // File name (without extension).
	SizeBytes   int64      // File size in bytes.
	CreatedTime time.Time  // File creation or last modification time.
	Type        ResultType // Classification of stored results.
	IPCount     int64      // Number of IP entries, or -1 if unknown.
	Path        string     // Absolute filesystem path.
}
