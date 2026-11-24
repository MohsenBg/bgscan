package fileutil

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	minBufferSize     = 4 * 1024
	defaultBufferSize = 64 * 1024
	maxBufferSize     = 4 * 1024 * 1024
)

// CSVConfig changes how a CSV file is read or written. The zero value is the
// usual comma-separated format with no header.
type CSVConfig struct {
	HasHeader        bool
	LazyQuotes       bool
	TrimLeadingSpace bool
	Comma            rune
	FieldsPerRecord  int
	BufferSize       int // Read and write buffer size in bytes. Zero or less uses 64KB, and the value is clamped to 4KB to 4MB.
}

func applyReaderConfig(r *csv.Reader, cfg CSVConfig) {
	if cfg.Comma != 0 {
		r.Comma = cfg.Comma
	}
	if cfg.FieldsPerRecord != 0 {
		r.FieldsPerRecord = cfg.FieldsPerRecord
	} else {
		r.FieldsPerRecord = -1
	}
	r.LazyQuotes = cfg.LazyQuotes
	r.TrimLeadingSpace = cfg.TrimLeadingSpace
}

func applyWriterConfig(w *csv.Writer, cfg CSVConfig) {
	if cfg.Comma != 0 {
		w.Comma = cfg.Comma
	}
}

// StreamCSV reads path one record at a time. Nothing is kept in memory after
// the handler returns, so the file can be as large as the disk allows.
func StreamCSV(path string, cfg CSVConfig, handler func([]string) error) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open csv file: %w", err)
	}

	defer func() {
		_ = f.Close()
	}()

	r := csv.NewReader(bufio.NewReaderSize(f, clampBufferSize(cfg.BufferSize)))
	applyReaderConfig(r, cfg)

	if cfg.HasHeader {
		if _, err := r.Read(); err != nil && err != io.EOF {
			return fmt.Errorf("read csv header: %w", err)
		}
	}

	for {
		rec, err := r.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("csv read error: %w", err)
		}

		if err := handler(rec); err != nil {
			return err
		}
	}
}

// StreamCSVWithOffset is StreamCSV plus the byte offset where each record
// starts.
//
// The offsets come from the raw file, so a caller can seek straight back to a
// record later.
func StreamCSVWithOffset(path string, cfg CSVConfig, handler func(record []string, offset int64) error) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open csv file: %w", err)
	}

	defer func() {
		_ = f.Close()
	}()

	br := bufio.NewReader(f)
	var currentByteOffset int64 = 0

	if cfg.HasHeader {
		header, err := br.ReadBytes('\n')
		currentByteOffset += int64(len(header))

		if err != nil && err != io.EOF {
			return fmt.Errorf("read csv header: %w", err)
		}
		if err == io.EOF {
			return nil
		}
	}

	for {
		lineOffset := currentByteOffset

		lineBytes, err := br.ReadBytes('\n')
		currentByteOffset += int64(len(lineBytes))

		if err != nil && err != io.EOF {
			return fmt.Errorf("read line bytes error: %w", err)
		}

		line := strings.TrimSpace(string(lineBytes))
		if line == "" || strings.HasPrefix(line, "#") {
			if err == io.EOF {
				break
			}
			continue
		}

		sr := strings.NewReader(line)
		cr := csv.NewReader(sr)
		applyReaderConfig(cr, cfg)

		record, err := cr.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		if err := handler(record, lineOffset); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if err == io.EOF {
			break
		}
	}
	return nil
}

// StreamCSVStrict reads every record and stops at the first malformed one.
//
// StreamCSVWithOffset drops a bad line and keeps going, which is fine when a
// list only needs the rows it can read.
func StreamCSVStrict(path string, cfg CSVConfig, handler func(record []string, line int64) error) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open csv file: %w", err)
	}

	defer func() {
		_ = f.Close()
	}()

	r := csv.NewReader(bufio.NewReaderSize(f, clampBufferSize(cfg.BufferSize)))
	applyReaderConfig(r, cfg)

	if cfg.HasHeader {
		if _, err := r.Read(); err != nil && err != io.EOF {
			return fmt.Errorf("read csv header: %w", err)
		}
	}

	var line int64

	for {
		rec, err := r.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			if parseErr, ok := errors.AsType[*csv.ParseError](err); ok {
				return fmt.Errorf(
					"csv parse error at line %d, column %d: %w",
					parseErr.Line,
					parseErr.Column,
					parseErr.Err,
				)
			}
			return fmt.Errorf("csv read error: %w", err)
		}

		line++
		if err := handler(rec, line); err != nil {
			return err
		}
	}
}

// StreamCSVToChan pushes each record into out. The caller owns the channel and
// has to keep reading from it.
func StreamCSVToChan(path string, cfg CSVConfig, out chan<- []string) error {
	return StreamCSV(path, cfg, func(rec []string) error {
		out <- rec
		return nil
	})
}

// WriteCSVFile replaces path with records.
func WriteCSVFile(path string, cfg CSVConfig, records [][]string) error {
	if err := EnsureParentDir(path); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create csv file %s : %w", path, err)
	}

	defer func() {
		_ = f.Close()
	}()

	bw := bufio.NewWriterSize(f, clampBufferSize(cfg.BufferSize))
	w := csv.NewWriter(bw)
	applyWriterConfig(w, cfg)

	// WriteAll flushes the CSV writer, so a buffered write error shows up here.
	if err := w.WriteAll(records); err != nil {
		return fmt.Errorf("write csv records: %w", err)
	}

	return bw.Flush()
}

// StreamWriteCSV gives fn a write callback and turns whatever it produces into
// a CSV file. Rows are written as they come, so the caller never builds the
// whole file in memory.
func StreamWriteCSV(path string, cfg CSVConfig, fn func(write func([]string) error) error) error {
	if err := EnsureParentDir(path); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create csv file: %w", err)
	}

	defer func() {
		_ = f.Close()
	}()

	bw := bufio.NewWriterSize(f, clampBufferSize(cfg.BufferSize))
	w := csv.NewWriter(bw)
	applyWriterConfig(w, cfg)

	writeFunc := func(rec []string) error {
		return w.Write(rec)
	}

	if err := fn(writeFunc); err != nil {
		return err
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}

	return bw.Flush()
}

// AppendCSVRow adds one row to path.
func AppendCSVRow(path string, cfg CSVConfig, row []string) error {
	return AppendCSVRows(path, cfg, [][]string{row})
}

// AppendCSVRows adds rows to path and creates the file when it is missing.
func AppendCSVRows(path string, cfg CSVConfig, rows [][]string) error {
	if err := EnsureParentDir(path); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open csv file append mode: %w", err)
	}

	defer func() {
		_ = f.Close()
	}()

	bw := bufio.NewWriterSize(f, clampBufferSize(cfg.BufferSize))
	w := csv.NewWriter(bw)
	applyWriterConfig(w, cfg)

	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return fmt.Errorf("append csv row error: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}

	return bw.Flush()
}

func clampBufferSize(n int) int {
	if n <= 0 {
		return defaultBufferSize
	}
	return min(maxBufferSize, max(minBufferSize, n))
}

// NewBufferedWriter wraps w in a 64KB buffer. That is enough that writing one
// CSV row at a time does not turn into one syscall per row.
func NewBufferedWriter(w io.Writer) *bufio.Writer {
	return bufio.NewWriterSize(w, defaultBufferSize)
}

// NewCSVWriter wraps w in a CSV writer using cfg.
func NewCSVWriter(w io.Writer, cfg CSVConfig) *csv.Writer {
	cw := csv.NewWriter(w)
	applyWriterConfig(cw, cfg)
	return cw
}
