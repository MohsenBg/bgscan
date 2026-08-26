// Package dns provides types for DNS-based scanning and tunneling.
package dns

import (
	"strings"
	"time"

	"bgscan/internal/core/fileutil"

	"github.com/miekg/dns"
)

type DNSTunProtocol string

const (
	DNSTunProtocolVayDNS     DNSTunProtocol = "vaydns"
	DNSTunProtocolDNSTT      DNSTunProtocol = "dnstt"
	DNSTunProtocolSlipstream DNSTunProtocol = "slipstream"
)

type DNSTunConfigFile struct {
	Name      string
	Path      string
	CreatedAt time.Time
	Protocol  DNSTunProtocol
	Proxy     string
	Config    any
}

type ConfigValidationResult struct {
	File   fileutil.FileEntry
	Errors map[string]error
}

// ResolverType identifies the transport used to communicate with a DNS resolver.
type ResolverType string

const (
	ResolverTypeUDP ResolverType = "udp"
	ResolverTypeTCP ResolverType = "tcp"
	ResolverTypeDOT ResolverType = "dot"
)

// IsValid reports whether the resolver type is supported.
func (t ResolverType) IsValid() bool {
	switch strings.ToLower(strings.TrimSpace(string(t))) {
	case string(ResolverTypeUDP), string(ResolverTypeTCP), string(ResolverTypeDOT):
		return true
	default:
		return false
	}
}

// ParseResolverType parses a resolver type.
// Parsing is case-insensitive and ignores surrounding whitespace.
// Unknown values default to ResolverTypeUDP.
func ParseResolverType(s string) ResolverType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case string(ResolverTypeTCP):
		return ResolverTypeTCP
	case string(ResolverTypeDOT):
		return ResolverTypeDOT
	case string(ResolverTypeUDP):
		return ResolverTypeUDP
	default:
		return ResolverTypeUDP
	}
}

// RecordType identifies a DNS record type used in queries.
type RecordType string

const (
	TypeA     RecordType = "A"
	TypeAAAA  RecordType = "AAAA"
	TypeCNAME RecordType = "CNAME"
	TypeNS    RecordType = "NS"
	TypeMX    RecordType = "MX"
	TypeTXT   RecordType = "TXT"
	TypeSRV   RecordType = "SRV"
	TypeNULL  RecordType = "NULL"
	TypeCAA   RecordType = "CAA"
)

// IsValid reports whether the record type is supported.
func (r RecordType) IsValid() bool {
	switch strings.ToUpper(strings.TrimSpace(string(r))) {
	case string(TypeA),
		string(TypeAAAA),
		string(TypeCNAME),
		string(TypeNS),
		string(TypeMX),
		string(TypeTXT),
		string(TypeSRV),
		string(TypeNULL),
		string(TypeCAA):
		return true
	default:
		return false
	}
}

// ParseRecordType parses a DNS record type.
// Parsing is case-insensitive and ignores surrounding whitespace.
// Unknown values return an empty RecordType.
func ParseRecordType(s string) RecordType {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case string(TypeA):
		return TypeA
	case string(TypeAAAA):
		return TypeAAAA
	case string(TypeCNAME):
		return TypeCNAME
	case string(TypeNS):
		return TypeNS
	case string(TypeMX):
		return TypeMX
	case string(TypeTXT):
		return TypeTXT
	case string(TypeSRV):
		return TypeSRV
	case string(TypeNULL):
		return TypeNULL
	case string(TypeCAA):
		return TypeCAA
	default:
		return ""
	}
}

func toMiekgDNS(record RecordType) uint16 {
	switch strings.ToUpper(strings.TrimSpace(string(record))) {
	case string(TypeA):
		return dns.TypeA
	case string(TypeAAAA):
		return dns.TypeAAAA
	case string(TypeCNAME):
		return dns.TypeCNAME
	case string(TypeNS):
		return dns.TypeNS
	case string(TypeMX):
		return dns.TypeMX
	case string(TypeTXT):
		return dns.TypeTXT
	case string(TypeSRV):
		return dns.TypeSRV
	case string(TypeNULL):
		return dns.TypeNULL
	case string(TypeCAA):
		return dns.TypeCAA
	default:
		return dns.TypeNone
	}
}

// ParseDNSRcode parses a textual DNS response code.
// Unknown values return dns.RcodeServerFailure.
func ParseDNSRcode(rCode string) int {
	switch strings.ToLower(strings.TrimSpace(rCode)) {
	case "noerror", "success":
		return dns.RcodeSuccess
	case "formerr", "formaterror":
		return dns.RcodeFormatError
	case "servfail", "serverfailure":
		return dns.RcodeServerFailure
	case "nxdomain", "nameerror":
		return dns.RcodeNameError
	case "notimp", "notimplemented":
		return dns.RcodeNotImplemented
	case "refused":
		return dns.RcodeRefused
	default:
		return dns.RcodeServerFailure
	}
}

// AuthMethod identifies how a connection authenticates.
type AuthMethod string

const (
	AuthNone     AuthMethod = "none"
	AuthPassword AuthMethod = "password"
	AuthKey      AuthMethod = "key"
)

// IsValid reports whether the authentication method is supported.
func (m AuthMethod) IsValid() bool {
	switch strings.ToLower(strings.TrimSpace(string(m))) {
	case string(AuthNone), string(AuthPassword), string(AuthKey):
		return true
	default:
		return false
	}
}

// ParseAuthMethod parses an authentication method.
// Parsing is case-insensitive and ignores surrounding whitespace.
// Unknown values return AuthNone.
func ParseAuthMethod(s string) AuthMethod {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case string(AuthPassword):
		return AuthPassword
	case string(AuthKey):
		return AuthKey
	case string(AuthNone):
		return AuthNone
	default:
		return AuthNone
	}
}

type ResolverProxyType string

const (
	ResolverProxySOCKS ResolverProxyType = "socks"
	ResolverProxySSH   ResolverProxyType = "ssh"
)

func (t ResolverProxyType) IsValid() bool {
	switch strings.ToLower(strings.TrimSpace(string(t))) {
	case string(ResolverProxySOCKS), string(ResolverProxySSH):
		return true
	default:
		return false
	}
}

func ParseResolverProxyType(s string) ResolverProxyType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case string(ResolverProxySOCKS):
		return ResolverProxySOCKS
	case string(ResolverProxySSH):
		return ResolverProxySSH
	default:
		return ""
	}
}
