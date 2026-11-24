// Package xray builds and runs Xray configs for probing.
//
// The package holds typed forms of the Xray config pieces bgscan uses,
// loads outbound templates from disk or from sharing links, points a
// template at a scanned address, and starts the result on the embedded
// Xray core behind a local SOCKS inbound.
package xray
