package result

import (
	"github.com/MohsenBg/bgscan/internal/core/fileutil"
	"github.com/MohsenBg/bgscan/internal/logger"
)

var csvConfig = fileutil.CSVConfig{Comma: ','}

// ReadResult contains summary information from a CSV read operation.
type ReadResult struct {
	Loaded  uint64 // records successfully parsed
	Skipped uint64 // records that failed to parse
}

// ReadCSV reads records from path and converts them with the schema parser.
// ReadCSV skips records that fail to parse, counts them in ReadResult.Skipped,
// and passes each valid result to fn. An error from fn stops the read.
func ReadCSV(
	log *logger.Logger,
	path string,
	schema ResultSchema,
	fn func(Result) error,
) (ReadResult, error) {
	var res ReadResult

	err := fileutil.StreamCSV(path, csvConfig, func(rec []string) error {
		result, err := schema.Parser(rec)
		if err != nil {
			res.Skipped++
			log.Error("failed to parse record: %v", err)
			return nil
		}

		res.Loaded++
		return fn(result)
	})

	return res, err
}

// StreamWriteResults writes results to path as CSV.
// StreamWriteResults calls fn with a write function, and fn returns
// results one at a time instead of building a slice first.
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
