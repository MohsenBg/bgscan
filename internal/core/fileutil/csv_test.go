package fileutil

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWriteCSVFile_And_StreamCSV(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "test.csv")

	expected := [][]string{
		{"ip", "port"},
		{"1.1.1.1", "443"},
		{"2.2.2.2", "80"},
	}

	err := WriteCSVFile(path, CSVConfig{}, expected)
	if err != nil {
		t.Fatalf("WriteCSVFile failed: %v", err)
	}

	var got [][]string

	err = StreamCSV(path, CSVConfig{}, func(row []string) error {
		got = append(got, row)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamCSV failed: %v", err)
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("records mismatch\nwant=%v\ngot=%v", expected, got)
	}
}

func TestStreamCSV_SkipsHeader(t *testing.T) {
	path := createTempCSV(t, "name,age\nbob,20\nalice,30\n")

	var rows [][]string

	err := StreamCSV(path, CSVConfig{
		HasHeader: true,
	}, func(row []string) error {
		rows = append(rows, row)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamCSV failed: %v", err)
	}

	expected := [][]string{
		{"bob", "20"},
		{"alice", "30"},
	}

	if !reflect.DeepEqual(rows, expected) {
		t.Fatalf("unexpected rows\nwant=%v\ngot=%v", expected, rows)
	}
}

func TestStreamCSV_HandlerError(t *testing.T) {
	path := createTempCSV(t, "a,b\n1,2\n")

	expectedErr := errors.New("stop")

	err := StreamCSV(path, CSVConfig{}, func([]string) error {
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected handler error, got %v", err)
	}
}

func TestStreamCSV_InvalidFile(t *testing.T) {
	err := StreamCSV(
		"does-not-exist.csv",
		CSVConfig{},
		func([]string) error {
			return nil
		},
	)

	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestWriteCSVFile_CustomDelimiter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")

	rows := [][]string{
		{"name", "value"},
		{"a", "b"},
	}

	err := WriteCSVFile(path, CSVConfig{
		Comma: ';',
	}, rows)
	if err != nil {
		t.Fatalf("WriteCSVFile failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	expected := "name;value\na;b\n"

	if string(data) != expected {
		t.Fatalf(
			"unexpected content\nwant=%q\ngot=%q",
			expected,
			string(data),
		)
	}
}

func TestAppendCSVRows_AppendsData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "append.csv")

	err := WriteCSVFile(
		path,
		CSVConfig{},
		[][]string{{"first"}},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = AppendCSVRows(
		path,
		CSVConfig{},
		[][]string{
			{"second"},
			{"third"},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	var rows [][]string

	err = StreamCSV(path, CSVConfig{}, func(row []string) error {
		rows = append(rows, row)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := [][]string{
		{"first"},
		{"second"},
		{"third"},
	}

	if !reflect.DeepEqual(rows, expected) {
		t.Fatalf(
			"unexpected rows\nwant=%v\ngot=%v",
			expected,
			rows,
		)
	}
}

func TestStreamWriteCSV_WritesUsingCallback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stream.csv")

	err := StreamWriteCSV(path, CSVConfig{}, func(write func([]string) error) error {
		if err := write([]string{"hello", "world"}); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		t.Fatalf("StreamWriteCSV failed: %v", err)
	}

	var rows [][]string

	err = StreamCSV(path, CSVConfig{}, func(row []string) error {
		rows = append(rows, row)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := [][]string{
		{"hello", "world"},
	}

	if !reflect.DeepEqual(rows, expected) {
		t.Fatalf("unexpected rows")
	}
}

func TestStreamCSVIndexed_ReturnsCorrectOffsets(t *testing.T) {
	path := createTempCSV(
		t,
		"first,row\nsecond,row\nthird,row\n",
	)

	type result struct {
		row    []string
		offset int64
	}

	var results []result

	err := StreamCSVIndexed(
		path,
		CSVConfig{},
		func(row []string, offset int64) error {
			results = append(results, result{
				row:    row,
				offset: offset,
			})
			return nil
		},
	)
	if err != nil {
		t.Fatalf("StreamCSVIndexed failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 records, got %d", len(results))
	}

	if results[0].offset != 0 {
		t.Fatalf("first offset should be 0 got %d", results[0].offset)
	}

	if results[1].offset <= results[0].offset {
		t.Fatal("offsets should increase")
	}

	if results[2].offset <= results[1].offset {
		t.Fatal("offsets should increase")
	}
}

func TestStreamCSVIndexed_SkipsCommentsAndBlankLines(t *testing.T) {
	path := createTempCSV(
		t,
		"# comment\n\none,two\n",
	)

	var rows [][]string

	err := StreamCSVIndexed(
		path,
		CSVConfig{},
		func(row []string, _ int64) error {
			rows = append(rows, row)
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := [][]string{
		{"one", "two"},
	}

	if !reflect.DeepEqual(rows, expected) {
		t.Fatalf("unexpected rows")
	}
}

func TestStreamCSVToChan(t *testing.T) {
	path := createTempCSV(
		t,
		"a\nb\nc\n",
	)

	ch := make(chan []string, 3)

	err := StreamCSVToChan(
		path,
		CSVConfig{},
		ch,
	)
	if err != nil {
		t.Fatal(err)
	}

	close(ch)

	var got [][]string

	for row := range ch {
		got = append(got, row)
	}

	expected := [][]string{
		{"a"},
		{"b"},
		{"c"},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected rows")
	}
}

func createTempCSV(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(
		t.TempDir(),
		"test.csv",
	)

	err := os.WriteFile(
		path,
		[]byte(content),
		0o644,
	)
	if err != nil {
		t.Fatalf("cannot create temp csv: %v", err)
	}

	return path
}
