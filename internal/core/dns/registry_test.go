package dns

import (
	"strings"
	"testing"
)

func TestProxyLabel(t *testing.T) {
	tests := []struct {
		proxyType ResolverProxyType
		auth      AuthMethod
		want      string
	}{
		{ResolverProxySOCKS, AuthNone, "socks-none"},
		{ResolverProxySOCKS, AuthPassword, "socks-password"},
		{ResolverProxySSH, AuthKey, "ssh-key"},
		{"", "", "-"},
	}

	for _, tt := range tests {
		if got := proxyLabel(tt.proxyType, tt.auth); got != tt.want {
			t.Errorf("proxyLabel(%q, %q) = %q, want %q", tt.proxyType, tt.auth, got, tt.want)
		}
	}
}

func TestRenameDNSTunConfigFileUnsupportedProtocol(t *testing.T) {
	err := RenameDNSTunConfigFile(DNSTunConfigFile{
		Protocol: DNSTunProtocol("carrier-pigeon"),
	}, "new-name")

	if err == nil || !strings.Contains(err.Error(), "unsupported DNS tunnel protocol") {
		t.Fatalf("RenameDNSTunConfigFile() error = %v, want unsupported-protocol error", err)
	}
}

func TestDNSTunProtocolConstants(t *testing.T) {
	want := map[DNSTunProtocol]string{
		DNSTunProtocolVayDNS:     "vaydns",
		DNSTunProtocolDNSTT:      "dnstt",
		DNSTunProtocolSlipstream: "slipstream",
		DNSTunProtocolMasterDNS:  "masterdns",
		DNSTunProtocolStormDNS:   "stormdns",
	}

	for protocol, value := range want {
		if string(protocol) != value {
			t.Errorf("protocol %q != %q", protocol, value)
		}
	}
}

func TestConfigFileAliases(t *testing.T) {
	// Passing an alias value where ConfigFile[C] is expected only
	// compiles when the alias is identical to the generic type, so
	// this pins every per-protocol alias to the shared store's shape.
	assertConfigFileAlias(VayDNSConfigFile{})
	assertConfigFileAlias(DNSTTConfigFile{})
	assertConfigFileAlias(SlipstreamConfigFile{})
	assertConfigFileAlias(MasterDNSConfigFile{})
	assertConfigFileAlias(StormDNSConfigFile{})
}

func assertConfigFileAlias[C any](_ ConfigFile[C]) {}
