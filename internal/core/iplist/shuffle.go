package iplist

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
	"strings"
	"time"

	"bgscan/internal/logger"
)

// shuffleMemWarnThreshold is a rough line-count estimate above which
// NewMasterIndexer will log a warning about peak memory usage.
// At ~65 bytes per rawEntry, 5M lines ≈ 325MB peak during indexing.
const shuffleMemWarnThreshold = 5_000_000

// avgLineBytes is the heuristic used to estimate line count from file size.
const avgLineBytes = 20

// NewMasterIndexer builds a hybrid index over an IP-list CSV file so that
// streamActiveIPsShuffled can visit every IP in O(1) space.
//
// It counts all IPs using big.Int to handle ranges larger than uint64.
// If the total fits in uint64, it builds a direct index.
// If it exceeds uint64, it proportionally slices each range at a random
// offset so the ranges share the 2^64 address space evenly.
//
// Memory note: all parsed entries are held in memory during indexing
// (~65 bytes per CSV line). A 1M-line file peaks at ~65MB; a 5M-line
// file peaks at ~325MB. For very large files prefer sequential mode.
func NewMasterIndexer(filePath string) (*MasterIndexer, error) {
	indexer := &MasterIndexer{
		FilePath:      filePath,
		CIDRBlocks:    make([]CIDRBlock, 0),
		SingleOffsets: make([]int64, 0),
	}

	fi, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if fi.Size() == 0 {
		return indexer, nil
	}

	// Warn early if the file is large enough to cause memory pressure.
	if estimatedLines := fi.Size() / avgLineBytes; estimatedLines > shuffleMemWarnThreshold {
		logger.CoreWarn(
			"shuffled mode: file %q has ~%dM estimated lines; peak memory during indexing may exceed 300MB — consider sequential mode for very large lists",
			filePath, estimatedLines/1_000_000,
		)
	}

	type rawEntry struct {
		startIP netip.Addr
		offset  int64
		size    *big.Int
		isCIDR  bool
	}

	// Pre-allocate based on file size to avoid repeated slice doubling.
	estimatedLines := fi.Size() / avgLineBytes
	entries := make([]rawEntry, 0, estimatedLines+1)

	totalCount := new(big.Int)
	one := big.NewInt(1)

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger.CoreError("failed to close file: %v", err)
		}
	}()

	r := bufio.NewReaderSize(f, 64*1024)

	for {
		filePos, _ := f.Seek(0, io.SeekCurrent)
		lineOffset := filePos - int64(r.Buffered())

		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line != "" {
			parts := strings.SplitN(line, ",", 2)
			rec := make([]string, len(parts))
			for i, p := range parts {
				rec[i] = strings.TrimSpace(p)
			}

			row, ok := ParseRecord(rec)
			if ok && row.Enable {
				if !row.IsCIDR() {
					entries = append(entries, rawEntry{offset: lineOffset, size: one, isCIDR: false})
					totalCount.Add(totalCount, one)
				} else {
					prefix, err := netip.ParsePrefix(row.IP)
					if err == nil {
						addr := prefix.Masked().Addr()
						hostBits := 32 - prefix.Bits()
						if addr.Is6() {
							hostBits = 128 - prefix.Bits()
						}
						size := new(big.Int).Lsh(one, uint(hostBits))
						entries = append(entries, rawEntry{
							startIP: addr,
							offset:  lineOffset,
							size:    size,
							isCIDR:  true,
						})
						totalCount.Add(totalCount, size)
					}
				}
			}
		}

		if err == io.EOF {
			break
		}
	}

	if totalCount.Sign() == 0 {
		return indexer, nil
	}

	maxUint64 := new(big.Int).SetUint64(^uint64(0)) // 2^64 - 1
	fitsInUint64 := totalCount.Cmp(maxUint64) <= 0

	// not crypto-sensitive — used only for shuffle offset randomisation
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var globalIdx uint64

	if fitsInUint64 {
		indexer.CIDRBlocks = make([]CIDRBlock, 0, len(entries))
		indexer.SingleOffsets = make([]int64, 0, len(entries))

		for _, e := range entries {
			if !e.isCIDR {
				indexer.SingleOffsets = append(indexer.SingleOffsets, e.offset)
				indexer.TotalSingles++
			} else {
				count := e.size.Uint64()
				indexer.CIDRBlocks = append(indexer.CIDRBlocks, CIDRBlock{
					StartIP:   e.startIP,
					TotalIPs:  count,
					GlobalIdx: globalIdx,
				})
				globalIdx = saturatingAdd(globalIdx, count)
				indexer.TotalCIDRIPs = globalIdx
			}
		}
		indexer.GrandTotal = saturatingAdd(globalIdx, indexer.TotalSingles)

	} else {
		// Total exceeds uint64: proportionally slice each range so they
		// collectively fit within 2^64.
		for _, e := range entries {
			quotaBig := new(big.Int).Mul(e.size, maxUint64)
			quotaBig.Div(quotaBig, totalCount)

			if !quotaBig.IsUint64() || quotaBig.Uint64() == 0 {
				continue
			}
			quota := quotaBig.Uint64()

			if !e.isCIDR {
				indexer.SingleOffsets = append(indexer.SingleOffsets, e.offset)
				indexer.TotalSingles++
				continue
			}

			maxOffsetBig := new(big.Int).Sub(e.size, new(big.Int).SetUint64(quota))

			var offsetBig *big.Int
			if maxOffsetBig.Sign() <= 0 {
				offsetBig = big.NewInt(0)
			} else {
				offsetBig = randBigIntBelow(rng, maxOffsetBig)
			}

			indexer.CIDRBlocks = append(indexer.CIDRBlocks, CIDRBlock{
				StartIP:   addBigOffset(e.startIP, offsetBig),
				TotalIPs:  quota,
				GlobalIdx: globalIdx,
			})
			globalIdx = saturatingAdd(globalIdx, quota)
			indexer.TotalCIDRIPs = globalIdx
		}
		indexer.GrandTotal = saturatingAdd(globalIdx, indexer.TotalSingles)
	}

	return indexer, nil
}

