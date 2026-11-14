package target

import (
	"bufio"
	"context"
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

	"github.com/MohsenBg/bgscan/internal/core/netutil"
	"github.com/MohsenBg/bgscan/internal/logger"
)

// Indexing keeps roughly 65 bytes per line in memory, so lists past ~5M
// lines can peak above 300MB. Sequential mode is the better deal there.
const shuffleMemWarnThreshold = 5_000_000

// avgLineBytes is a rough guess, used only for the warning and preallocation.
const avgLineBytes = 20

// maxIndexPrealloc caps the up-front allocation so a huge file doesn't blow
// up memory before we've read a single line.
const maxIndexPrealloc = 1 << 20

// indexEntry is one enabled row, before it lands in the index.
type indexEntry struct {
	startIP netip.Addr
	offset  int64
	size    *big.Int
	isCIDR  bool
}

// NewMasterIndexer builds the index streamActiveShuffled walks.
//
// If the address count fits in uint64, every address gets its own index.
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
			"shuffled mode: file %q has ~%dM estimated lines; peak memory during indexing may exceed 300MB — consider sequential mode for very large lists",
			filePath, estimatedLines/1_000_000,
		)
	}

	prealloc := estimatedLines
	if prealloc > maxIndexPrealloc {
		prealloc = maxIndexPrealloc
	}
	entries := make([]indexEntry, 0, int(prealloc))

	totalCount := new(big.Int)
	one := big.NewInt(1)

	err = ReadCSV(filePath, func(row Entry, offset int64) error {
		if !row.Enable {
			return nil
		}
		if row.Kind != KindIP {
			return fmt.Errorf("domain target %q requires DNS resolution, which is not implemented yet", row.Domain)
		}
		if row.Port != 0 || row.Path != "" {
			return fmt.Errorf("per-target port/path is not supported in shuffled mode yet")
		}
		if row.IsSingle() {
			entries = append(entries, indexEntry{offset: offset, size: one})
			totalCount.Add(totalCount, one)
			return nil
		}
		prefix := row.IP.Masked()
		addr := prefix.Addr()
		size := new(big.Int).Lsh(one, uint(addr.BitLen()-prefix.Bits()))
		entries = append(entries, indexEntry{startIP: addr, offset: offset, size: size, isCIDR: true})
		totalCount.Add(totalCount, size)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if totalCount.Sign() == 0 {
		return indexer, nil
	}

	maxUint64 := new(big.Int).SetUint64(^uint64(0))
	if totalCount.Cmp(maxUint64) <= 0 {
		indexer.buildDirect(entries)
	} else {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		indexer.buildSampled(entries, totalCount, maxUint64, rng)
	}
	return indexer, nil
}

// buildDirect gives every address its own index. Only used when the total
// fits in uint64.
func (mi *MasterIndexer) buildDirect(entries []indexEntry) {
	mi.CIDRBlocks = make([]CIDRBlock, 0, len(entries))
	mi.SingleOffsets = make([]int64, 0, len(entries))
	for _, e := range entries {
		if !e.isCIDR {
			mi.SingleOffsets = append(mi.SingleOffsets, e.offset)
			mi.TotalSingles++
			continue
		}
		count := e.size.Uint64()
		mi.CIDRBlocks = append(mi.CIDRBlocks, CIDRBlock{
			StartIP:   e.startIP,
			TotalIPs:  count,
			GlobalIdx: mi.TotalCIDRIPs,
		})
		mi.TotalCIDRIPs = saturatingAdd(mi.TotalCIDRIPs, count)
	}
	mi.GrandTotal = saturatingAdd(mi.TotalCIDRIPs, mi.TotalSingles)
}

// buildSampled hands each range a quota of the 2^64 space proportional to its
// size, then starts it at a random offset inside the range. Ranges too small
// to earn a quota are dropped; single IPs are always kept.
func (mi *MasterIndexer) buildSampled(entries []indexEntry, totalCount, maxUint64 *big.Int, rng *rand.Rand) {
	for _, e := range entries {
		if !e.isCIDR {
			mi.SingleOffsets = append(mi.SingleOffsets, e.offset)
			mi.TotalSingles++
			continue
		}
		quotaBig := new(big.Int).Mul(e.size, maxUint64)
		quotaBig.Div(quotaBig, totalCount)
		if quotaBig.Sign() == 0 {
			continue // range too small to matter at this scale
		}
		quota := quotaBig.Uint64() // always fits: quota <= maxUint64

		offset := new(big.Int)
		if maxOffset := new(big.Int).Sub(e.size, quotaBig); maxOffset.Sign() > 0 {
			offset.Rand(rng, maxOffset)
		}
		mi.CIDRBlocks = append(mi.CIDRBlocks, CIDRBlock{
			StartIP:   addBigOffset(e.startIP, offset),
			TotalIPs:  quota,
			GlobalIdx: mi.TotalCIDRIPs,
		})
		mi.TotalCIDRIPs = saturatingAdd(mi.TotalCIDRIPs, quota)
	}
	mi.GrandTotal = saturatingAdd(mi.TotalCIDRIPs, mi.TotalSingles)
}

// streamActiveShuffled walks the index in pseudo-random order with an LCG,
// so memory stays flat no matter how many addresses the list holds.
func streamActiveShuffled(log *logger.Logger, ctx context.Context, path string, limit uint64, out chan<- netip.Addr) error {
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

	// Knuth/Newlib constants: full period over any power-of-two modulus.
	const lcgA = uint64(6364136223846793005)
	const lcgC = uint64(1442695040888963407)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Round the modulus up to a power of two so rejection stays under 50%.
	// Past 2^63 there is no bigger power of two, so use the whole ring.
	useFullUint64 := indexer.GrandTotal > 1<<63
	var mSize, state uint64
	if useFullUint64 {
		state = rng.Uint64()
	} else {
		mBits := bits.Len64(indexer.GrandTotal)
		if indexer.GrandTotal&(indexer.GrandTotal-1) == 0 {
			mBits--
		}
		mSize = 1 << mBits
		state = rng.Uint64() % mSize
	}

	reader := bufio.NewReader(file)
	var dispatched, count uint64
	for dispatched < indexer.GrandTotal {
		if limit > 0 && count >= limit {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if useFullUint64 {
			state = state*lcgA + lcgC
		} else {
			state = (state*lcgA + lcgC) % mSize
		}
		if state >= indexer.GrandTotal {
			continue // landed in the padding above GrandTotal
		}

		var ip netip.Addr
		if state < indexer.TotalCIDRIPs {
			ip = indexer.getIPFromCIDRBlocks(state)
		} else {
			var err error
			ip, err = readIPAtCSVOffset(file, reader, indexer.SingleOffsets[state-indexer.TotalCIDRIPs])
			if err != nil {
				// Skip unreadable lines rather than aborting the whole scan.
				dispatched++
				continue
			}
		}
		select {
		case out <- ip:
			count++
		case <-ctx.Done():
			return ctx.Err()
		}
		dispatched++
	}
	return nil
}

// getIPFromCIDRBlocks finds the block owning globalIdx and returns the
// address at that offset inside it.
func (mi *MasterIndexer) getIPFromCIDRBlocks(globalIdx uint64) netip.Addr {
	// Last block whose GlobalIdx <= globalIdx.
	i := sort.Search(len(mi.CIDRBlocks), func(i int) bool {
		return mi.CIDRBlocks[i].GlobalIdx > globalIdx
	}) - 1
	block := mi.CIDRBlocks[i]
	return addOffsetToAddr(block.StartIP, globalIdx-block.GlobalIdx)
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

// readIPAtCSVOffset seeks to a byte offset and parses the first CSV field.
// The reader is reused across calls; Reset drops the stale buffer after Seek.
func readIPAtCSVOffset(file *os.File, reader *bufio.Reader, offset int64) (netip.Addr, error) {
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return netip.Addr{}, err
	}
	reader.Reset(file)
	lineBytes, err := reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return netip.Addr{}, err
	}
	line := strings.TrimSpace(string(lineBytes))
	if i := strings.IndexByte(line, ','); i >= 0 {
		line = strings.TrimSpace(line[:i])
	}
	ip, ok := netutil.ParseIPOrCIDR(line)
	if !ok {
		return netip.Addr{}, fmt.Errorf("invalid IP or CIDR: %q", line)
	}
	return ip.Addr(), nil
}

// saturatingAdd adds a+b, clamping at MaxUint64 instead of wrapping.
func saturatingAdd(a, b uint64) uint64 {
	sum := a + b
	if sum < a {
		return ^uint64(0)
	}
	return sum
}
