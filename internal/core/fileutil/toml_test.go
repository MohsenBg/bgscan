package fileutil

import (
	"errors"
	"path/filepath"
	"testing"
)

type testConfig struct {
	Name    string `toml:"name"`
	Enabled bool   `toml:"enabled"`
	Count   int    `toml:"count"`
}

func TestWriteTOMLFile_WritesAndCreatesDirectory(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"nested",
		"config.toml",
	)

	input := testConfig{
		Name:    "bgscan",
		Enabled: true,
		Count:   10,
	}

	err := WriteTOMLFile(path, input)
	if err != nil {
		t.Fatalf(
			"WriteTOMLFile failed: %v",
			err,
		)
	}

	var got testConfig

	err = GetTOMLFile(path, &got)
	if err != nil {
		t.Fatal(err)
	}

	if got != input {
		t.Fatalf(
			"unexpected config\nwant=%+v\ngot=%+v",
			input,
			got,
		)
	}
}

func TestWriteTOMLFile_OverwritesExistingFile(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"config.toml",
	)

	err := WriteTOMLFile(
		path,
		testConfig{
			Name: "old",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = WriteTOMLFile(
		path,
		testConfig{
			Name: "new",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	var got testConfig

	err = GetTOMLFile(
		path,
		&got,
	)
	if err != nil {
		t.Fatal(err)
	}

	if got.Name != "new" {
		t.Fatalf(
			"expected overwrite, got %q",
			got.Name,
		)
	}
}

func TestWriteTOMLFileIfNotExist_CreatesMissingFile(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"config.toml",
	)

	err := WriteTOMLFileIfNotExist(
		path,
		testConfig{
			Name: "created",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	var got testConfig

	err = GetTOMLFile(
		path,
		&got,
	)
	if err != nil {
		t.Fatal(err)
	}

	if got.Name != "created" {
		t.Fatalf(
			"unexpected value",
		)
	}
}

func TestWriteTOMLFileIfNotExist_DoesNotOverwrite(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"config.toml",
	)

	err := WriteTOMLFile(
		path,
		testConfig{
			Name: "original",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = WriteTOMLFileIfNotExist(
		path,
		testConfig{
			Name: "changed",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	var got testConfig

	err = GetTOMLFile(
		path,
		&got,
	)
	if err != nil {
		t.Fatal(err)
	}

	if got.Name != "original" {
		t.Fatalf(
			"file was overwritten",
		)
	}
}

func TestGetTOMLFile_ReturnsErrorForMissingFile(t *testing.T) {
	var config testConfig

	err := GetTOMLFile(
		filepath.Join(
			t.TempDir(),
			"missing.toml",
		),
		&config,
	)

	if err == nil {
		t.Fatal(
			"expected error",
		)
	}
}

func TestGetTOMLFile_InvalidTOML(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"broken.toml",
	)

	err := WriteTextFile(
		path,
		"invalid = [",
	)
	if err != nil {
		t.Fatal(err)
	}

	var config testConfig

	err = GetTOMLFile(
		path,
		&config,
	)

	if err == nil {
		t.Fatal(
			"expected TOML decode error",
		)
	}
}

func TestGetTOMLFileOrDefault_UsesExistingValue(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"config.toml",
	)

	err := WriteTOMLFile(
		path,
		testConfig{
			Name: "existing",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	var got testConfig

	err = GetTOMLFileOrDefault(
		path,
		&got,
		testConfig{
			Name: "default",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if got.Name != "existing" {
		t.Fatalf(
			"expected existing value",
		)
	}
}

func TestGetTOMLFileOrDefault_CreatesDefaultWhenMissing(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"config.toml",
	)

	var got testConfig

	defaultValue := testConfig{
		Name:  "default",
		Count: 5,
	}

	err := GetTOMLFileOrDefault(
		path,
		&got,
		defaultValue,
	)
	if err != nil {
		t.Fatal(err)
	}

	if got != defaultValue {
		t.Fatalf(
			"default was not applied",
		)
	}

	var stored testConfig

	err = GetTOMLFile(
		path,
		&stored,
	)
	if err != nil {
		t.Fatal(err)
	}

	if stored != defaultValue {
		t.Fatalf(
			"default was not persisted",
		)
	}
}

func TestUpdateTOMLFile_UpdatesValue(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"config.toml",
	)

	err := WriteTOMLFile(
		path,
		testConfig{
			Count: 1,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	var config testConfig

	err = UpdateTOMLFile(
		path,
		&config,
		func(value any) error {
			cfg := value.(*testConfig)

			cfg.Count++

			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	var result testConfig

	err = GetTOMLFile(
		path,
		&result,
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Count != 2 {
		t.Fatalf(
			"expected count 2 got %d",
			result.Count,
		)
	}
}

func TestUpdateTOMLFile_ReturnsUpdateError(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"config.toml",
	)

	err := WriteTOMLFile(
		path,
		testConfig{},
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := errors.New("update failed")

	var config testConfig

	err = UpdateTOMLFile(
		path,
		&config,
		func(any) error {
			return expected
		},
	)

	if err == nil {
		t.Fatal(
			"expected error",
		)
	}
}