// streamActiveIPsShuffled streams IPs in pseudo-random order without loading
// the entire dataset into memory. It uses a Linear Congruential Generator (LCG)
// with rejection sampling to achieve an O(1)-space permutation over the index
// built by NewMasterIndexer.
func streamActiveIPsShuffled(ctx context.Context, path string, limit uint64, out chan<- string) error {
	indexer, err := NewMasterIndexer(path)
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
			logger.CoreError("error closing file: %v", err)
		}
	}()

	// LCG constants (Knuth / Newlib). These produce a full-period sequence
	// over the uint64 space regardless of starting state.
	const lcgA = uint64(6364136223846793005)
	const lcgC = uint64(1442695040888963407)

	// not crypto-sensitive — shuffle seed only
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// If GrandTotal is large enough to cover more than half of uint64,
	// run the LCG over the full uint64 ring (rejection rate stays low).
	// Otherwise use the next power-of-two above GrandTotal as the modulus
	// so the rejection rate stays below 50%.
	useFullUint64 := indexer.GrandTotal > (uint64(1) << 63)

	var mSize, state uint64
	if useFullUint64 {
		state = rng.Uint64()
	} else {
		mBits := bits.Len64(indexer.GrandTotal)
		if indexer.GrandTotal&(indexer.GrandTotal-1) == 0 {
			mBits-- // GrandTotal is already a power of two; don't double it
		}
		mSize = uint64(1) << mBits
		state = rng.Uint64() % mSize
	}

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
			continue // rejection: outside valid range, advance LCG
		}

		if state < indexer.TotalCIDRIPs {
			ip := indexer.getIPFromCIDRBlocks(state)
			select {
			case out <- ip.String():
				count++
			case <-ctx.Done():
				return ctx.Err()
			}
		} else {
			singleIdx := state - indexer.TotalCIDRIPs
			ip, err := readIPAtCSVOffset(file, indexer.SingleOffsets[singleIdx])
			if err != nil {
				// Stale or corrupt offset — skip and keep going.
				dispatched++
				continue
			}
			select {
			case out <- ip.String():
				count++
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		dispatched++
	}

	return nil
}

