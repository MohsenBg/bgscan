package target

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
)

// ImportOption tunes how Import deduplicates a list.
type ImportOption struct {
	// MaxMapFileSize is the size threshold below which the in-memory map
	// strategy is used. Bigger files fall back to sort + stream on disk.
	MaxMapFileSize int64
	// MaxInMemoryEntries caps the dedup map size. Zero means unlimited.
	MaxInMemoryEntries uint64
}

func DefaultImportOption() ImportOption {
	return ImportOption{
		MaxMapFileSize:     100_000_000,
		MaxInMemoryEntries: 0,
	}
}

// ImportResult reports how many entries of each kind were imported.
type ImportResult struct {
	IPs     uint64
	Domains uint64
}

// Import reads a mixed target list from srcPath and splits it into
// targets/ips/<name>.csv and targets/domains/<name>.csv. Each kind is
// deduplicated separately, and a file is only written when that kind has at
// least one entry. Existing files are never overwritten.
func Import(ctx context.Context, srcPath, name string, option ImportOption) (ImportResult, error) {
	var result ImportResult

	if err := ctx.Err(); err != nil {
		return result, fmt.Errorf("context cancelled: %w", err)
	}

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return result, fmt.Errorf("stat %q: %w", srcPath, err)
	}
	if srcInfo.IsDir() {
		return result, fmt.Errorf("%q is a directory", srcPath)
	}

	ipDst, err := ResolveTargetFilePath(name, KindIP)
	if err != nil {
		return result, err
	}
	domainDst, err := ResolveTargetFilePath(name, KindDomain)
	if err != nil {
		return result, err
	}

	if TargetFileExists(name, KindIP) || TargetFileExists(name, KindDomain) {
		return result, fmt.Errorf("target list %q already exists", name)
	}

	ipTmp, err := createTempFile("bgscan-import-ip-")
	if err != nil {
		return result, err
	}
	defer func() { _ = os.Remove(ipTmp) }()

	domainTmp, err := createTempFile("bgscan-import-domain-")
	if err != nil {
		return result, err
	}
	defer func() { _ = os.Remove(domainTmp) }()

	// Split the source in one pass. The writer callbacks keep both temp files
	// open, and the parser has already normalized every entry.
	var split ImportResult

	err = WriteCSV(ipTmp, func(writeIP func(Entry) error) error {
		return WriteCSV(domainTmp, func(writeDomain func(Entry) error) error {
			return ReadCSV(srcPath, func(entry Entry, _ int64) error {
				if err := ctx.Err(); err != nil {
					return fmt.Errorf("context cancelled: %w", err)
				}

				switch entry.Kind {
				case KindIP:
					entry.IP = entry.IP.Masked()
					split.IPs++
					return writeIP(entry)
				case KindDomain:
					split.Domains++
					return writeDomain(entry)
				default:
					return nil
				}
			})
		})
	})

	if err != nil {
		return result, fmt.Errorf("split list: %w", err)
	}

	if split.IPs == 0 && split.Domains == 0 {
		return result, fmt.Errorf("no valid targets in %q", srcPath)
	}

	type pendingImport struct {
		tmp string
		dst string
	}

	var pending []pendingImport

	cleanup := func() {
		for _, p := range pending {
			_ = os.Remove(p.tmp)
		}
	}

	// Deduplicate into temp files in the destination directory first, then
	// rename both. A failure on the domain side therefore cannot leave a
	// half-finished IP import behind.
	if split.IPs > 0 {
		tmp, count, err := dedupeToTemp(ctx, ipTmp, ipDst, option)
		if err != nil {
			cleanup()
			return result, fmt.Errorf("import IP targets: %w", err)
		}
		result.IPs = count
		pending = append(pending, pendingImport{tmp: tmp, dst: ipDst})
	}

	if split.Domains > 0 {
		tmp, count, err := dedupeToTemp(ctx, domainTmp, domainDst, option)
		if err != nil {
			cleanup()
			return result, fmt.Errorf("import domain targets: %w", err)
		}
		result.Domains = count
		pending = append(pending, pendingImport{tmp: tmp, dst: domainDst})
	}

	for _, p := range pending {
		if err := os.Rename(p.tmp, p.dst); err != nil {
			cleanup()
			return result, fmt.Errorf("publish %q: %w", p.dst, err)
		}
	}

	return result, nil
}

func createTempFile(pattern string) (string, error) {
	f, path, err := fileutil.CreateTmpFile(pattern)
	if err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("close temp file: %w", err)
	}
	return path, nil
}

