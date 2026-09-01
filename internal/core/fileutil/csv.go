package fileutil

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/MohsenBg/bgscan/internal/logger"
)

// CSVConfig controls CSV reader and writer behavior.
type CSVConfig struct {
	HasHeader        bool
	LazyQuotes       bool
	TrimLeadingSpace bool
	Comma            rune
	FieldsPerRecord  int
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

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

// StreamCSV calls handler for each record in path without retaining all records in memory.
func StreamCSV(path string, cfg CSVConfig, handler func([]string) error) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open csv file: %w", err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			logger.CoreError("error closing file: %v", err)
		}
	}()

	r := csv.NewReader(f)
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

// StreamCSVIndexed calls handler with each record and its starting byte offset.
func StreamCSVIndexed(path string, cfg CSVConfig, handler func(record []string, offset int64) error) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open csv file: %w", err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			logger.CoreError("error closing file: %v", err)
		}
	}()

	br := bufio.NewReader(f)
	var currentByteOffset int64 = 0

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

// StreamCSVToChan sends each CSV record to out.
func StreamCSVToChan(path string, cfg CSVConfig, out chan<- []string) error {
	return StreamCSV(path, cfg, func(rec []string) error {
		out <- rec
		return nil
	})
}

// WriteCSVFile overwrites path with records.
func WriteCSVFile(path string, cfg CSVConfig, records [][]string) error {
	if err := ensureDir(path); err != nil {
		return fmt.Errorf("ensure directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create csv file %s : %w", path, err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			logger.CoreError("error closing file: %v", err)
		}
	}()

	w := csv.NewWriter(f)
	applyWriterConfig(w, cfg)

	if err := w.WriteAll(records); err != nil {
		return fmt.Errorf("write csv records: %w", err)
	}

	w.Flush()
	return w.Error()
}

// StreamWriteCSV provides fn with a callback that writes CSV records to path.
func StreamWriteCSV(path string, cfg CSVConfig, fn func(write func([]string) error) error) error {
	if err := ensureDir(path); err != nil {
		return fmt.Errorf("ensure directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create csv file: %w", err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			logger.CoreError("error closing file: %v", err)
		}
	}()

	w := csv.NewWriter(f)
	applyWriterConfig(w, cfg)

	writeFunc := func(rec []string) error {
		return w.Write(rec)
	}

	if err := fn(writeFunc); err != nil {
		return err
	}

	w.Flush()
	return w.Error()
}

// AppendCSVRow appends row to path.
func AppendCSVRow(path string, cfg CSVConfig, row []string) error {
	return AppendCSVRows(path, cfg, [][]string{row})
}

// AppendCSVRows appends rows to path.
func AppendCSVRows(path string, cfg CSVConfig, rows [][]string) error {
	if err := ensureDir(path); err != nil {
		return fmt.Errorf("ensure directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open csv file append mode: %w", err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			logger.CoreError("error closing file: %v", err)
		}
	}()

	w := csv.NewWriter(f)
	applyWriterConfig(w, cfg)

	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return fmt.Errorf("append csv row error: %w", err)
		}
	}

	w.Flush()
	return w.Error()
}
