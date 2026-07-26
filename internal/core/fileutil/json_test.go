package fileutil

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWriteJSONFile_WritesAndCreatesDirectory(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"nested",
		"config.json",
	)

	input := map[string]any{
		"name":    "bgscan",
		"version": 2,
	}

	err := WriteJSONFile(path, input)
	if err != nil {
		t.Fatalf("WriteJSONFile failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected json content")
	}
}

func TestWriteJSONFile_OverwritesExistingFile(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"config.json",
	)

	err := WriteJSONFile(path, map[string]string{
		"value": "old",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = WriteJSONFile(path, map[string]string{
		"value": "new",
	})
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]string

	err = GetJSONFile(path, &result)
	if err != nil {
		t.Fatal(err)
	}

	if result["value"] != "new" {
		t.Fatalf("expected overwrite, got %q", result["value"])
	}
}

func TestGetJSONFile_ReadsJSON(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"data.json",
	)

	expected := struct {
		Name string `json:"name"`
	}{
		Name: "test",
	}

	if err := WriteJSONFile(path, expected); err != nil {
		t.Fatal(err)
	}

	var got struct {
		Name string `json:"name"`
	}

	err := GetJSONFile(path, &got)
	if err != nil {
		t.Fatalf("GetJSONFile failed: %v", err)
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf(
			"unexpected result\nwant=%v\ngot=%v",
			expected,
			got,
		)
	}
}

func TestGetJSONFile_ReturnsErrorWhenMissing(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"missing.json",
	)

	var result map[string]any

	err := GetJSONFile(path, &result)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetJSONFile_ReturnsErrorForInvalidJSON(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"broken.json",
	)

	err := os.WriteFile(
		path,
		[]byte("{invalid json"),
		0o644,
	)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]any

	err = GetJSONFile(path, &result)

	if err == nil {
		t.Fatal("expected invalid json error")
	}
}

func TestWriteJSONFileIfNotExist_CreatesMissingFile(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"config.json",
	)

	err := WriteJSONFileIfNotExist(
		path,
		map[string]string{
			"value": "created",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]string

	if err := GetJSONFile(path, &result); err != nil {
		t.Fatal(err)
	}

	if result["value"] != "created" {
		t.Fatalf("unexpected value")
	}
}

func TestWriteJSONFileIfNotExist_DoesNotOverwrite(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"config.json",
	)

	err := WriteJSONFile(
		path,
		map[string]string{
			"value": "original",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = WriteJSONFileIfNotExist(
		path,
		map[string]string{
			"value": "changed",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]string

	err = GetJSONFile(path, &result)
	if err != nil {
		t.Fatal(err)
	}

	if result["value"] != "original" {
		t.Fatalf(
			"file was overwritten: %q",
			result["value"],
		)
	}
}

func TestGetJSONFileOrDefault_UsesExistingFile(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"config.json",
	)

	err := WriteJSONFile(
		path,
		map[string]string{
			"value": "existing",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	dest := map[string]string{}

	err = GetJSONFileOrDefault(
		path,
		&dest,
		map[string]string{
			"value": "default",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if dest["value"] != "existing" {
		t.Fatalf(
			"expected existing value, got %q",
			dest["value"],
		)
	}
}

func TestGetJSONFileOrDefault_CreatesDefaultWhenMissing(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"config.json",
	)

	dest := map[string]string{}

	defaultValue := map[string]string{
		"value": "default",
	}

	err := GetJSONFileOrDefault(
		path,
		&dest,
		defaultValue,
	)
	if err != nil {
		t.Fatal(err)
	}

	if dest["value"] != "default" {
		t.Fatalf("default was not applied")
	}

	var stored map[string]string

	err = GetJSONFile(path, &stored)
	if err != nil {
		t.Fatal(err)
	}

	if stored["value"] != "default" {
		t.Fatalf("default was not written")
	}
}

func TestUpdateJSONFile_UpdatesData(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"config.json",
	)

	err := WriteJSONFile(
		path,
		map[string]int{
			"count": 1,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	var data map[string]int

	err = UpdateJSONFile(
		path,
		&data,
		func(v any) error {
			cfg := v.(*map[string]int)

			(*cfg)["count"]++

			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]int

	err = GetJSONFile(path, &result)
	if err != nil {
		t.Fatal(err)
	}

	if result["count"] != 2 {
		t.Fatalf(
			"expected count 2 got %d",
			result["count"],
		)
	}
}

func TestUpdateJSONFile_ReturnsUpdateError(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"config.json",
	)

	err := WriteJSONFile(
		path,
		map[string]string{"x": "y"},
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := errors.New("update failed")

	var data map[string]string

	err = UpdateJSONFile(
		path,
		&data,
		func(any) error {
			return expected
		},
	)

	if err == nil {
		t.Fatal("expected error")
	}
}
