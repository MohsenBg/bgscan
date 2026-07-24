package dns

import (
	"testing"

	"github.com/miekg/dns"
)

// ─────────────────────────────────────────────────────────────────────────────
// type.go — ParseTransport
// ─────────────────────────────────────────────────────────────────────────────

func TestParseTransport_KnownValues(t *testing.T) {
	tests := []struct {
		input string
		want  Transport
	}{
		{"UDP", UDP},
		{"TCP", TCP},
		{"DOT", DOT},
		{"DOH", DOT},
		{"udp", UDP},
		{"tcp", TCP},
		{"dot", DOT},
		{"doh", DOT},
		{"Udp", UDP},
		{"Tcp", TCP},
		{"Dot", DOT},
		{" UDP ", UDP},
		{"\tTCP\t", TCP},
		{"  DOT  ", DOT},
		{"", UDP},
		{"QUIC", UDP},
		{"dns", UDP},
		{"123", UDP},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			got := ParseTransport(tc.input)
			if got != tc.want {
				t.Errorf("ParseTransport(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseTransport_DOH_NeverReturned(t *testing.T) {
	if DOH != "DOH" {
		t.Errorf("DOH constant value changed: got %q, want %q", DOH, "DOH")
	}
	got := ParseTransport("DOH")
	if got == DOH {
		t.Errorf("ParseTransport(\"DOH\") = DOH: the DOH transport is now returned, update this test if DOH support was intentionally added. got %q", got)
	}
	if got != DOT {
		t.Errorf("ParseTransport(\"DOH\") = %q; want DOT (current documented fallback)", got)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// type.go — ParseDNSRcode
// ─────────────────────────────────────────────────────────────────────────────

func TestParseDNSRcode_KnownValues(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"noerror", dns.RcodeSuccess},
		{"NOERROR", dns.RcodeSuccess},
		{"success", dns.RcodeSuccess},
		{"SUCCESS", dns.RcodeSuccess},
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
		{"Refused", dns.RcodeRefused},
		{"", dns.RcodeServerFailure},
		{"badcode", dns.RcodeServerFailure},
		{"1234", dns.RcodeServerFailure},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			got := ParseDNSRcode(tc.input)
			if got != tc.want {
				t.Errorf("ParseDNSRcode(%q) = %d; want %d", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseDNSRcode_Trimming(t *testing.T) {
	got := ParseDNSRcode("  noerror  ")
	if got != dns.RcodeSuccess {
		t.Errorf("ParseDNSRcode with leading/trailing spaces: got %d, want %d (RcodeSuccess)", got, dns.RcodeSuccess)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// type.go — Transport / RecordType constants
// ─────────────────────────────────────────────────────────────────────────────

func TestTransport_ConstantValues(t *testing.T) {
	tests := []struct{ name, want string }{
		{"UDP", "UDP"}, {"TCP", "TCP"}, {"DOT", "DOT"}, {"DOH", "DOH"},
	}
	values := map[string]Transport{
		"UDP": UDP, "TCP": TCP, "DOT": DOT, "DOH": DOH,
	}
	for _, tc := range tests {
		if string(values[tc.name]) != tc.want {
			t.Errorf("Transport %s = %q; want %q", tc.name, values[tc.name], tc.want)
		}
	}
}

func TestRecordType_ConstantValues(t *testing.T) {
	tests := []struct {
		name string
		rt   RecordType
		want string
	}{
		{"TypeA", TypeA, "A"},
		{"TypeAAAA", TypeAAAA, "AAAA"},
		{"TypeCNAME", TypeCNAME, "CNAME"},
		{"TypeNS", TypeNS, "NS"},
		{"TypeMX", TypeMX, "MX"},
		{"TypeTXT", TypeTXT, "TXT"},
	}
	for _, tc := range tests {
		if string(tc.rt) != tc.want {
			t.Errorf("RecordType %s = %q; want %q", tc.name, tc.rt, tc.want)
		}
	}
}
