package dns

import (
	"testing"

	"github.com/miekg/dns"
)

func TestResolverTypeIsValid(t *testing.T) {
	tests := []struct {
		input ResolverType
		want  bool
	}{
		{ResolverTypeUDP, true},
		{ResolverTypeTCP, true},
		{ResolverTypeDOT, true},
		{"UDP", true},
		{" TCP ", true},
		{"dot", true},
		{"doh", false},
		{"quic", false},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(string(tc.input), func(t *testing.T) {
			if got := tc.input.IsValid(); got != tc.want {
				t.Errorf("ResolverType(%q).IsValid() = %v; want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseResolverType(t *testing.T) {
	tests := []struct {
		input string
		want  ResolverType
	}{
		{"udp", ResolverTypeUDP},
		{"UDP", ResolverTypeUDP},
		{"Udp", ResolverTypeUDP},
		{" udp ", ResolverTypeUDP},
		{"tcp", ResolverTypeTCP},
		{"TCP", ResolverTypeTCP},
		{" tcp ", ResolverTypeTCP},
		{"dot", ResolverTypeDOT},
		{"DOT", ResolverTypeDOT},
		{" Dot ", ResolverTypeDOT},
		{"doh", ResolverTypeUDP},
		{"quic", ResolverTypeUDP},
		{"", ResolverTypeUDP},
		{"unknown", ResolverTypeUDP},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if got := ParseResolverType(tc.input); got != tc.want {
				t.Errorf("ParseResolverType(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestRecordTypeIsValid(t *testing.T) {
	tests := []struct {
		input RecordType
		want  bool
	}{
		{TypeA, true},
		{TypeAAAA, true},
		{TypeCNAME, true},
		{TypeNS, true},
		{TypeMX, true},
		{TypeTXT, true},
		{TypeSRV, true},
		{TypeNULL, true},
		{"a", true},
		{" a ", true},
		{"aaaa", true},
		{"txt", true},
		{"", false},
		{"SOA", false},
		{"HTTPS", false},
		{"unknown", false},
	}

	for _, tc := range tests {
		t.Run(string(tc.input), func(t *testing.T) {
			if got := tc.input.IsValid(); got != tc.want {
				t.Errorf("RecordType(%q).IsValid() = %v; want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseRecordType(t *testing.T) {
	tests := []struct {
		input string
		want  RecordType
	}{
		{"A", TypeA},
		{"a", TypeA},
		{" A ", TypeA},
		{"AAAA", TypeAAAA},
		{"aaaa", TypeAAAA},
		{"CNAME", TypeCNAME},
		{"cname", TypeCNAME},
		{"NS", TypeNS},
		{"ns", TypeNS},
		{"MX", TypeMX},
		{"mx", TypeMX},
		{"TXT", TypeTXT},
		{"txt", TypeTXT},
		{"SRV", TypeSRV},
		{"srv", TypeSRV},
		{"NULL", TypeNULL},
		{"null", TypeNULL},
		{"", ""},
		{"SOA", ""},
		{"HTTPS", ""},
		{"unknown", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if got := ParseRecordType(tc.input); got != tc.want {
				t.Errorf("ParseRecordType(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestToMiekgDNS(t *testing.T) {
	tests := []struct {
		input RecordType
		want  uint16
	}{
		{TypeA, dns.TypeA},
		{TypeAAAA, dns.TypeAAAA},
		{TypeCNAME, dns.TypeCNAME},
		{TypeNS, dns.TypeNS},
		{TypeMX, dns.TypeMX},
		{TypeTXT, dns.TypeTXT},
		{TypeSRV, dns.TypeSRV},
		{TypeNULL, dns.TypeNULL},
		{"a", dns.TypeA},
		{" a ", dns.TypeA},
		{"txt", dns.TypeTXT},
		{"unknown", dns.TypeNone},
		{"", dns.TypeNone},
	}

	for _, tc := range tests {
		t.Run(string(tc.input), func(t *testing.T) {
			if got := toMiekgDNS(tc.input); got != tc.want {
				t.Errorf("toMiekgDNS(%q) = %d; want %d", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseDNSRcode(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"noerror", dns.RcodeSuccess},
		{"NOERROR", dns.RcodeSuccess},
		{"success", dns.RcodeSuccess},
		{" SUCCESS ", dns.RcodeSuccess},

		{"formerr", dns.RcodeFormatError},
		{"FORMERR", dns.RcodeFormatError},
		{"formaterror", dns.RcodeFormatError},
		{"FormatError", dns.RcodeFormatError},

		{"servfail", dns.RcodeServerFailure},
		{"SERVFAIL", dns.RcodeServerFailure},
		{"serverfailure", dns.RcodeServerFailure},
		{"ServerFailure", dns.RcodeServerFailure},

		{"nxdomain", dns.RcodeNameError},
		{"NXDOMAIN", dns.RcodeNameError},
		{"nameerror", dns.RcodeNameError},
		{"NameError", dns.RcodeNameError},

		{"notimp", dns.RcodeNotImplemented},
		{"NOTIMP", dns.RcodeNotImplemented},
		{"notimplemented", dns.RcodeNotImplemented},
		{"NotImplemented", dns.RcodeNotImplemented},

		{"refused", dns.RcodeRefused},
		{"REFUSED", dns.RcodeRefused},
		{" Refused ", dns.RcodeRefused},

		{"", dns.RcodeServerFailure},
		{"unknown", dns.RcodeServerFailure},
		{"1234", dns.RcodeServerFailure},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if got := ParseDNSRcode(tc.input); got != tc.want {
				t.Errorf("ParseDNSRcode(%q) = %d; want %d", tc.input, got, tc.want)
			}
		})
	}
}

func TestAuthMethodIsValid(t *testing.T) {
	tests := []struct {
		input AuthMethod
		want  bool
	}{
		{AuthNone, true},
		{AuthPassword, true},
		{AuthKey, true},
		{"NONE", true},
		{" password ", true},
		{"KEY", true},
		{"token", false},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(string(tc.input), func(t *testing.T) {
			if got := tc.input.IsValid(); got != tc.want {
				t.Errorf("AuthMethod(%q).IsValid() = %v; want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseAuthMethod(t *testing.T) {
	tests := []struct {
		input string
		want  AuthMethod
	}{
		{"none", AuthNone},
		{"NONE", AuthNone},
		{" none ", AuthNone},
		{"password", AuthPassword},
		{"PASSWORD", AuthPassword},
		{" Password ", AuthPassword},
		{"key", AuthKey},
		{"KEY", AuthKey},
		{" Key ", AuthKey},
		{"", AuthNone},
		{"unknown", AuthNone},
		{"token", AuthNone},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if got := ParseAuthMethod(tc.input); got != tc.want {
				t.Errorf("ParseAuthMethod(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}
