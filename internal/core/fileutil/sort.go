package fileutil

import (
	"bufio"
	"container/heap"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	inMemoryThreshold = 100 * 1024 * 1024 // 100MB
	defaultChunkLines = 100_000
)

// SortFile sorts inputFile line by line and writes the result to outputFile.
//
// Files up to 100MB are sorted in memory. Bigger ones are cut into chunks,
// sorted separately, and merged back, so memory stays flat no matter how large
// the input gets.
func SortFile(ctx context.Context, inputFile, outputFile string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled before start: %w", err)
	}

	info, err := os.Stat(inputFile)
	if err != nil {
		return fmt.Errorf("stat input file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputFile), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	if info.Size() <= inMemoryThreshold {
		return sortInMemory(ctx, inputFile, outputFile)
	}

	return sortExternal(ctx, inputFile, outputFile)
}

// sortInMemory reads the whole file, sorts it, and writes it back out. Used
// when the file is small enough that holding every line is cheap.
func sortInMemory(ctx context.Context, inputFile, outputFile string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	text := strings.TrimRight(string(data), "\n")

	var lines []string
	if text != "" {
		lines = strings.Split(text, "\n")
	}
	sort.Strings(lines)

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled after sort: %w", err)
	}

	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer func() {
		_ = out.Close()
	}()

	w := bufio.NewWriter(out)
	for _, l := range lines {
		if _, err := fmt.Fprintln(w, l); err != nil {
			return fmt.Errorf("write line: %w", err)
		}
	}
	return w.Flush()
}

// sortExternal spills sorted chunks into the system temp directory and merges
// them. The chunks are scratch, so they do not need to sit next to the output.
func sortExternal(ctx context.Context, inputFile, outputFile string) error {
	tmpDir, err := os.MkdirTemp("", "bgscan-chunks-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	chunkFiles, err := splitAndSort(ctx, inputFile, tmpDir)
	if err != nil {
		return fmt.Errorf("split and sort: %w", err)
	}

	return mergeChunks(ctx, chunkFiles, outputFile)
}

// splitAndSort cuts the input into chunks small enough to sort in memory.
func splitAndSort(ctx context.Context, inputFile, tmpDir string) ([]string, error) {
	f, err := os.Open(inputFile)
	if err != nil {
		return nil, fmt.Errorf("open input file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 10*1024*1024), 10*1024*1024)

	var chunkFiles []string
	var lines []string
	chunkNum := 0

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context cancelled during scan: %w", err)
		}

		lines = append(lines, scanner.Text())
		if len(lines) >= defaultChunkLines {
			name, err := writeSortedChunk(lines, tmpDir, chunkNum)
			if err != nil {
				return nil, fmt.Errorf("write chunk %d: %w", chunkNum, err)
			}
			chunkFiles = append(chunkFiles, name)
			lines = lines[:0]
			chunkNum++
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}
	if len(lines) > 0 {
		name, err := writeSortedChunk(lines, tmpDir, chunkNum)
		if err != nil {
			return nil, fmt.Errorf("write final chunk: %w", err)
		}
		chunkFiles = append(chunkFiles, name)
	}

	return chunkFiles, nil
}

// writeSortedChunk sorts one chunk and writes it to dir.
func writeSortedChunk(lines []string, dir string, n int) (string, error) {
	sort.Strings(lines)
	name := filepath.Join(dir, fmt.Sprintf("chunk_%d.tmp", n))

	f, err := os.Create(name)
	if err != nil {
		return "", fmt.Errorf("create chunk file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	w := bufio.NewWriter(f)
	for _, l := range lines {
		if _, err := fmt.Fprintln(w, l); err != nil {
			return "", fmt.Errorf("write chunk line: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		return "", fmt.Errorf("flush chunk: %w", err)
	}
	return name, nil
}

// item is the next line from one chunk, plus the chunk it came from.
type item struct {
	line    string
	fileIdx int
}

type minHeap []item

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].line < h[j].line }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x any)        { *h = append(*h, x.(item)) }
func (h *minHeap) Pop() any          { old := *h; n := len(old); x := old[n-1]; *h = old[:n-1]; return x }

// mergeChunks merges the sorted chunks into one stream with a min-heap. Only
// one line per chunk is held at a time.
func mergeChunks(ctx context.Context, chunkFiles []string, outputFile string) error {
	files := make([]*os.File, len(chunkFiles))
	readers := make([]*bufio.Scanner, len(chunkFiles))

	h := &minHeap{}
	heap.Init(h)

	for i, name := range chunkFiles {
		f, err := os.Open(name)
		if err != nil {
			return fmt.Errorf("open chunk %d: %w", i, err)
		}
		files[i] = f
		readers[i] = bufio.NewScanner(f)
		if readers[i].Scan() {
			heap.Push(h, item{readers[i].Text(), i})
		}
	}

	defer func() {
		for _, f := range files {
			_ = f.Close()
		}
	}()

	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer func() {
		_ = out.Close()
	}()

	w := bufio.NewWriter(out)
	for h.Len() > 0 {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled during merge: %w", err)
		}

		min := heap.Pop(h).(item)
		if _, err := fmt.Fprintln(w, min.line); err != nil {
			return fmt.Errorf("write merged line: %w", err)
		}
		if readers[min.fileIdx].Scan() {
			heap.Push(h, item{readers[min.fileIdx].Text(), min.fileIdx})
		}
	}

	return w.Flush()
}
