package iplist

import (
	"context"
	"fmt"
	"math"
	"net/netip"
	"os"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
)

// ImportOption controls the behaviour of ImportIPList.
type ImportOption struct {
	// MaxMapFileSize is the file-size threshold in bytes below which the
	// in-memory map strategy is used. Files at or above this size fall back
	// to the disk-based (sort + stream) strategy.
	MaxMapFileSize int64

	// MaxInMemoryEntries caps the number of entries held in the dedup map.
	// Zero means unlimited.
	MaxInMemoryEntries uint64
}

func DefaultImportOption() ImportOption {
	return ImportOption{
		MaxMapFileSize:     100_000_000,
		MaxInMemoryEntries: 0,
	}
}

// ImportIPList reads a CSV IP-prefix list from srcPath, deduplicates it, and
// writes the result to dstPath. It picks an in-memory or disk strategy
// automatically based on file size and ImportOption.
func ImportIPList(ctx context.Context, srcPath, dstPath string, option ImportOption) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("stat %q: %w", srcPath, err)
	}
	if srcInfo.IsDir() {
		return fmt.Errorf("%q is a directory", srcPath)
	}

	if srcInfo.Size() <= option.MaxMapFileSize {
		return importIPListMap(ctx, srcPath, dstPath, option)
	}
	return importIPListDisk(ctx, srcPath, dstPath)
}

// importIPListMap deduplicates using an in-memory map. Fast, but bounded by RAM.
func importIPListMap(ctx context.Context, srcPath, dstPath string, option ImportOption) error {
	estimated := estimateMapEntries(srcPath, option.MaxInMemoryEntries)
	seen := make(map[netip.Prefix]struct{}, estimated)

	return WriteCSV(dstPath, func(write func(IPList) error) error {
		return ReadCSV(srcPath, func(entry IPList, _ int64) error {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("context cancelled: %w", err)
			}

			entry.IP = entry.IP.Masked()

			if _, ok := seen[entry.IP]; ok {
				return nil
			}
			if option.MaxInMemoryEntries > 0 && uint64(len(seen)) >= option.MaxInMemoryEntries {
				return fmt.Errorf("in-memory entry limit of %d exceeded", option.MaxInMemoryEntries)
			}

			seen[entry.IP] = struct{}{}
			return write(entry)
		})
	})
}

// importIPListDisk sorts the file on disk first, then streams it once to
// deduplicate adjacent equal prefixes. O(1) memory regardless of file size.
func importIPListDisk(ctx context.Context, srcPath, dstPath string) error {
	tmpFile, tmpPath, err := fileutil.CreateTmpFile("bgscan-import-ip-")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	_ = tmpFile.Close()
	defer func() {
		if err := os.Remove(tmpPath); err != nil && !os.IsNotExist(err) {
			_ = err
		}
	}()

	if err := fileutil.SortFile(ctx, srcPath, tmpPath); err != nil {
		return fmt.Errorf("sort file: %w", err)
	}

	var last netip.Prefix

	return WriteCSV(dstPath, func(write func(IPList) error) error {
		return ReadCSV(tmpPath, func(entry IPList, _ int64) error { // ← sorted tmp, not srcPath
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("context cancelled: %w", err)
			}

			entry.IP = entry.IP.Masked()

			if last.IsValid() && last == entry.IP {
				return nil
			}
			last = entry.IP

			return write(entry)
		})
	})
}

// estimateMapEntries returns a conservative initial capacity for the dedup map.
func estimateMapEntries(path string, max uint64) int {
	const minAvgRowSize = 16

	fi, err := os.Stat(path)
	if err != nil || fi.Size() <= 0 {
		return 0
	}

	n := uint64(fi.Size()) / minAvgRowSize

	if max > 0 && n > max {
		n = max
	}

	const maxInt = uint64(^uint(0) >> 1)
	if n > maxInt {
		n = maxInt
	}

	return int(n)
}

// LoadAll loads the entire IP list file into memory.
// This should only be used for relatively small files.
// For large lists prefer streaming APIs like ReadCSV or StreamActiveIPs.
func LoadAll(path string) ([]IPList, error) {
	items := make([]IPList, 0, 1024)
	err := ReadCSV(path, func(entry IPList, _ int64) error {
		items = append(items, entry)
		return nil
	})

	return items, err
}

func CountIPs(path string) (uint64, error) {
	var total uint64
	err := ReadCSV(path, func(entry IPList, _ int64) error {
		total = saturatingAdd(total, countIPEntry(entry.IP))
		return nil
	})
	return total, err
}

func CountActiveIPs(path string) (uint64, error) {
	var total uint64
	err := ReadCSV(path, func(entry IPList, _ int64) error {
		if !entry.Enable || total == math.MaxUint64 {
			return nil
		}
		total = saturatingAdd(total, countIPEntry(entry.IP))
		return nil
	})
	return total, err
}

func countIPEntry(p netip.Prefix) uint64 {
	hostBits := p.Addr().BitLen() - p.Bits()
	if hostBits >= 64 {
		return math.MaxUint64
	}
	return uint64(1) << hostBits
}
