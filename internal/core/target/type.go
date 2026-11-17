package target

import (
	"net"
	"net/netip"
	"strconv"
)

// Kind is what a list entry points at.
type Kind uint8

const (
	KindIP Kind = iota
	KindDomain
	KindHost
)

// String returns a short label for the entry kind.
func (k Kind) String() string {
	switch k {
	case KindIP:
		return "IP"
	case KindDomain:
		return "Domain"
	case KindHost:
		return "Host"
	default:
		return "Unknown"
	}
}

func (k Kind) Dir() string {
	switch k {
	case KindIP:
		return "/ips"
	case KindDomain:
		return "/domains"
	default:
		return ""
	}
}

// Entry is one parsed row of a target list.
//
// Canonical CSV form: value,port,path,enable
// Port 0 means "use the scan settings"; path is empty or starts with "/".
type Entry struct {
	Kind   Kind
	IP     netip.Prefix
	Domain string
	Path   string
	Port   uint16
	Enable bool
}

// New builds an IP entry. Only enable == 1 is enabled.
func New(ip netip.Prefix, enable int) Entry {
	return Entry{Kind: KindIP, IP: ip, Enable: enable == 1}
}

func NewEnabled(ip netip.Prefix) Entry  { return New(ip, 1) }
func NewDisabled(ip netip.Prefix) Entry { return New(ip, 0) }

// NewDomain builds a domain entry. Only enable == 1 is enabled.
func NewDomain(domain string, enable int) Entry {
	return Entry{Kind: KindDomain, Domain: domain, Enable: enable == 1}
}

// Value is the host part: prefix string for IPs, name for domains.
//
// The path is a separate column, so it must not leak in here: Value is also
// the first field EncodeCSV writes, and a value like "example.com/x" no
// longer parses back as a domain.
func (e Entry) Value() string {
	switch e.Kind {
	case KindDomain:
		return e.Domain
	case KindIP:
		return e.IP.String()
	default:
		return ""
	}
}

// prefix length. Value is for CSV; Host is for dialing and URLs.
// Host returns the dialable host, including an optional port and path.
func (e Entry) Host() string {
	var host string

	switch e.Kind {
	case KindDomain:
		host = e.Domain
	case KindIP:
		host = e.IP.Addr().String()
	default:
		return ""
	}

	if e.Port != 0 {
		host = net.JoinHostPort(host, strconv.Itoa(int(e.Port)))
	}

	return host + e.Path
}

// Key is the dedup identity: same host, port and path means same target.
func (e Entry) Key() string {
	return e.Value() + ":" +
		strconv.FormatUint(uint64(e.Port), 10) + ":" + e.Path
}

// IsSingle reports whether the entry is exactly one address (/32 or /128).
func (e Entry) IsSingle() bool {
	return e.Kind == KindIP && e.IP.Bits() == e.IP.Addr().BitLen()
}

// EncodeCSV renders the row as value,port,path,enable.
func (e Entry) EncodeCSV() []string {
	enable := "0"
	if e.Enable {
		enable = "1"
	}
	return []string{
		e.Value(),
		strconv.FormatUint(uint64(e.Port), 10),
		e.Path,
		enable,
	}
}

// CIDRBlock is one contiguous run of addresses in the shuffle index.
//
// Entry is the row the block came from. Port, path, and enable are the same
// for every address the block expands into, so the block keeps one copy
// instead of the index storing a full Entry per address.
type CIDRBlock struct {
	StartIP   netip.Addr
	TotalIPs  uint64
	GlobalIdx uint64 // index of StartIP in the global address space
	Entry     Entry
}

// MasterIndexer is the shuffle index for a list. CIDR ranges stay ranges;
// single IPs are kept as byte offsets into the file and read on demand.
type MasterIndexer struct {
	FilePath      string
	CIDRBlocks    []CIDRBlock
	SingleOffsets []int64
	TotalCIDRIPs  uint64
	TotalSingles  uint64
	GrandTotal    uint64
}
