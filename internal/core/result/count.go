package result

// Count returns number of valid records in CSV file
func Count(path string, schema ResultSchema) (uint64, error) {
	res, err := ReadCSV(nil, path, schema, func(_ Result) error {
		return nil
	})

	return res.Loaded, err
}
