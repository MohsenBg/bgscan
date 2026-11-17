package target

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math/big"
	"math/bits"
	"math/rand"
	"net/netip"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/MohsenBg/bgscan/internal/logger"
)

// Each single IP or domain costs 8 bytes in memory (one file offset),
// so even 10M lines stay near 100MB.
const shuffleMemWarnThreshold = 10_000_000

// avgLineBytes is a rough guess, used only for the warning and preallocation.
const avgLineBytes = 20

// maxIndexPrealloc caps the up-front allocation so a huge file doesn't blow
// up memory before we've read a single line.
const maxIndexPrealloc = 1 << 20

// cidrRow is one CIDR range found while reading the file.
// There is usually only a handful of these, so keeping an Entry is cheap.
type cidrRow struct {
	start netip.Addr
	size  *big.Int
	entry Entry
}

// NewMasterIndexer builds the index streamActiveShuffled walks.
//
// Single IPs and domains are "points": one target each. We store only their
// byte offset in the file (8 bytes each) and read the row back later.
// CIDR ranges stay as ranges.
//
// If the total count fits in uint64, every address gets its own index.
// Otherwise each range gets a proportional slice of the 2^64 space, starting
// at a random offset, so huge IPv6 ranges still get spread out.
func NewMasterIndexer(log *logger.Logger, filePath string) (*MasterIndexer, error) {
	indexer := &MasterIndexer{FilePath: filePath}

	fi, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if fi.Size() == 0 {
		return indexer, nil
	}

	estimatedLines := fi.Size() / avgLineBytes
	if estimatedLines > shuffleMemWarnThreshold && log != nil {
		log.Warn(
			"shuffled mode: file %q has ~%dM estimated lines; indexing needs ~8 bytes per line — consider sequential mode for very large lists",
			filePath, estimatedLines/1_000_000,
		)
	}

	prealloc := min(estimatedLines, maxIndexPrealloc)

	// points is the FINAL slice (no copy later): offsets of single IPs and
	// domains. cidrs holds the ranges.
	points := make([]int64, 0, int(prealloc))
	var cidrs []cidrRow

	err = ReadCSV(filePath, func(row Entry, offset int64) error {
		if !row.Enable {
			return nil
		}

		// A domain or a single IP is exactly one target:
		// remember where its line is, read it when its turn comes.
		if row.Kind == KindDomain || row.IsSingle() {
			points = append(points, offset)
			return nil
		}

		prefix := row.IP.Masked()
		addr := prefix.Addr()
		cidrs = append(cidrs, cidrRow{
			start: addr,
			size:  new(big.Int).Lsh(big.NewInt(1), uint(addr.BitLen()-prefix.Bits())), //
			entry: row,
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Total number of targets = points + every address in every range.
	totalCount := new(big.Int).SetUint64(uint64(len(points)))
	for _, c := range cidrs {
		totalCount.Add(totalCount, c.size)
	}
	if totalCount.Sign() == 0 {
		return indexer, nil
	}

	// SingleOffsets / TotalSingles now mean "single IPs and domains".
	indexer.SingleOffsets = points
	indexer.TotalSingles = uint64(len(points))

	maxUint64 := new(big.Int).SetUint64(^uint64(0))
	if totalCount.Cmp(maxUint64) <= 0 {
		indexer.buildDirect(cidrs)
	} else {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		indexer.buildSampled(cidrs, totalCount, maxUint64, rng)
	}
	return indexer, nil
}

// buildDirect gives every address its own index. Only used when the total
// fits in uint64.
//
// Global index layout:  [ range addresses ... | points ... ]
func (mi *MasterIndexer) buildDirect(cidrs []cidrRow) {
	mi.CIDRBlocks = make([]CIDRBlock, 0, len(cidrs))
	for _, c := range cidrs {
		count := c.size.Uint64()
		mi.CIDRBlocks = append(mi.CIDRBlocks, CIDRBlock{
			StartIP:   c.start,
			TotalIPs:  count,
			GlobalIdx: mi.TotalCIDRIPs,
			Entry:     c.entry,
		})
		mi.TotalCIDRIPs = saturatingAdd(mi.TotalCIDRIPs, count)
	}
	mi.GrandTotal = saturatingAdd(mi.TotalCIDRIPs, mi.TotalSingles)
}

// buildSampled hands each range a quota of the 2^64 space proportional to its
// size, then starts it at a random offset inside the range. Ranges too small
// to earn a quota are dropped; single IPs and domains are always kept.
func (mi *MasterIndexer) buildSampled(cidrs []cidrRow, totalCount, maxUint64 *big.Int, rng *rand.Rand) {
	for _, c := range cidrs {
		quotaBig := new(big.Int).Mul(c.size, maxUint64)
		quotaBig.Div(quotaBig, totalCount)
		if quotaBig.Sign() == 0 {
			continue // range too small to matter at this scale
		}
		quota := quotaBig.Uint64() // always fits: quota <= maxUint64

		offset := new(big.Int)
		if maxOffset := new(big.Int).Sub(c.size, quotaBig); maxOffset.Sign() > 0 {
			offset.Rand(rng, maxOffset)
		}
		mi.CIDRBlocks = append(mi.CIDRBlocks, CIDRBlock{
			StartIP:   addBigOffset(c.start, offset),
			TotalIPs:  quota,
			GlobalIdx: mi.TotalCIDRIPs,
			Entry:     c.entry,
		})
		mi.TotalCIDRIPs = saturatingAdd(mi.TotalCIDRIPs, quota)
	}
	mi.GrandTotal = saturatingAdd(mi.TotalCIDRIPs, mi.TotalSingles)
}

// permuter is a keyed bijection on [0, 2^bits): every input maps to exactly
// one output, and no two inputs share an output. Each round is invertible:
//
//	x ^= key        (xor again to undo)
//	x ^= x >> s     (high bits stay the same, so it can be undone)
//	x *= odd        (an odd number has an inverse mod 2^bits)
//
// Not cryptographic, just well mixed. Memory is O(1).
type permuter struct {
	mask uint64
	s    uint
	mul  [3]uint64 // odd multipliers
	xor  [3]uint64 // xor keys
}

func newPermuter(nbits uint, rng *rand.Rand) permuter {
	p := permuter{mask: ^uint64(0), s: (nbits + 1) / 2}
	if nbits < 64 {
		p.mask = uint64(1)<<nbits - 1
	}
	if p.s == 0 {
		p.s = 1
	}
	for i := range p.mul {
		p.mul[i] = rng.Uint64() | 1 // odd => invertible
		p.xor[i] = rng.Uint64() & p.mask
	}
	return p
}

// at returns the shuffled value for counter i (i must be < 2^nbits).
func (p *permuter) at(i uint64) uint64 {
	x := i
	for r := range p.mul {
		x ^= p.xor[r]
		x ^= x >> p.s
		x = (x * p.mul[r]) & p.mask
	}
	return x
}

// streamActiveShuffled walks the index in pseudo-random order, so memory
// stays flat no matter how many addresses the list holds.
func streamActiveShuffled(log *logger.Logger, ctx context.Context, path string, limit uint64, out chan<- Entry) error {
	indexer, err := NewMasterIndexer(log, path)
	if err != nil {
		return fmt.Errorf("shuffled pre-scan initialization failed: %w", err)
	}
	if indexer.GrandTotal == 0 {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Error("failed to close file: %v", err)
		}
	}()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Round the range up to a power of two so rejection stays under 50%.
	// Past 2^63 there is no bigger power of two, so use all 64 bits.
	nbits := uint(64)
	if indexer.GrandTotal <= 1<<63 {
		nbits = uint(bits.Len64(indexer.GrandTotal))
		if indexer.GrandTotal&(indexer.GrandTotal-1) == 0 {
			nbits-- // already a power of two
		}
	}
	perm := newPermuter(nbits, rng)

	// We only ever read one line at a time, so a small buffer is enough.
	reader := bufio.NewReaderSize(file, 512)

	var dispatched, count, ctr uint64
	for dispatched < indexer.GrandTotal {
		if limit > 0 && count >= limit {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		state := perm.at(ctr)
		ctr++
		if state >= indexer.GrandTotal {
			continue // landed in the padding above GrandTotal
		}

		var entry Entry
		if state < indexer.TotalCIDRIPs {
			// An address inside a range.
			block, addr := indexer.cidrBlockAt(state)
			entry = block.Entry
			entry.IP = netip.PrefixFrom(addr, addr.BitLen())
		} else {
			// A single IP or a domain: read its row by offset.
			var err error
			entry, err = readEntryAtCSVOffset(file, reader, indexer.SingleOffsets[state-indexer.TotalCIDRIPs])
			if err != nil {
				// Skip unreadable lines rather than aborting the whole scan.
				dispatched++
				continue
			}
		}

		select {
		case out <- entry:
			count++
		case <-ctx.Done():
			return ctx.Err()
		}
		dispatched++
	}
	return nil
}

// cidrBlockAt finds the block owning globalIdx and returns it together with
// the address at that offset inside it. The block carries the source row, so
// the caller can copy its port and path onto the address.
func (mi *MasterIndexer) cidrBlockAt(globalIdx uint64) (CIDRBlock, netip.Addr) {
	// Last block whose GlobalIdx <= globalIdx.
	i := sort.Search(len(mi.CIDRBlocks), func(i int) bool {
		return mi.CIDRBlocks[i].GlobalIdx > globalIdx
	}) - 1
	block := mi.CIDRBlocks[i]
	return block, addOffsetToAddr(block.StartIP, globalIdx-block.GlobalIdx)
}

// addOffsetToAddr adds offset to addr byte by byte, carrying across bytes.
// Works for IPv4 and IPv6 alike.
func addOffsetToAddr(addr netip.Addr, offset uint64) netip.Addr {
	b := addr.As16()
	carry := offset
	for i := 15; i >= 0 && carry > 0; i-- {
		sum := uint64(b[i]) + (carry & 0xff)
		b[i] = byte(sum)
		carry = (carry >> 8) + (sum >> 8)
	}
	return netip.AddrFrom16(b).Unmap()
}

// addBigOffset is addOffsetToAddr for offsets that don't fit in uint64
// (IPv6 /63 and wider). The sum always fits in 128 bits: the offset is
// smaller than the range it points into.
func addBigOffset(addr netip.Addr, offset *big.Int) netip.Addr {
	b := addr.As16()
	sum := new(big.Int).SetBytes(b[:])
	sum.Add(sum, offset)
	var out [16]byte
	sum.FillBytes(out[:])
	return netip.AddrFrom16(out).Unmap()
}

// readEntryAtCSVOffset seeks to a byte offset and parses the row there.
// The row can be a single IP or a domain.
//
// The index was built from this same file, so the row is expected to still
// be a canonical record.
func readEntryAtCSVOffset(file *os.File, reader *bufio.Reader, offset int64) (Entry, error) {
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return Entry{}, err
	}
	reader.Reset(file)
	lineBytes, err := reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return Entry{}, err
	}
	line := strings.TrimSpace(string(lineBytes))
	if line == "" {
		return Entry{}, fmt.Errorf("empty target row at offset %d", offset)
	}

	// Fast path: no quotes means plain fields. Splitting is much cheaper than
	// creating a csv.Reader for every entry. If the field count is not what
	// we expect, fall through to the full CSV parser.
	if !strings.ContainsRune(line, '"') {
		rec := strings.Split(line, string(DefaultCSVConfig.Comma))
		if len(rec) == 4 {
			if DefaultCSVConfig.TrimLeadingSpace {
				for i := range rec {
					rec[i] = strings.TrimLeft(rec[i], " \t")
				}
			}
			entry, err := ParseRecord(rec)
			if err != nil {
				return Entry{}, fmt.Errorf("target row at offset %d: %w", offset, err)
			}
			return entry, nil
		}
	}

	// Slow path: a path may contain a comma or quotes, so parse the line the
	// way ReadCSV does.
	cr := csv.NewReader(strings.NewReader(line))
	cr.Comma = DefaultCSVConfig.Comma
	cr.FieldsPerRecord = DefaultCSVConfig.FieldsPerRecord
	cr.LazyQuotes = DefaultCSVConfig.LazyQuotes
	cr.TrimLeadingSpace = DefaultCSVConfig.TrimLeadingSpace

	rec, err := cr.Read()
	if err != nil {
		return Entry{}, fmt.Errorf("read target row at offset %d: %w", offset, err)
	}

	entry, err := ParseRecord(rec)
	if err != nil {
		return Entry{}, fmt.Errorf("target row at offset %d: %w", offset, err)
	}
	return entry, nil
}

// saturatingAdd adds a+b, clamping at MaxUint64 instead of wrapping.
func saturatingAdd(a, b uint64) uint64 {
	sum := a + b
	if sum < a {
		return ^uint64(0)
	}
	return sum
}
