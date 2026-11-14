package result

import (
	"bufio"
	"encoding/csv"
	"io"
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

	// Keep only the highest-scored record per key. remove duplicates
	seen := make(map[string]struct{}, len(results))
	deduped := results[:0]
	for _, r := range results {
		if _, ok := seen[r.Key()]; ok {
			continue
		}
		seen[r.Key()] = struct{}{}
		deduped = append(deduped, r)
	}
	results = deduped

	return fileutil.WriteFileAtomicFunc(resultPath, fileutil.DefaultFileMode, func(w io.Writer) error {
		bw := fileutil.NewBufferedWriter(w)
		cw := fileutil.NewCSVWriter(bw, csvConfig)

		if fileutil.FileExists(resultPath) {
			if err := mergeWithExisting(log, resultPath, schema, results, cw); err != nil {
				return err
			}
		} else {
			if err := writeResults(results, cw); err != nil {
				return err
			}
		}

		return finalizeWriters(cw, bw)
	})
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
		for cursor < len(delta) && delta[cursor].Score() > existing.Score() {
			if err := emitDelta(cursor); err != nil {
				return err
			}
			cursor++
		}

		if idx, found := holds(existing.Key()); found {
			if delta[idx].Score() > existing.Score() {
				return emitDelta(idx)
			}
			emitted[idx] = true
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

func finalizeWriters(cw *csv.Writer, bw *bufio.Writer) error {
	cw.Flush()
	if err := cw.Error(); err != nil {
		return err
	}
	return bw.Flush()
}
