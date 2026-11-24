package result

// Count returns the number of valid records in the CSV file at path.
// Count skips records that the schema parser rejects.
func Count(path string, schema ResultSchema) (uint64, error) {
	res, err := ReadCSV(nil, path, schema, func(_ Result) error {
		return nil
	})

	return res.Loaded, err
}
