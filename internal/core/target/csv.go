package target

import "github.com/MohsenBg/bgscan/internal/core/fileutil"

// WriteCSV writes a stream of Entry values to a CSV file.
// The provided callback receives a writer function that writes a single row.
func WriteCSV(path string, fn func(func(Entry) error) error) error {
	return fileutil.StreamWriteCSV(path, DefaultCSVConfig, func(write func([]string) error) error {
		return fn(func(item Entry) error {
			return write(item.EncodeCSV())
		})
	})
}

// ReadCSV streams rows along with their raw disk byte offset position.
func ReadCSV(path string, fn func(Entry, int64) error) error {
	return fileutil.StreamCSVWithOffset(path, DefaultCSVConfig, func(rec []string, offset int64) error {
		entry, err := ParseRecord(rec)
		if err != nil {
			return nil
		}
		return fn(entry, offset)
	})
}

// ReadInputCSV streams a user-supplied import list, accepting the older column
// counts. Stored target lists are read with ReadCSV, which is strict.
func ReadInputCSV(path string, fn func(Entry, int64) error) error {
	return fileutil.StreamCSVWithOffset(path, DefaultCSVConfig, func(rec []string, offset int64) error {
		entry, err := ParseInputRecord(rec)
		if err != nil {
			return nil
		}
		return fn(entry, offset)
	})
}
