package iplist

import "github.com/MohsenBg/bgscan/internal/core/fileutil"

// WriteCSV writes a stream of IPList entries to a CSV file.
// The provided callback receives a writer function that writes a single row.
func WriteCSV(path string, fn func(func(IPList) error) error) error {
	return fileutil.StreamWriteCSV(path, DefaultCSVConfig, func(write func([]string) error) error {
		return fn(func(item IPList) error {
			return write(item.EncodeCSV())
		})
	})
}

// ReadCSV streams rows along with their raw disk byte offset position.
func ReadCSV(path string, fn func(IPList, int64) error) error {
	return fileutil.StreamCSVIndexed(path, DefaultCSVConfig, func(rec []string, offset int64) error {
		entry, ok := ParseRecord(rec)
		if !ok {
			return nil
		}
		return fn(entry, offset)
	})
}
