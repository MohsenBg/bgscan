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

// NewMasterIndexer counts all IPs in the file using big.Int.
// If they fit in uint64, it creates a standard metadata map.
// If they exceed uint64, it proportionally slices each range at a random offset
// so they perfectly share the 2^64 limit.
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

	type rawEntry struct {
		startIP netip.Addr
		offset  int64
		size    *big.Int
		isCIDR  bool
	}

	var entries []rawEntry
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
		buffered := int64(r.Buffered())
		lineOffset := filePos - buffered

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
						entries = append(entries, rawEntry{startIP: addr, offset: lineOffset, size: size, isCIDR: true})
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

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var globalIdx uint64

	if fitsInUint64 {
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

			newStartIP := addBigOffset(e.startIP, offsetBig)

			indexer.CIDRBlocks = append(indexer.CIDRBlocks, CIDRBlock{
				StartIP:   newStartIP,
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

// streamActiveIPsShuffled streams IPs in a pseudo-random order without loading
// the entire dataset into memory. It uses a Linear Congruential Generator (LCG)
// with rejection sampling to achieve O(1) space permutation.
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

	a := uint64(6364136223846793005)
	c := uint64(1442695040888963407)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// FIX: Use full uint64 wraparound if GrandTotal > 2^63.
	// If it's > 2^63, the next power of 2 is 2^64, which we can't store in a uint64
	// variable (1 << 64 becomes 0, causing a divide-by-zero panic on modulo).
	useFullUint64 := indexer.GrandTotal > (uint64(1) << 63)

	var mSize uint64
	var state uint64
	if useFullUint64 {
		state = rng.Uint64()
	} else {
		mBits := bits.Len64(indexer.GrandTotal)
		if indexer.GrandTotal&(indexer.GrandTotal-1) == 0 {
			mBits--
		}
		mSize = uint64(1) << mBits
		state = rng.Uint64() % mSize
	}

	var dispatched uint64 = 0
	var count uint64 = 0

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
			state = state*a + c
		} else {
			state = (state*a + c) % mSize
		}

		if state >= indexer.GrandTotal {
			continue
		}

		if state < indexer.TotalCIDRIPs {
			generatedIP := indexer.getIPFromCIDRBlocks(state)
			select {
			case out <- generatedIP.String():
				count++
			case <-ctx.Done():
				return ctx.Err()
			}
		} else {
			singleIdx := state - indexer.TotalCIDRIPs
			offset := indexer.SingleOffsets[singleIdx]
			generatedIP, err := readIPAtCSVOffset(file, offset)
			if err != nil {
				continue
			}
			select {
			case out <- generatedIP.String():
				count++
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		dispatched++
	}

	return nil
}

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

	offset := globalIdx - target.GlobalIdx
	return addOffsetToAddr(target.StartIP, offset)
}

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

func randBigIntBelow(rng *rand.Rand, max *big.Int) *big.Int {
	nbits := max.BitLen()
	for {
		b := make([]byte, (nbits+7)/8)
		for i := range b {
			b[i] = byte(rng.Intn(256))
		}
		n := new(big.Int).SetBytes(b)
		n.And(n, new(big.Int).Sub(
			new(big.Int).Lsh(big.NewInt(1), uint(nbits)),
			big.NewInt(1),
		))
		if n.Cmp(max) < 0 {
			return n
		}
	}
}

func readIPAtCSVOffset(file *os.File, offset int64) (netip.Addr, error) {
	_, err := file.Seek(offset, 0)
	if err != nil {
		return netip.Addr{}, err
	}

	reader := bufio.NewReader(file)
	lineBytes, err := reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return netip.Addr{}, err
	}

	line := strings.TrimSpace(string(lineBytes))
	if parts := strings.Split(line, ","); len(parts) > 0 {
		line = strings.TrimSpace(parts[0])
	}

	ipAddr, err := netip.ParseAddr(line)
	if err != nil {
		return netip.Addr{}, err
	}
	return ipAddr, nil
}

func saturatingAdd(a, b uint64) uint64 {
	result := a + b
	if result < a {
		return ^uint64(0)
	}
	return result
}
