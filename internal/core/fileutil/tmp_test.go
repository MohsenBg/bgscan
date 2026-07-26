package fileutil

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateTmpFile_CreatesFile(t *testing.T) {
	f, path, err := CreateTmpFile("test-*")
	if err != nil {
		t.Fatalf(
			"CreateTmpFile failed: %v",
			err,
		)
	}

	defer func() {
		_ = f.Close()
		_ = os.Remove(path)
	}()

	if f == nil {
		t.Fatal("expected file handle")
	}

	if path == "" {
		t.Fatal("expected path")
	}

	if !filepath.IsAbs(path) {
		t.Fatalf(
			"path should be absolute: %s",
			path,
		)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf(
			"stat failed: %v",
			err,
		)
	}

	if info.IsDir() {
		t.Fatal(
			"expected file, got directory",
		)
	}
}

func TestCreateTmpFile_UsesPattern(t *testing.T) {
	f, path, err := CreateTmpFile("bgscan-test-*")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = f.Close()
		_ = os.Remove(path)
	}()

	name := filepath.Base(path)

	if !strings.HasPrefix(
		name,
		"bgscan-test-",
	) {
		t.Fatalf(
			"unexpected temp name: %s",
			name,
		)
	}
}

func TestCreateTmpFile_ReturnsWritableFile(t *testing.T) {
	f, path, err := CreateTmpFile("write-*")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = f.Close()
		_ = os.Remove(path)
	}()

	content := []byte("hello")

	_, err = f.Write(content)
	if err != nil {
		t.Fatalf(
			"temp file is not writable: %v",
			err,
		)
	}
}

func TestCreateTmpJSONFile_WritesJSON(t *testing.T) {
	data := map[string]any{
		"name":    "bgscan",
		"version": 1,
	}

	path, err := CreateTmpJSONFile(
		"json-*",
		data,
	)
	if err != nil {
		t.Fatalf(
			"CreateTmpJSONFile failed: %v",
			err,
		)
	}

	defer func() {
		_ = os.Remove(path)
	}()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any

	err = json.Unmarshal(
		content,
		&got,
	)
	if err != nil {
		t.Fatalf(
			"invalid json: %v",
			err,
		)
	}

	if got["name"] != "bgscan" {
		t.Fatalf(
			"unexpected json content",
		)
	}
}

func TestCreateTmpJSONFile_ReturnsAbsolutePath(t *testing.T) {
	path, err := CreateTmpJSONFile(
		"absolute-*",
		map[string]string{
			"a": "b",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = os.Remove(path)
	}()

	if !filepath.IsAbs(path) {
		t.Fatalf(
			"path should be absolute: %s",
			path,
		)
	}
}

func TestCreateTmpJSONFile_CreatesReadableClosedFile(t *testing.T) {
	path, err := CreateTmpJSONFile(
		"closed-*",
		[]string{
			"a",
			"b",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = os.Remove(path)
	}()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf(
			"file should be readable after creation: %v",
			err,
		)
	}

	if len(data) == 0 {
		t.Fatal(
			"expected json content",
		)
	}
}

func TestCreateTmpJSONFile_InvalidJSONValue(t *testing.T) {
	// channels cannot be JSON encoded
	invalid := make(chan int)

	path, err := CreateTmpJSONFile(
		"invalid-*",
		invalid,
	)

	if err == nil {
		if path != "" {
			_ = os.Remove(path)
		}

		t.Fatal(
			"expected json encoding error",
		)
	}

	if !errors.Is(err, &json.UnmarshalTypeError{}) {
		t.Logf(
			"received expected error: %v",
			err,
		)
	}

	if path != "" {
		if _, statErr := os.Stat(path); statErr == nil {
			t.Fatalf(
				"temporary file should be removed after encoding failure",
			)
		}
	}
}
