package result

import (
	"bgscan/internal/core/fileutil"
	"bgscan/internal/logger"
)

var csvConfig = fileutil.CSVConfig{Comma: ','}

// ReadResult contains summary information from a CSV read operation.
type ReadResult struct {
	Loaded  uint64 // records successfully parsed
	Skipped uint64 // records that failed to parse
}

// ReadCSV reads records and converts them using the schema parser.
func ReadCSV(
	path string,
	schema ResultSchema,
	fn func(Result) error,
) (ReadResult, error) {
	var res ReadResult

	err := fileutil.StreamCSV(path, csvConfig, func(rec []string) error {
		result, err := schema.Parser(rec)
		if err != nil {
			res.Skipped++
			logger.CoreError("failed to parse record: %v", err)
			return nil
		}

		res.Loaded++
		return fn(result)
	})

	return res, err
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
