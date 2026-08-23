package result

import "io"

// LoadResult streams valid scan results from CSV into a channel.
func LoadResult(path string, schema ResultSchema, out chan<- Result) (ReadResult, error) {
	return ReadCSV(path, schema, func(r Result) error {
		out <- r
		return nil
	})
}

// CountResultKeys counts valid records in streaming mode.
func CountResultKeys(path string, schema ResultSchema) (uint64, error) {
	return Count(path, schema)
}

// LoadAll loads the entire result file into memory.
func LoadAll(path string, schema ResultSchema, maxResults uint32) ([]Result, error) {
	results := make([]Result, 0, 1024)
	var count uint32

	_, err := ReadCSV(path, schema, func(r Result) error {
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
