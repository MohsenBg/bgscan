package fileutil

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// WriteTOMLFile encodes value as TOML and writes it to path.
// It creates the parent directory when needed.
func WriteTOMLFile(path string, value any) error {
	if err := EnsureFileDir(path); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create TOML file %q: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	if err := toml.NewEncoder(file).Encode(value); err != nil {
		return fmt.Errorf("encode TOML file %q: %w", path, err)
	}

	return nil
}

// WriteSecretTOMLFile encodes value as TOML into path, creating it with mode
// 0600 and restricting any pre-existing file to 0600 as well.
//
// Use it for any configuration that carries credentials (private keys,
// passwords). It is deliberately separate from WriteTOMLFile, which writes
// 0644 and must not be used for secrets.
//
// The Chmod is not redundant with the OpenFile mode: that mode applies only
// when the file is created, so it heals 0644 files written by older versions
// rather than leaving their permissions untouched.
func WriteSecretTOMLFile(path string, value any) error {
	if err := EnsureFileDir(path); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create TOML file %q: %w", path, err)
	}

	defer func() {
		_ = file.Close()
	}()

	if err := file.Chmod(0o600); err != nil {
		return err
	}

	return toml.NewEncoder(file).Encode(value)
}

// WriteTOMLFileIfNotExist writes value only when path does not exist.
func WriteTOMLFileIfNotExist(path string, value any) error {
	if CheckFileExists(path) {
		return nil
	}

	return WriteTOMLFile(path, value)
}

// ReadTOMLFile decodes the TOML file at path into a new value of type T.
func ReadTOMLFile[T any](path string) (T, error) {
	var value T

	file, err := os.Open(path)
	if err != nil {
		return value, fmt.Errorf("open TOML file %q: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	if err := toml.NewDecoder(file).Decode(&value); err != nil {
		return value, fmt.Errorf("decode TOML file %q: %w", path, err)
	}

	return value, nil
}

// ReadTOMLFileOrDefault returns the value in path, or defaultValue if it cannot
// be read or decoded. It does not write defaultValue to disk.
func ReadTOMLFileOrDefault[T any](path string, defaultValue T) T {
	value, err := ReadTOMLFile[T](path)
	if err != nil {
		return defaultValue
	}

	return value
}

// UpdateTOMLFile reads a TOML value, passes it to update, then saves the
// returned value.
func UpdateTOMLFile[T any](path string, update func(T) (T, error)) error {
	value, err := ReadTOMLFile[T](path)
	if err != nil {
		return fmt.Errorf("read TOML file %q: %w", path, err)
	}

	value, err = update(value)
	if err != nil {
		return fmt.Errorf("update TOML file %q: %w", path, err)
	}

	if err := WriteTOMLFile(path, value); err != nil {
		return fmt.Errorf("write TOML file %q: %w", path, err)
	}

	return nil
}
