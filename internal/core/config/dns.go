package config

import (
	"time"

	"bgscan/internal/logger"
)

// DNSConfig defines configuration for DNS resolver and DNS tunneling tests.
type DNSConfig struct {
	Resolver     ResolverConfig `toml:"resolver"`
	DNSTunneling DNSTunneling   `toml:"dns_tunneling"`
}

// ResolverConfig defines settings for DNS resolver testing.
type ResolverConfig struct {
	Workers         int        `toml:"workers" comment:"Number of concurrent DNS resolver tests."`
	Transport       string     `toml:"protocol" comment:"Transport protocol used for DNS queries, such as UDP or TCP."`
	Domain          string     `toml:"domain" comment:"Domain name used for DNS resolver tests."`
	Port            uint16     `toml:"port" comment:"DNS server port to query."`
	CheckTypes      []string   `toml:"check_types" comment:"DNS record types to query during resolver checks."`
	EDNSBufSize     uint16     `toml:"ends_buffer_size" comment:"EDNS buffer size advertised in DNS queries, in bytes."`
	Timeout         DurationMS `toml:"timeout" comment:"Maximum time to wait for a DNS query, in milliseconds."`
	Tries           int        `toml:"tries" comment:"Maximum number of DNS query attempts per target."`
	RandomSubdomain bool       `toml:"random_subdomain" comment:"Whether to use a random subdomain for DNS queries."`
	AcceptedRCodes  []string   `toml:"accepted_rcodes" comment:"DNS response codes accepted as successful resolver responses."`
	OutputPrefix    string     `toml:"output_prefix" comment:"Filename prefix for DNS resolver results."`
	DPI             DPIConfig  `toml:"dpi"`
}

// DPIConfig defines configuration for DNS DPI detection.
type DPIConfig struct {
	Enabled bool       `toml:"enabled" comment:"Whether DNS DPI detection is enabled."`
	Timeout DurationMS `toml:"timeout" comment:"Maximum time to wait for a DPI test, in milliseconds."`
	Tries   int        `toml:"tries" comment:"Maximum number of DPI test attempts per target."`
}

// DNSTunneling defines configuration for DNS tunneling protocol tests.
type DNSTunneling struct {
	Workers          int        `toml:"workers" comment:"Number of concurrent DNS tunneling tests."`
	Tries            int        `toml:"tries" comment:"Maximum number of attempts per target."`
	Timeout          DurationMS `toml:"timeout" comment:"Maximum time to wait for a DNS tunneling test, in milliseconds."`
	CheckDNSResolver bool       `toml:"check_dns_resolver" comment:"Check DNS resolver availability before testing the tunnel. Disable to test the tunnel directly."`
	OutputPrefix     string     `toml:"output_prefix" comment:"Filename prefix for DNS tunneling results."`
}

func DefaultDNSConfig() DNSConfig {
	platform := DetectPlatform()
	tier := SelectTier(CheckResources())

	tiers, ok := dnsDefaults[platform]
	if !ok {
		logger.CoreWarn("bgscan: no DNS defaults for platform %q, falling back to %q", platform, Desktop)
		platform = Desktop
		tiers = dnsDefaults[platform]
	}

	cfg, ok := tiers[tier]
	if !ok {
		logger.CoreWarn("bgscan: no DNS defaults for platform %q tier %q, falling back to %q", platform, tier, Mid)
		cfg = tiers[Mid]
	}

	return cfg
}

// resolverBase holds ResolverConfig fields shared across every platform/tier.
func resolverBase() ResolverConfig {
	return ResolverConfig{
		Transport:       "udp",
		Domain:          "example.com",
		Port:            53,
		CheckTypes:      []string{"TXT"},
		EDNSBufSize:     1232,
		RandomSubdomain: true,
		AcceptedRCodes:  []string{"NOERROR", "NXDOMAIN", "SERVFAIL"},
		OutputPrefix:    "dns_",
		DPI: DPIConfig{
			Enabled: true,
			Timeout: NewDurationMS(2 * time.Second),
			Tries:   1,
		},
	}
}

// withResolver returns a copy of base with Workers/Timeout/Tries overridden.
func withResolver(base ResolverConfig, workers, tries int, timeout time.Duration) ResolverConfig {
	base.Workers = workers
	base.Tries = tries
	base.Timeout = NewDurationMS(timeout)
	return base
}

// tunnelBase holds DNSTunneling fields shared across every platform/tier.
func tunnelBase() DNSTunneling {
	return DNSTunneling{
		CheckDNSResolver: true,
	}
}

// withTunnel returns a copy of base with Workers/Timeout/Tries overridden.
func withTunnel(base DNSTunneling, workers, tries int, timeout time.Duration) DNSTunneling {
	base.Workers = workers
	base.Tries = tries
	base.Timeout = NewDurationMS(timeout)
	base.OutputPrefix = "dns_tun"
	return base
}

var dnsDefaults = map[Platform]map[Tier]DNSConfig{
	Server: {
		Low: DNSConfig{
			Resolver:     withResolver(resolverBase(), 100, 1, 2*time.Second),
			DNSTunneling: withTunnel(tunnelBase(), 12, 1, 10*time.Second),
		},
		Mid: DNSConfig{
			Resolver:     withResolver(resolverBase(), 400, 1, 2*time.Second),
			DNSTunneling: withTunnel(tunnelBase(), 24, 1, 10*time.Second),
		},
		High: DNSConfig{
			Resolver:     withResolver(resolverBase(), 1000, 1, 1500*time.Millisecond),
			DNSTunneling: withTunnel(tunnelBase(), 64, 1, 8*time.Second),
		},
	},
	Desktop: {
		Low: DNSConfig{
			Resolver:     withResolver(resolverBase(), 30, 1, 2*time.Second),
			DNSTunneling: withTunnel(tunnelBase(), 8, 1, 10*time.Second),
		},
		Mid: DNSConfig{
			Resolver:     withResolver(resolverBase(), 150, 1, 2*time.Second),
			DNSTunneling: withTunnel(tunnelBase(), 16, 1, 10*time.Second),
		},
		High: DNSConfig{
			Resolver:     withResolver(resolverBase(), 300, 1, 2*time.Second),
			DNSTunneling: withTunnel(tunnelBase(), 32, 1, 10*time.Second),
		},
	},
	Android: {
		Low: DNSConfig{
			Resolver:     withResolver(resolverBase(), 15, 1, 3*time.Second),
			DNSTunneling: withTunnel(tunnelBase(), 3, 1, 10*time.Second),
		},
		Mid: DNSConfig{
			Resolver:     withResolver(resolverBase(), 60, 1, 3*time.Second),
			DNSTunneling: withTunnel(tunnelBase(), 6, 1, 10*time.Second),
		},
		High: DNSConfig{
			Resolver:     withResolver(resolverBase(), 100, 1, 2500*time.Millisecond),
			DNSTunneling: withTunnel(tunnelBase(), 12, 1, 10*time.Second),
		},
	},
}
