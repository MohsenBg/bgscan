package fileutil

import (
	"bufio"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWriteTextFile_WritesContentAndCreatesDirectory(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"nested",
		"file.txt",
	)

	content := "hello world"

	err := WriteTextFile(path, content)
	if err != nil {
		t.Fatalf("WriteTextFile failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != content {
		t.Fatalf(
			"unexpected content\nwant=%q\ngot=%q",
			content,
			string(got),
		)
	}
}

func TestWriteTextFile_OverwritesExistingFile(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"file.txt",
	)

	err := WriteTextFile(path, "old")
	if err != nil {
		t.Fatal(err)
	}

	err = WriteTextFile(path, "new")
	if err != nil {
		t.Fatal(err)
	}

	content, err := GetTextFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if content != "new" {
		t.Fatalf(
			"expected overwrite got %q",
			content,
		)
	}
}

func TestGetTextFile_ReturnsContent(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"file.txt",
	)

	err := os.WriteFile(
		path,
		[]byte("test"),
		0o644,
	)
	if err != nil {
		t.Fatal(err)
	}

	got, err := GetTextFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if got != "test" {
		t.Fatalf(
			"unexpected content %q",
			got,
		)
	}
}

func TestGetTextFile_MissingFile(t *testing.T) {
	_, err := GetTextFile(
		filepath.Join(
			t.TempDir(),
			"missing.txt",
		),
	)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWriteTextFileIfNotExist_CreatesFile(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"file.txt",
	)

	err := WriteTextFileIfNotExist(
		path,
		"content",
	)
	if err != nil {
		t.Fatal(err)
	}

	got, err := GetTextFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if got != "content" {
		t.Fatalf(
			"unexpected content",
		)
	}
}

func TestWriteTextFileIfNotExist_DoesNotOverwrite(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"file.txt",
	)

	err := WriteTextFile(path, "original")
	if err != nil {
		t.Fatal(err)
	}

	err = WriteTextFileIfNotExist(
		path,
		"changed",
	)
	if err != nil {
		t.Fatal(err)
	}

	got, err := GetTextFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if got != "original" {
		t.Fatalf(
			"file was overwritten: %q",
			got,
		)
	}
}

func TestAppendTextFile_AppendsContent(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"append.txt",
	)

	err := AppendTextFile(path, "one")
	if err != nil {
		t.Fatal(err)
	}

	err = AppendTextFile(path, "two")
	if err != nil {
		t.Fatal(err)
	}

	got, err := GetTextFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if got != "onetwo" {
		t.Fatalf(
			"unexpected content %q",
			got,
		)
	}
}

func TestStreamTextFile_ReadsLines(t *testing.T) {
	path := createTextFile(
		t,
		"one\ntwo\nthree\n",
	)

	var got []string

	err := StreamTextFile(
		context.Background(),
		path,
		TextStreamConfig{},
		func(value string) error {
			got = append(got, value)
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"one",
		"two",
		"three",
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf(
			"unexpected tokens\nwant=%v\ngot=%v",
			expected,
			got,
		)
	}
}

func TestStreamTextFile_HandlerError(t *testing.T) {
	path := createTextFile(
		t,
		"one\n",
	)

	expected := errors.New("handler failed")

	err := StreamTextFile(
		context.Background(),
		path,
		TextStreamConfig{},
		func(string) error {
			return expected
		},
	)

	if !errors.Is(err, expected) {
		t.Fatalf(
			"expected handler error got %v",
			err,
		)
	}
}

func TestStreamTextFile_ContextCancellation(t *testing.T) {
	path := createTextFile(
		t,
		"one\ntwo\n",
	)

	ctx, cancel := context.WithCancel(context.Background())

	called := false

	err := StreamTextFile(
		ctx,
		path,
		TextStreamConfig{},
		func(string) error {
			if !called {
				called = true
				cancel()
			}
			return nil
		},
	)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"expected cancellation error got %v",
			err,
		)
	}
}

func TestStreamTextToChan(t *testing.T) {
	path := createTextFile(
		t,
		"a\nb\nc\n",
	)

	ch := make(chan string, 3)

	err := StreamTextToChan(
		context.Background(),
		path,
		TextStreamConfig{},
		ch,
	)
	if err != nil {
		t.Fatal(err)
	}

	close(ch)

	var got []string

	for value := range ch {
		got = append(got, value)
	}

	expected := []string{
		"a",
		"b",
		"c",
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf(
			"unexpected values",
		)
	}
}

func TestCopyFile_CopiesContent(t *testing.T) {
	dir := t.TempDir()

	src := filepath.Join(dir, "source.txt")
	dst := filepath.Join(dir, "nested", "dest.txt")

	content := "copy me"

	err := os.WriteFile(
		src,
		[]byte(content),
		0o644,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = CopyFile(src, dst)
	if err != nil {
		t.Fatal(err)
	}

	got, err := GetTextFile(dst)
	if err != nil {
		t.Fatal(err)
	}

	if got != content {
		t.Fatalf(
			"unexpected copied content",
		)
	}
}

func TestCopyFile_OverwritesDestination(t *testing.T) {
	dir := t.TempDir()

	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	if err := os.WriteFile(src, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(dst, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := CopyFile(src, dst)
	if err != nil {
		t.Fatal(err)
	}

	content, err := GetTextFile(dst)
	if err != nil {
		t.Fatal(err)
	}

	if content != "new" {
		t.Fatalf(
			"destination was not replaced",
		)
	}
}

func createTextFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(
		t.TempDir(),
		"test.txt",
	)

	if err := os.WriteFile(
		path,
		[]byte(content),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestStreamTextFile_CustomSplitFunc(t *testing.T) {
	path := createTextFile(
		t,
		"hello,world",
	)

	var got []string

	err := StreamTextFile(
		context.Background(),
		path,
		TextStreamConfig{
			SplitFunc: bufio.ScanWords,
		},
		func(value string) error {
			got = append(got, value)
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"hello,world",
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf(
			"unexpected values",
		)
	}
}
