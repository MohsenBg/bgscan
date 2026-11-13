package fileutil

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// WriteTOMLFile writes value as TOML, creating the parent directory if it is
// missing.
func WriteTOMLFile(path string, value any) error {
	if err := EnsureParentDir(path); err != nil {
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

// WriteSecretTOMLFile writes TOML that nobody else should be able to read.
// The file ends up with mode 0600 whether it is new or already existed.
//
// Use it for anything carrying credentials: private keys, passwords, tokens.
// WriteTOMLFile writes 0644 on purpose, so keep the two apart.
func WriteSecretTOMLFile(path string, value any) error {
	if err := EnsureParentDir(path); err != nil {
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

// WriteTOMLFileIfNotExist creates path only when it is missing. An existing
// file is left alone.
func WriteTOMLFileIfNotExist(path string, value any) error {
	if FileExists(path) {
		return nil
	}

	return WriteTOMLFile(path, value)
}

// ReadTOMLFile decodes path into a fresh T.
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

// ReadTOMLFileOrDefault returns the decoded value, or defaultValue when the
// file is missing or broken. The default is not written back to disk.
func ReadTOMLFileOrDefault[T any](path string, defaultValue T) T {
	value, err := ReadTOMLFile[T](path)
	if err != nil {
		return defaultValue
	}

	return value
}

// UpdateTOMLFile reads the file, hands the value to update, and writes the
// result back. A failed update leaves the file untouched.
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
