package result

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/MohsenBg/bgscan/internal/core/fileutil"
	"github.com/MohsenBg/bgscan/internal/logger"
)

// mergeResults merges a batch of results into the main result file.
//
// Guarantees:
//   - Sorted output by score (descending)
//   - Duplicate replacement by Key()
//   - Streaming merge with existing file
//   - Atomic file replacement
func mergeResults(
	log *logger.Logger,
	resultPath string,
	schema ResultSchema,
	results []Result,
) error {
	if len(results) == 0 {
		return nil
	}

	// Sort by score descending, then by key ascending to group duplicates.
	slices.SortFunc(results, func(a, b Result) int {
		if a.Score() > b.Score() {
			return -1
		}
		if a.Score() < b.Score() {
			return 1
		}
		return strings.Compare(a.Key(), b.Key())
	})

	// Keep only the highest-scored record per key.
	results = slices.CompactFunc(results, func(a, b Result) bool {
		return a.Key() == b.Key()
	})

	tmpPath := resultPath + ".tmp"
	if err := os.MkdirAll(filepath.Dir(tmpPath), 0o755); err != nil {
		return fmt.Errorf("create result dir: %w", err)
	}

	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	// Ensure the temp file is cleaned up if the merge fails at any point.
	success := false
	defer func() {
		if !success {
			_ = out.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	bw := fileutil.NewBufferedWriter(out, 0)
	cw := csv.NewWriter(bw)

	if fileutil.CheckFileExists(resultPath) {
		if err := mergeWithExisting(log, resultPath, schema, results, cw); err != nil {
			return err
		}
	} else {
		if err := writeResults(results, cw); err != nil {
			return err
		}
	}

	if err := finalizeFile(cw, bw, out); err != nil {
		return err
	}

	success = true
	if err := out.Close(); err != nil {
		return err
	}

	if err := replaceFile(tmpPath, resultPath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	return syncDir(filepath.Dir(resultPath))
}

// mergeWithExisting streams the existing CSV and merges it with the new batch.
// It uses a binary search index over the batch keys to efficiently find replacements.
func mergeWithExisting(
	log *logger.Logger,
	resultPath string,
	schema ResultSchema,
	delta []Result,
	cw *csv.Writer,
) error {
	// Build an index of delta positions sorted by key for O(log N) lookups.
	keyOrder := make([]int, len(delta))
	for i := range keyOrder {
		keyOrder[i] = i
	}
	slices.SortFunc(keyOrder, func(a, b int) int {
		return strings.Compare(delta[a].Key(), delta[b].Key())
	})

	holds := func(key string) (int, bool) {
		pos, found := slices.BinarySearchFunc(keyOrder, key, func(order int, target string) int {
			return strings.Compare(delta[order].Key(), target)
		})
		if !found {
			return 0, false
		}
		return keyOrder[pos], true
	}

	emitted := make([]bool, len(delta))
	cursor := 0

	// emitDelta writes a delta record if it hasn't been written yet.
	emitDelta := func(idx int) error {
		if emitted[idx] {
			return nil
		}
		emitted[idx] = true
		return cw.Write(delta[idx].ToRecord())
	}

	_, err := ReadCSV(log, resultPath, schema, func(existing Result) error {
		// Emit any delta records that score strictly higher than the current existing record.
		for cursor < len(delta) && delta[cursor].Score() > existing.Score() {
			if err := emitDelta(cursor); err != nil {
				return err
			}
			cursor++
		}

		// If the existing record is superseded by a delta record, emit the delta version.
		if idx, found := holds(existing.Key()); found {
			return emitDelta(idx)
		}

		// Otherwise, keep the existing record.
		return cw.Write(existing.ToRecord())
	})
	if err != nil {
		return err
	}

	// Emit any remaining delta records (those with scores lower than all existing records).
	for i := cursor; i < len(delta); i++ {
		if err := emitDelta(i); err != nil {
			return err
		}
	}

	return nil
}

func writeResults(results []Result, cw *csv.Writer) error {
	for _, r := range results {
		if err := cw.Write(r.ToRecord()); err != nil {
			return err
		}
	}
	return nil
}

func finalizeFile(cw *csv.Writer, bw *bufio.Writer, out *os.File) error {
	cw.Flush()
	if err := cw.Error(); err != nil {
		return err
	}
	if err := bw.Flush(); err != nil {
		return err
	}
	return out.Sync()
}
