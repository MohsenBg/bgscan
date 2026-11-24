// Package tcpprobe probes host targets with plain TCP connections.
// TCPProbe reports whether a handshake to its fixed port succeeds and
// how long it takes. It checks no application protocol, so a success
// only shows that something accepts TCP on that port.
package tcpprobe
