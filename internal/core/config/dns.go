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
	Workers         int        `toml:"workers" comment:"Concurrent DNS workers. Range: 1-2500. Higher = faster but more CPU/network."`
	Transport       string     `toml:"protocol" comment:"DNS transport: udp (fastest), tcp (reliable), dot (DNS-over-TLS, encrypted)."`
	Domain          string     `toml:"domain" comment:"Domain to query. Used as the base domain for resolver tests."`
	Port            uint16     `toml:"port" comment:"DNS server port. Range: 1-65535. Standard: 53."`
	CheckTypes      []string   `toml:"check_types" comment:"DNS record types to query. Common: A, AAAA, TXT. Use TXT for DNSTT compatibility."`
	EDNSBufSize     uint16     `toml:"edns_buffer_size" comment:"EDNS buffer size in bytes. 0 = disabled. Standard = 1232. Only affects DNS scans."`
	Timeout         DurationMS `toml:"timeout" comment:"Max wait for DNS response, in ms. Range: 100-30000. Lower = faster but may miss slow resolvers."`
	Tries           int        `toml:"tries" comment:"Retry attempts per target. Range: 1-10. Only retries on network errors, not on bad responses."`
	RandomSubdomain bool       `toml:"random_subdomain" comment:"Add random prefix to domain. Prevents resolver caching and forces fresh lookups."`
	AcceptedRCodes  []string   `toml:"accepted_rcodes" comment:"DNS response codes treated as success. Common: NOERROR, NXDOMAIN, SERVFAIL."`
	OutputPrefix    string     `toml:"output_prefix" comment:"Filename prefix for result files."`
	DPI             DPIConfig  `toml:"dpi"`
}

// DPIConfig defines configuration for DNS DPI detection.
type DPIConfig struct {
	Enabled bool       `toml:"enabled" comment:"Enable anti-hijacking check. Queries a fake .invalid domain and flags resolvers that return success."`
	Timeout DurationMS `toml:"timeout" comment:"Max wait for DPI test, in ms. Range: 100-10000. Should be shorter than main timeout."`
	Tries   int        `toml:"tries" comment:"DPI verification attempts. Range: 1-10. Higher = more reliable detection."`
}

// DNSTunneling defines configuration for DNS tunneling protocol tests.
type DNSTunneling struct {
	Workers          int        `toml:"workers" comment:"Concurrent tunnel test workers. Range: 1-500. Higher = faster but more bandwidth."`
	Tries            int        `toml:"tries" comment:"Retry attempts per target. Range: 1-10."`
	Timeout          DurationMS `toml:"timeout" comment:"Max wait for tunnel test, in ms. Range: 100-60000. Tunnel tests need more time."`
	CheckDNSResolver bool       `toml:"check_dns_resolver" comment:"Test DNS resolver before tunnel. true = chain resolver scan first, false = test tunnel directly."`
	AdaptiveResolver bool       `toml:"adaptive_resolver" comment:"When true, overrides resolver settings (type, port, domain) to match the DNS tunnel config. When false, uses the Resolver scanner settings as-is."`
	OutputPrefix     string     `toml:"output_prefix" comment:"Filename prefix for result files."`
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
	base.OutputPrefix = "dns_tun_"
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
