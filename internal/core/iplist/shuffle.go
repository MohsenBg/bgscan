package iplist

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math/bits"
	"math/rand"
	"net/netip"
	"os"
	"strings"
	"time"
)

// NewMasterIndexer parses data records through the CSV engine to register positions.
func NewMasterIndexer(filePath string) (*MasterIndexer, error) {
	indexer := &MasterIndexer{
		FilePath:      filePath,
		CIDRBlocks:    make([]CIDRBlock, 0),
		SingleOffsets: make([]int64, 0),
	}

	err := ReadCSV(filePath, func(row IPList, offset int64) error {
		if !row.Enable {
			return nil
		}

		if row.IsCIDR() {
			prefix, err := netip.ParsePrefix(row.IP)
			if err != nil {
				return nil
			}
			addr := prefix.Masked().Addr()
			maxBits := 32
			if addr.Is6() {
				maxBits = 128
			}
			count := uint64(1) << (maxBits - prefix.Bits())

			indexer.CIDRBlocks = append(indexer.CIDRBlocks, CIDRBlock{
				StartIP:   addr,
				TotalIPs:  count,
				GlobalIdx: indexer.TotalCIDRIPs,
			})
			indexer.TotalCIDRIPs += count
		} else {
			indexer.SingleOffsets = append(indexer.SingleOffsets, offset)
			indexer.TotalSingles++
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	indexer.GrandTotal = indexer.TotalCIDRIPs + indexer.TotalSingles
	return indexer, nil
}

// streamActiveIPsShuffled handles the LCG generation engine routines.
func streamActiveIPsShuffled(ctx context.Context, path string, limit int, out chan<- string) error {
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
	defer file.Close()

	mBits := bits.Len64(indexer.GrandTotal)
	if indexer.GrandTotal&(indexer.GrandTotal-1) == 0 {
		mBits--
	}
	mSize := uint64(1) << mBits

	a := uint64(6364136223846793005) | 1
	c := uint64(1442695040888963407)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	state := rng.Uint64() % mSize

	var dispatched uint64 = 0
	count := 0

	for dispatched < indexer.GrandTotal {
		if limit > 0 && count >= limit {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		state = (state*a + c) % mSize

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
	ipAddr := target.StartIP
	for range offset {
		ipAddr = ipAddr.Next()
	}
	return ipAddr
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
