package config

import "time"

// DNSConfig represents the top‑level DNS configuration, combining resolver,
// DNSTT, and SlipStream settings.
type DNSConfig struct {
	Resolver   *ResolverConfig   `toml:"resolver"`
	DNSTT      *DNSTTConfig      `toml:"dnstt"`
	SlipStream *SlipStreamConfig `toml:"slip_stream"`
}

// ResolverConfig defines settings for traditional DNS resolvers.
type ResolverConfig struct {
	Workers         int        `toml:"workers"`
	Protocol        string     `toml:"protocol"`
	Domain          string     `toml:"domain"`
	Port            uint16     `toml:"port"`
	CheckTypes      []string   `toml:"check_types"`
	EDNSBufSize     uint16     `toml:"ends_buffer_size"`
	Timeout         DurationMS `toml:"timeout"`
	Tries           int        `toml:"tries"`
	RandomSubdomain bool       `toml:"random_subdomain"`
	AcceptedRCodes  []string   `toml:"accepted_rcodes"`
	CheckDPI        bool       `toml:"check_dpi"`
	DPITimeout      DurationMS `toml:"dpi_timeout"`
	DPITries        int        `toml:"dpi_tries"`
	PrefixOutput    string     `toml:"prefix_output"`
}

// DNSTTConfig defines configuration for DNSTT (DNS Tunnel Transport) scanning.
type DNSTTConfig struct {
	Enabled      bool       `toml:"enabled"`
	Workers      int        `toml:"workers"`
	Domain       string     `toml:"domain"`
	PublicKey    string     `toml:"public_key"`
	Timeout      DurationMS `toml:"timeout"`
	PrefixOutput string     `toml:"prefix_output"`
}

// SlipStreamConfig defines configuration for SlipStream-based DNS scanning.
type SlipStreamConfig struct {
	Enabled      bool       `toml:"enabled"`
	Workers      int        `toml:"workers"`
	Domain       string     `toml:"domain"`
	CertPath     string     `toml:"cert_path"`
	Timeout      DurationMS `toml:"timeout"`
	PrefixOutput string     `toml:"prefix_output"`
}

// DefaultDNSConfig returns the default configuration for DNS‑based scanning methods.
func DefaultDNSConfig() *DNSConfig {
	return &DNSConfig{
		Resolver: &ResolverConfig{
			Workers:         100,
			Protocol:        "udp",
			Domain:          "google.com",
			Port:            53,
			CheckTypes:      []string{"A"},
			EDNSBufSize:     1234,
			Timeout:         NewDurationMS(2 * time.Second),
			Tries:           1,
			RandomSubdomain: true,
			AcceptedRCodes:  []string{"noerror", "nxdomain"},
			CheckDPI:        true,
			DPITimeout:      NewDurationMS(500 * time.Millisecond),
			DPITries:        2,
			PrefixOutput:    "dns_resolver_",
		},
		DNSTT: &DNSTTConfig{
			Enabled:      false,
			Workers:      20,
			Domain:       "ns.example.com",
			PublicKey:    "",
			Timeout:      NewDurationMS(8 * time.Second),
			PrefixOutput: "dns_dnstt_",
		},
		SlipStream: &SlipStreamConfig{
			Enabled:      false,
			Workers:      20,
			Domain:       "ns.example.com",
			CertPath:     "",
			Timeout:      NewDurationMS(8 * time.Second),
			PrefixOutput: "dns_slipstream_",
		},
	}
}
