// Package slipstreamprobe checks targets by running a Slipstream DNS
// tunnel through them. The probe starts the tunnel with the target as
// resolver and measures latency through its local proxy listener,
// which uses SSH or SOCKS as the config sets. Each Run leases a local
// port and stops its own tunnel when it returns.
package slipstreamprobe
