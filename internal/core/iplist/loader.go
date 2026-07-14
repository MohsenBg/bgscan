package iplist

import (
	"fmt"
	"math"
	"net/netip"
)

// CopyIPFile copies an IP list file from srcPath to dstPath.
// Each entry is validated and normalized through the CSV parser,
// so invalid rows are automatically skipped.
func CopyIPFile(srcPath, dstPath string) error {
	return WriteCSV(dstPath, func(write func(IPList) error) error {
		return ReadCSV(srcPath, func(entry IPList, _ int64) error {
			return write(entry)
		})
	})
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

// CountIPs counts the total number of valid entries in an IP list file.
func CountIPs(path string) (uint64, error) {
	var total uint64
	err := ReadCSV(path, func(entry IPList, _ int64) error {
		total += countIPEntry(entry.IP)
		return nil
	})

	return total, err
}

// CountActiveIPs counts the number of enabled entries in the file.
func CountActiveIPs(path string) (uint64, error) {
	var total uint64
	err := ReadCSV(path, func(entry IPList, _ int64) error {
		if total == math.MaxUint64 {
			return nil
		}

		if entry.Enable {
			total += countIPEntry(entry.IP)
		}
		return nil
	})

	return total, err
}

// ValidateFile verifies that a file is a valid IP list.
// It ensures all parsed rows contain a valid normalized IP or CIDR.
func ValidateFile(path string) error {
	var (
		total   int
		invalid int
	)

	err := ReadCSV(path, func(entry IPList, _ int64) error {
		total++

		if entry.IP == "" {
			invalid++
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("read ip list %q: %w", path, err)
	}

	if invalid > 0 {
		return fmt.Errorf("invalid entries: %d of %d", invalid, total)
	}

	return nil
}

// MergeFiles merges multiple IP list files into a single destination file.
//
// Duplicate IP/CIDR entries are removed. The first occurrence is kept
// and later duplicates are skipped.
func MergeFiles(dstPath string, srcPaths ...string) error {
	seen := make(map[string]struct{}, 1024)

	return WriteCSV(dstPath, func(write func(IPList) error) error {
		for _, src := range srcPaths {

			err := ReadCSV(src, func(entry IPList, _ int64) error {
				if _, ok := seen[entry.IP]; ok {
					return nil
				}

				seen[entry.IP] = struct{}{}
				return write(entry)
			})
			if err != nil {
				return fmt.Errorf("merge file %q: %w", src, err)
			}
		}

		return nil
	})
}

// countIPEntry handles total IP calculation using zero-allocation netip primitives.
func countIPEntry(ipStr string) uint64 {
	if prefix, err := netip.ParsePrefix(ipStr); err == nil {
		addr := prefix.Masked().Addr()

		maxBits := 32
		if addr.Is6() {
			maxBits = 128
		}

		hostBits := maxBits - prefix.Bits()

		if hostBits >= 64 {
			return math.MaxUint64
		}

		return uint64(1) << hostBits
	}

	if _, err := netip.ParseAddr(ipStr); err == nil {
		return 1
	}

	return 0
}
