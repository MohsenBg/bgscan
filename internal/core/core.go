package core

import (
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner/probe/dnsttprobe"
	"bgscan/internal/core/scanner/probe/httpprobe"
	"bgscan/internal/core/scanner/probe/icmpprobe"
	"bgscan/internal/core/scanner/probe/resolveprobe"
	"bgscan/internal/core/scanner/probe/slipstreamprobe"
	"bgscan/internal/core/scanner/probe/tcpprobe"
	"bgscan/internal/core/scanner/probe/xrayprobe"
)

// Init registers the result schemas of all built-in probe types
// into the global result.DefaultRegistry.
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

	if err := result.DefaultRegistry.Register(slipstreamprobe.Schema); err != nil {
		return err
	}

	if err := result.DefaultRegistry.Register(xrayprobe.Schema); err != nil {
		return err
	}
	return nil
}
