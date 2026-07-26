package fileutil

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestListFiles_ReturnsFilesInDirectory(t *testing.T) {
	dir := t.TempDir()

	createFile(t, filepath.Join(dir, "one.txt"))
	createFile(t, filepath.Join(dir, "two.txt"))

	files, err := ListFiles(dir, nil)
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf(
			"expected 2 files, got %d",
			len(files),
		)
	}

	names := fileNames(files)

	sort.Strings(names)

	expected := []string{
		"one.txt",
		"two.txt",
	}

	for i := range expected {
		if names[i] != expected[i] {
			t.Fatalf(
				"unexpected file names\nwant=%v\ngot=%v",
				expected,
				names,
			)
		}
	}
}

func TestListFiles_ReturnsAbsolutePaths(t *testing.T) {
	dir := t.TempDir()

	createFile(t, filepath.Join(dir, "test.txt"))

	files, err := ListFiles(dir, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Fatalf("expected one file")
	}

	if !filepath.IsAbs(files[0].Path) {
		t.Fatalf(
			"path should be absolute: %s",
			files[0].Path,
		)
	}
}

func TestListFiles_SkipsDirectories(t *testing.T) {
	dir := t.TempDir()

	createFile(t, filepath.Join(dir, "file.txt"))

	err := os.Mkdir(
		filepath.Join(dir, "subdir"),
		0o755,
	)
	if err != nil {
		t.Fatal(err)
	}

	files, err := ListFiles(dir, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Fatalf(
			"expected only files, got %d entries",
			len(files),
		)
	}

	if files[0].Name != "file.txt" {
		t.Fatalf(
			"unexpected file: %s",
			files[0].Name,
		)
	}
}

func TestListFiles_Filter(t *testing.T) {
	dir := t.TempDir()

	createFile(t, filepath.Join(dir, "keep.txt"))
	createFile(t, filepath.Join(dir, "skip.log"))

	files, err := ListFiles(
		dir,
		func(name string, info os.FileInfo) bool {
			return filepath.Ext(name) == ".txt"
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 {
		t.Fatalf(
			"expected one filtered file, got %d",
			len(files),
		)
	}

	if files[0].Name != "keep.txt" {
		t.Fatalf(
			"unexpected file: %s",
			files[0].Name,
		)
	}
}

func TestListFiles_NilFilterReturnsEverything(t *testing.T) {
	dir := t.TempDir()

	createFile(t, filepath.Join(dir, "a"))
	createFile(t, filepath.Join(dir, "b"))

	files, err := ListFiles(dir, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 2 {
		t.Fatalf(
			"expected 2 files got %d",
			len(files),
		)
	}
}

func TestListFiles_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	files, err := ListFiles(dir, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 0 {
		t.Fatalf(
			"expected no files got %d",
			len(files),
		)
	}
}

func TestListFiles_ReturnsErrorForMissingDirectory(t *testing.T) {
	dir := filepath.Join(
		t.TempDir(),
		"does-not-exist",
	)

	files, err := ListFiles(dir, nil)

	if err == nil {
		t.Fatal("expected error")
	}

	if files != nil {
		t.Fatalf(
			"expected nil files on error",
		)
	}
}

func TestListFiles_FilterReceivesFileInfo(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "test.txt")

	createFile(t, path)

	called := false

	_, err := ListFiles(
		dir,
		func(name string, info os.FileInfo) bool {
			called = true

			if name != "test.txt" {
				t.Fatalf(
					"unexpected name %s",
					name,
				)
			}

			if info.Name() != "test.txt" {
				t.Fatalf(
					"unexpected info name",
				)
			}

			return true
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if !called {
		t.Fatal(
			"filter was not called",
		)
	}
}

func createFile(t *testing.T, path string) {
	t.Helper()

	err := os.WriteFile(
		path,
		[]byte("test"),
		0o644,
	)
	if err != nil {
		t.Fatalf(
			"create file failed: %v",
			err,
		)
	}
}

func fileNames(entries []FileEntry) []string {
	names := make([]string, 0, len(entries))

	for _, entry := range entries {
		names = append(names, entry.Name)
	}

	return names
}
