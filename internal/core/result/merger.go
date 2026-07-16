package result

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"bgscan/internal/core/fileutil"
)

// mergeResults merges a batch of results into the main result file.
//
// Guarantees:
//   - Sorted output by score (higher score first)
//   - Duplicate replacement by Key()
//   - Streaming merge with existing file
//   - Atomic file replacement
func mergeResults(
	resultPath string,
	schema ResultSchema,
	results []Result,
) error {
	if len(results) == 0 {
		return nil
	}

	// Higher score first.
	slices.SortFunc(results, func(a, b Result) int {
		switch {
		case a.Score() > b.Score():
			return -1
		case a.Score() < b.Score():
			return 1
		default:
			return 0
		}
	})

	tmpPath := resultPath + ".tmp"

	dir := filepath.Dir(tmpPath)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir failed: %w", err)
	}

	out, err := os.OpenFile(
		tmpPath,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0o644,
	)
	if err != nil {
		return err
	}

	bw := bufio.NewWriterSize(out, DefaultBatchSize)
	cw := csv.NewWriter(bw)

	cleanup := func(err error) error {
		_ = out.Close()
		_ = os.Remove(tmpPath)
		return err
	}

	if fileutil.CheckFileExists(resultPath) {
		if err := mergeWithExisting(
			resultPath,
			schema,
			results,
			cw,
		); err != nil {
			return cleanup(err)
		}
	} else {
		if err := writeResults(results, cw); err != nil {
			return cleanup(err)
		}
	}

	if err := finalizeFile(cw, bw, out); err != nil {
		return cleanup(err)
	}

	if err := replaceFile(tmpPath, resultPath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	return syncDir(filepath.Dir(resultPath))
}

// mergeWithExisting merges new results with existing CSV.
//
// Existing file is streamed.
// Only one existing record is kept in memory.
func mergeWithExisting(
	resultPath string,
	schema ResultSchema,
	delta []Result,
	cw *csv.Writer,
) error {
	index := 0

	err := ReadCSV(
		resultPath,
		schema,
		func(existing Result) error {
			// Write better scored new results first.
			for index < len(delta) {

				current := delta[index]

				if current.Score() <= existing.Score() {
					break
				}

				if err := cw.Write(
					current.ToRecord(),
				); err != nil {
					return err
				}

				index++
			}

			// Replace duplicate record.
			if index < len(delta) {

				current := delta[index]

				if current.Key() == existing.Key() {

					if err := cw.Write(
						current.ToRecord(),
					); err != nil {
						return err
					}

					index++

					return nil
				}
			}

			// Keep old record.
			return cw.Write(existing.ToRecord())
		},
	)
	if err != nil {
		return err
	}

	// Write remaining new results.
	for ; index < len(delta); index++ {
		if err := cw.Write(
			delta[index].ToRecord(),
		); err != nil {
			return err
		}
	}

	return nil
}

// writeResults writes results directly to CSV.
func writeResults(
	results []Result,
	cw *csv.Writer,
) error {
	for _, r := range results {
		if err := cw.Write(
			r.ToRecord(),
		); err != nil {
			return err
		}
	}

	return nil
}

// finalizeFile flushes buffers and syncs the file.
func finalizeFile(
	cw *csv.Writer,
	bw *bufio.Writer,
	out *os.File,
) error {
	cw.Flush()

	if err := cw.Error(); err != nil {
		return err
	}

	if err := bw.Flush(); err != nil {
		return err
	}

	if err := out.Sync(); err != nil {
		return err
	}

	return out.Close()
}
