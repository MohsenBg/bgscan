package iplist

import (
	"net/netip"
)

// IPList represents a single normalized parsed entry from your data source.
type IPList struct {
	IP     netip.Prefix
	Enable bool
}

// New instantiates a normalized IPList entry.
func New(ip netip.Prefix, enable int) IPList {
	return IPList{
		IP:     ip,
		Enable: enable == 1,
	}
}

// NewEnabled creates an enabled IPList entry.
func NewEnabled(ip netip.Prefix) IPList {
	return IPList{
		IP:     ip,
		Enable: true,
	}
}

// NewDisabled creates a disabled IPList entry.
func NewDisabled(ip netip.Prefix) IPList {
	return IPList{
		IP:     ip,
		Enable: false,
	}
}

// IsSingle returns true if the entry is a single host address (e.g. /32 for IPv4 or /128 for IPv6).
func (e IPList) IsSingle() bool {
	return e.IP.Bits() == e.IP.Addr().BitLen()
}

// EncodeCSV returns the CSV representation of the entry.
//
// Format:
//
//	ip,enable
//
// Example:
//
//	192.168.1.1,1
//	10.0.0.0/24,0
func (e IPList) EncodeCSV() []string {
	enable := "0"
	if e.Enable {
		enable = "1"
	}

	return []string{e.IP.String(), enable}
}

// CIDRBlock stores the mathematical boundaries for subnets.
type CIDRBlock struct {
	StartIP   netip.Addr
	TotalIPs  uint64
	GlobalIdx uint64
}

// MasterIndexer tracks the hybrid data schema across files.
type MasterIndexer struct {
	FilePath      string
	CIDRBlocks    []CIDRBlock
	SingleOffsets []int64
	TotalCIDRIPs  uint64
	TotalSingles  uint64
	GrandTotal    uint64
}
