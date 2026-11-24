// Package masterdnsprobe checks targets by running a MasterDNS tunnel
// through them. The probe starts the embedded MasterDNS client with the
// target as resolver, waits for its local SOCKS5 listener, and measures
// latency through that listener. Each Run leases a local port and
// closes its own tunnel when it returns.
package masterdnsprobe
