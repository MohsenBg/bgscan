// Package speedtest measures latency and transfer speed over HTTP.
// A Service reaches its test endpoints either through a local SOCKS5
// proxy port or through a DialContext the caller supplies, which is
// how probes measure through a tunnel. Speeds use BitsPerSec, and
// the config types set the byte counts, timeouts, and limits for
// each measurement.
package speedtest
