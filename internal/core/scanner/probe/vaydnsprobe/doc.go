// Package vaydnsprobe checks connectivity through a VayDNS tunnel.
// VayDNSProbe opens a tunnel that uses the target as its DNS
// resolver, connects an SSH or SOCKS proxy through it, and measures
// latency through that proxy. VayDNSResult is the result type, and
// Schema describes its stored record layout.
package vaydnsprobe
