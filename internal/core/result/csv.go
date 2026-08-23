package result

import (
	"bgscan/internal/core/fileutil"
	"bgscan/internal/logger"
)

var csvConfig = fileutil.CSVConfig{Comma: ','}

// ReadCSV reads records and converts them using the schema parser.
func ReadCSV(
	path string,
	schema ResultSchema,
	fn func(Result) error,
) error {
	return fileutil.StreamCSV(path, csvConfig, func(rec []string) error {
		result, err := schema.Parser(rec)
		if err != nil {
			logger.CoreError("failed to parse record: %v", err)
			return nil
		}

		return fn(result)
	})
}

// StreamWriteResults writes results to CSV.
func StreamWriteResults(
	path string,
	fn func(func(Result) error) error,
) error {
	return fileutil.StreamWriteCSV(path, csvConfig, func(write func([]string) error) error {
		return fn(func(r Result) error {
			return write(r.ToRecord())
		})
	})
}
