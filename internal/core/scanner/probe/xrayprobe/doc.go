// Package xrayprobe checks a target through a temporary local Xray
// SOCKS proxy. XrayProbe starts one Xray instance per target, then
// measures latency and, when the test mode asks for it, download
// and upload speed through the proxy. XrayResult is the result
// type, and Schema describes its stored record layout.
package xrayprobe
