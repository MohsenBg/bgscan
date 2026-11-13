package fileutil

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateSystemTempFile_CreatesFile(t *testing.T) {
	f, path, err := CreateSystemTempFile("test-*")
	if err != nil {
		t.Fatalf(
			"CreateSystemTempFile failed: %v",
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

func TestCreateSystemTempFile_UsesPattern(t *testing.T) {
	f, path, err := CreateSystemTempFile("bgscan-test-*")
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

func TestCreateSystemTempFile_ReturnsWritableFile(t *testing.T) {
	f, path, err := CreateSystemTempFile("write-*")
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

func TestCreateSystemTempJSONFile_WritesJSON(t *testing.T) {
	data := map[string]any{
		"name":    "bgscan",
		"version": 1,
	}

	path, err := CreateSystemTempJSONFile(
		"json-*",
		data,
	)
	if err != nil {
		t.Fatalf(
			"CreateSystemTempJSONFile failed: %v",
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

func TestCreateSystemTempJSONFile_ReturnsAbsolutePath(t *testing.T) {
	path, err := CreateSystemTempJSONFile(
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

func TestCreateSystemTempJSONFile_CreatesReadableClosedFile(t *testing.T) {
	path, err := CreateSystemTempJSONFile(
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

func TestCreateSystemTempJSONFile_InvalidJSONValue(t *testing.T) {
	// channels cannot be JSON encoded
	invalid := make(chan int)

	path, err := CreateSystemTempJSONFile(
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

func TestCreateTempIn_UsesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sub")

	f, path, err := CreateTempIn(dir, "bgscan-test-*")
	if err != nil {
		t.Fatalf("CreateTempIn: %v", err)
	}

	t.Cleanup(func() {
		_ = f.Close()
		_ = os.Remove(path)
	})

	if got := filepath.Dir(path); got != dir {
		t.Fatalf("temp dir = %q, want %q", got, dir)
	}
	if !filepath.IsAbs(path) {
		t.Fatalf("temp path should be absolute: %q", path)
	}
}
