package target

import (
	"net/netip"
	"strconv"
)

// Kind is what a list entry points at.
type Kind uint8

const (
	KindIP Kind = iota
	KindDomain
)

// String returns a short label for the entry kind.
func (k Kind) String() string {
	switch k {
	case KindIP:
		return "IP"
	case KindDomain:
		return "Domain"
	default:
		return "Unknown"
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

// Key is the dedup identity: same host, port and path means same target.
func (e Entry) Key() string {
	return strconv.Itoa(int(e.Kind)) + ":" + e.Value() + ":" +
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
type CIDRBlock struct {
	StartIP   netip.Addr
	TotalIPs  uint64
	GlobalIdx uint64 // index of StartIP in the global address space
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
