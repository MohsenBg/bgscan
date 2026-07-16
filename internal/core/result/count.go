package result

// Count returns number of valid records in CSV file
func Count(path string, schema ResultSchema) (uint64, error) {
	var count uint64

	err := ReadCSV(path, schema, func(_ Result) error {
		count++
		return nil
	})

	return count, err
}
