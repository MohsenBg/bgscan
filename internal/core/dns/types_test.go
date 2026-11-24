package dns

import (
	"testing"

	"github.com/miekg/dns"
)

// TestResolverTypeIsValid checks that ResolverType.IsValid accepts the known transports in any case.
func TestResolverTypeIsValid(t *testing.T) {
	tests := []struct {
		input ResolverType
		want  bool
	}{
		{ResolverTypeUDP, true},
		{ResolverTypeTCP, true},
		{ResolverTypeDOT, true},
		{ResolverTypeDOH, true},
		{"UDP", true},
		{" TCP ", true},
		{"dot", true},
		{"doh", true},
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

// TestParseResolverType checks that ParseResolverType parses known transports and falls back for unknown values.
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
		{"doh", ResolverTypeDOH},
		{"DOH", ResolverTypeDOH},
		{" doh ", ResolverTypeDOH},
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

// TestRecordTypeIsValid checks that RecordType.IsValid accepts the known record types.
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

// TestParseRecordType checks that ParseRecordType parses known record types and falls back for unknown values.
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

// TestToMiekgDNS checks that toMiekgDNS maps each record type to the miekg/dns query type.
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

// TestParseDNSRcode checks that ParseDNSRcode parses response code names.
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

// TestAuthMethodIsValid checks that AuthMethod.IsValid accepts the known methods.
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

// TestParseAuthMethod checks that ParseAuthMethod parses known methods and returns AuthNone for unknown values.
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

// TestResolverProxyTypeIsValid checks that ResolverProxyType.IsValid accepts socks and ssh in any case.
func TestResolverProxyTypeIsValid(t *testing.T) {
	tests := []struct {
		input ResolverProxyType
		want  bool
	}{
		{ResolverProxySOCKS, true},
		{ResolverProxySSH, true},
		{"SOCKS", true},
		{" ssh ", true},
		{"http", false},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(string(tc.input), func(t *testing.T) {
			if got := tc.input.IsValid(); got != tc.want {
				t.Errorf("ResolverProxyType(%q).IsValid() = %v; want %v", tc.input, got, tc.want)
			}
		})
	}
}

// TestParseResolverProxyType checks that ParseResolverProxyType parses known types and returns an empty value for unknown ones.
func TestParseResolverProxyType(t *testing.T) {
	tests := []struct {
		input string
		want  ResolverProxyType
	}{
		{"socks", ResolverProxySOCKS},
		{"SOCKS", ResolverProxySOCKS},
		{" Socks ", ResolverProxySOCKS},
		{"ssh", ResolverProxySSH},
		{"SSH", ResolverProxySSH},
		{"ssh ", ResolverProxySSH},
		{"http", ""},
		{"", ""},
		{"unknown", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if got := ParseResolverProxyType(tc.input); got != tc.want {
				t.Errorf("ParseResolverProxyType(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

// TestEncMethodString checks that enc method string.
func TestEncMethodString(t *testing.T) {
	tests := []struct {
		method EncMethod
		want   string
	}{
		{EncNone, "None"},
		{EncXOR, "XOR"},
		{EncChaCha20, "ChaCha20"},
		{EncAES128GCM, "AES-128-GCM"},
		{EncAES192GCM, "AES-192-GCM"},
		{EncAES256GCM, "AES-256-GCM"},
		{EncMethod(6), "Unknown"},
		{EncMethod(-1), "Unknown"},
	}

	for _, tc := range tests {
		if got := tc.method.String(); got != tc.want {
			t.Errorf("EncMethod(%d).String() = %q, want %q", int(tc.method), got, tc.want)
		}
	}
}

// TestParseEncMethod checks that parse enc method.
func TestParseEncMethod(t *testing.T) {
	tests := []struct {
		input string
		want  EncMethod
	}{
		{"XOR", EncXOR},
		{"xor", EncXOR},
		{" None ", EncNone},
		{"ChaCha20", EncChaCha20},
		{"AES-128-GCM", EncAES128GCM},
		{"aes-256-gcm", EncAES256GCM},
		{"unknown", EncXOR},
		{"", EncXOR},
	}

	for _, tc := range tests {
		if got := ParseEncMethod(tc.input); got != tc.want {
			t.Errorf("ParseEncMethod(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}