// getIPFromCIDRBlocks binary-searches the sorted CIDRBlocks slice to find
// which block owns globalIdx, then computes the specific address within it.
func (mi *MasterIndexer) getIPFromCIDRBlocks(globalIdx uint64) netip.Addr {
	low, high := 0, len(mi.CIDRBlocks)-1
	var target CIDRBlock

	for low <= high {
		mid := (low + high) / 2
		if mi.CIDRBlocks[mid].GlobalIdx <= globalIdx {
			target = mi.CIDRBlocks[mid]
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return addOffsetToAddr(target.StartIP, globalIdx-target.GlobalIdx)
}

// addOffsetToAddr adds a uint64 offset to a netip.Addr using byte-level
// arithmetic, correctly wrapping across byte boundaries.
func addOffsetToAddr(addr netip.Addr, offset uint64) netip.Addr {
	b := addr.As16()
	carry := offset
	for i := 15; i >= 0 && carry > 0; i-- {
		sum := uint64(b[i]) + (carry & 0xff)
		b[i] = byte(sum)
		carry = (carry >> 8) + (sum >> 8)
	}
	result, _ := netip.AddrFromSlice(b[:])
	return result.Unmap()
}

// addBigOffset adds a big.Int offset to a netip.Addr, used when individual
// CIDR ranges exceed uint64 (IPv6 /64 and larger).
func addBigOffset(addr netip.Addr, offset *big.Int) netip.Addr {
	b := addr.As16()
	carry := new(big.Int).Set(offset)
	mask := big.NewInt(0xff)
	for i := 15; i >= 0 && carry.Sign() > 0; i-- {
		sum := new(big.Int).Add(big.NewInt(int64(b[i])), new(big.Int).And(carry, mask))
		b[i] = byte(sum.Int64() & 0xff)
		carry.Rsh(carry, 8)
		carry.Add(carry, new(big.Int).Rsh(sum, 8))
	}
	result, _ := netip.AddrFromSlice(b[:])
	return result.Unmap()
}

// randBigIntBelow returns a cryptographically-unbiased random big.Int in
// [0, max) using rejection sampling.
// not crypto-sensitive — used only for CIDR slice offset randomisation.
func randBigIntBelow(rng *rand.Rand, max *big.Int) *big.Int {
	nbits := max.BitLen()
	bitMask := new(big.Int).Sub(
		new(big.Int).Lsh(big.NewInt(1), uint(nbits)),
		big.NewInt(1),
	)
	for {
		b := make([]byte, (nbits+7)/8)
		for i := range b {
			b[i] = byte(rng.Intn(256))
		}
		n := new(big.Int).SetBytes(b)
		n.And(n, bitMask)
		if n.Cmp(max) < 0 {
			return n
		}
	}
}

// readIPAtCSVOffset seeks to a byte offset in the file and parses the
// first field of the CSV line there as a netip.Addr.
func readIPAtCSVOffset(file *os.File, offset int64) (netip.Addr, error) {
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return netip.Addr{}, err
	}

	reader := bufio.NewReader(file)
	lineBytes, err := reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return netip.Addr{}, err
	}

	line := strings.TrimSpace(string(lineBytes))
	if parts := strings.SplitN(line, ",", 2); len(parts) > 0 {
		line = strings.TrimSpace(parts[0])
	}

	return netip.ParseAddr(line)
}

// saturatingAdd adds two uint64 values, returning ^uint64(0) on overflow
// instead of wrapping.
func saturatingAdd(a, b uint64) uint64 {
	result := a + b
	if result < a {
		return ^uint64(0)
	}
	return result
}
