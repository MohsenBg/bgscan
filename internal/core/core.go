// Package core initializes bgscan internal components by registering
// all built-in probe result schemas with the global result registry.
package core

import (
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/dnsttprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/httpprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/icmpprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/resolveprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/slipstreamprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/tcpprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/vaydnsprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/xrayprobe"
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