// dedupeToTemp writes a deduplicated copy of srcPath into a temporary file
// inside dstPath's directory and returns that path and the entry count. The
// caller renames it into place once every kind has been written successfully.
func dedupeToTemp(ctx context.Context, srcPath, dstPath string, option ImportOption) (string, uint64, error) {
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return "", 0, fmt.Errorf("create target directory: %w", err)
	}

	f, err := os.CreateTemp(filepath.Dir(dstPath), ".bgscan-import-*")
	if err != nil {
		return "", 0, fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := f.Name()
	if err := f.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", 0, fmt.Errorf("close temp file: %w", err)
	}

	count, err := importFile(ctx, srcPath, tmpPath, option)
	if err != nil {
		_ = os.Remove(tmpPath)
		return "", 0, err
	}

	if err := os.Chmod(tmpPath, 0o644); err != nil {
		_ = os.Remove(tmpPath)
		return "", 0, fmt.Errorf("chmod temp file: %w", err)
	}

	return tmpPath, count, nil
}

func importFile(ctx context.Context, srcPath, dstPath string, option ImportOption) (uint64, error) {
	fi, err := os.Stat(srcPath)
	if err != nil {
		return 0, fmt.Errorf("stat %q: %w", srcPath, err)
	}

	if fi.Size() <= option.MaxMapFileSize {
		return importMap(ctx, srcPath, dstPath, option)
	}
	return importDisk(ctx, srcPath, dstPath)
}

// importMap deduplicates with an in-memory map. Fast, but bounded by RAM.
func importMap(ctx context.Context, srcPath, dstPath string, option ImportOption) (uint64, error) {
	seen := make(map[string]struct{}, estimateMapEntries(srcPath, option.MaxInMemoryEntries))
	var written uint64

	err := WriteCSV(dstPath, func(write func(Entry) error) error {
		return ReadCSV(srcPath, func(entry Entry, _ int64) error {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("context cancelled: %w", err)
			}
			key := entry.Key()
			if _, ok := seen[key]; ok {
				return nil
			}
			if option.MaxInMemoryEntries > 0 && uint64(len(seen)) >= option.MaxInMemoryEntries {
				return fmt.Errorf("in-memory entry limit of %d exceeded", option.MaxInMemoryEntries)
			}
			seen[key] = struct{}{}
			if err := write(entry); err != nil {
				return err
			}
			written++
			return nil
		})
	})

	return written, err
}

// importDisk sorts the file on disk, then streams it once to drop adjacent
// duplicates. Flat memory regardless of file size.
func importDisk(ctx context.Context, srcPath, dstPath string) (uint64, error) {
	tmpFile, tmpPath, err := fileutil.CreateTmpFile("bgscan-sort-")
	if err != nil {
		return 0, fmt.Errorf("create temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return 0, fmt.Errorf("close temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmpPath) }()

	if err := fileutil.SortFile(ctx, srcPath, tmpPath); err != nil {
		return 0, fmt.Errorf("sort file: %w", err)
	}

	var lastKey string // Key() is never empty, so "" is a safe sentinel
	var written uint64

	err = WriteCSV(dstPath, func(write func(Entry) error) error {
		return ReadCSV(tmpPath, func(entry Entry, _ int64) error {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("context cancelled: %w", err)
			}
			key := entry.Key()
			if key == lastKey {
				return nil
			}
			lastKey = key
			if err := write(entry); err != nil {
				return err
			}
			written++
			return nil
		})
	})

	return written, err
}

// estimateMapEntries guesses a starting capacity for the dedup map.
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

// LoadAll reads the whole list into memory. Fine for small files; for big
// ones prefer ReadCSV or StreamActive.
func LoadAll(path string) ([]Entry, error) {
	items := make([]Entry, 0, 1024)
	err := ReadCSV(path, func(entry Entry, _ int64) error {
		items = append(items, entry)
		return nil
	})
	return items, err
}

func CountAll(path string) (uint64, error) {
	var total uint64
	err := ReadCSV(path, func(entry Entry, _ int64) error {
		total = saturatingAdd(total, countEntry(entry))
		return nil
	})
	return total, err
}

func CountActive(path string) (uint64, error) {
	var total uint64
	err := ReadCSV(path, func(entry Entry, _ int64) error {
		if !entry.Enable || total == math.MaxUint64 {
			return nil
		}
		total = saturatingAdd(total, countEntry(entry))
		return nil
	})
	return total, err
}

// countEntry is how many scan targets an entry stands for: one for a domain,
// every address in the prefix for an IP. Prefixes wider than /64 saturate.
func countEntry(e Entry) uint64 {
	if e.Kind != KindIP {
		return 1
	}
	hostBits := e.IP.Addr().BitLen() - e.IP.Bits()
	if hostBits >= 64 {
		return math.MaxUint64
	}
	return uint64(1) << hostBits
}
