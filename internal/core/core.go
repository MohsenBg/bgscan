// Package core initializes bgscan internal components by registering
// all built-in probe result schemas with the global result registry.
package core

import (
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/dnsttprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/httpprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/icmpprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/masterdnsprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/resolveprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/slipstreamprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/stormdnsprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/tcpprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/thefeedprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/vaydnsprobe"
	"github.com/MohsenBg/bgscan/internal/core/scanner/probe/xrayprobe"
)

// Init registers all built-in probe result schemas with the global
// registry. It returns an error when a schema fails validation or a
// schema for the same directory is already registered, so callers
// should run it once at startup.
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

	if err := result.DefaultRegistry.Register(thefeedprobe.Schema); err != nil {
		return err
	}

	if err := result.DefaultRegistry.Register(masterdnsprobe.Schema); err != nil {
		return err
	}

	if err := result.DefaultRegistry.Register(stormdnsprobe.Schema); err != nil {
		return err
	}

	if err := result.DefaultRegistry.Register(xrayprobe.Schema); err != nil {
		return err
	}

	return nil
}
