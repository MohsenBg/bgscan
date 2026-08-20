// Package core initializes bgscan internal components by registering
// all built-in probe result schemas with the global result registry.
package core

import (
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/probe/dnsttprobe"
	"bgscan/internal/core/scanner/probe/httpprobe"
	"bgscan/internal/core/scanner/probe/icmpprobe"
	"bgscan/internal/core/scanner/probe/resolveprobe"
	"bgscan/internal/core/scanner/probe/slipstreamprobe"
	"bgscan/internal/core/scanner/probe/tcpprobe"
	"bgscan/internal/core/scanner/probe/vaydnsprobe"
	"bgscan/internal/core/scanner/probe/xrayprobe"
)

// Init registers all built-in probe result schemas with the global registry.
func Init() error {
	if err := result.DefaultRegistry.Register(icmpprobe.Schema); err != nil {
		return err
	}

	if err := result.DefaultRegistry.Register(tcpprobe.Schema); err != nil {
		return err
	}

	if err := result.DefaultRegistry.Register(httpprobe.Schema); err != nil {
		return err
	}

	if err := result.DefaultRegistry.Register(resolveprobe.Schema); err != nil {
		return err
	}

	if err := result.DefaultRegistry.Register(dnsttprobe.Schema); err != nil {
		return err
	}

	if err := result.DefaultRegistry.Register(vaydnsprobe.Schema); err != nil {
		return err
	}

	if err := result.DefaultRegistry.Register(slipstreamprobe.Schema); err != nil {
		return err
	}

	if err := result.DefaultRegistry.Register(xrayprobe.Schema); err != nil {
		return err
	}

	return nil
}
