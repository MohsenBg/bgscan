// Package icmpprobe probes IP targets with ICMP echo requests.
// ICMPProbe sends pings over shared sockets and matches replies to
// waiting callers by ID and sequence number. Call Init before Run so
// the reader goroutines run, and call Close to stop them.
package icmpprobe
