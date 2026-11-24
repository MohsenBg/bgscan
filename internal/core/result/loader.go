package result

import (
	"io"

	"github.com/MohsenBg/bgscan/internal/logger"
)

// LoadResult streams valid scan results from the CSV file at path into out.
// The caller owns out and must keep reading from it. Records that fail
// to parse are skipped.
func LoadResult(log *logger.Logger, path string, schema ResultSchema, out chan<- Result) (ReadResult, error) {
	return ReadCSV(log, path, schema, func(r Result) error {
		out <- r
		return nil
	})
}

// CountResultKeys counts the valid records in the file at path.
// CountResultKeys is an alias for Count.
func CountResultKeys(path string, schema ResultSchema) (uint64, error) {
	return Count(path, schema)
}

// LoadAll loads valid results from the file at path into memory.
// LoadAll stops after maxResults results. A maxResults of 0 returns no
// results and no error.
func LoadAll(log *logger.Logger, path string, schema ResultSchema, maxResults uint32) ([]Result, error) {
	results := make([]Result, 0, 1024)
	var count uint32

	_, err := ReadCSV(log, path, schema, func(r Result) error {
		if count >= maxResults {
			return io.EOF
		}

		results = append(results, r)
		count++

		return nil
	})

	if err == io.EOF {
		err = nil
	}

	return results, err
}
