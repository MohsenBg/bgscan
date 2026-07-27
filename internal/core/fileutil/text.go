// Package fileutil provides helpers for reading, writing, and processing files.
package fileutil

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"bgscan/internal/logger"
)

// TextStreamConfig configures tokenization and buffer limits for StreamTextFile.
type TextStreamConfig struct {
	SplitFunc  bufio.SplitFunc
	BufferSize int
	MaxToken   int
}

// WriteTextFile overwrites path with content, creating parent directories as needed.
func WriteTextFile(path string, content string) error {
	if err := EnsureDir(path); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write text file: %w", err)
	}
	return nil
}

// WriteTextFileIfNotExist writes content only when path does not exist.
func WriteTextFileIfNotExist(path string, content string) error {
	if CheckFileExists(path) {
		return nil
	}
	return WriteTextFile(path, content)
}

// GetTextFile reads and returns the contents of path.
func GetTextFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read text file: %w", err)
	}
	return string(data), nil
}

// AppendTextFile appends content to path, creating parent directories as needed.
func AppendTextFile(path string, content string) error {
	if err := EnsureDir(path); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open text file append mode: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger.CoreError("error closing file: %v", err)
		}
	}()

	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("append text file: %w", err)
	}
	return nil
}

// StreamTextFile calls handler for each token produced by the configured scanner.
func StreamTextFile(ctx context.Context, path string, cfg TextStreamConfig, handler func(string) error) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open text file stream: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger.CoreError("error closing file: %v", err)
		}
	}()

	scanner := bufio.NewScanner(f)

	if cfg.SplitFunc != nil {
		scanner.Split(cfg.SplitFunc)
	}
	bufSize := cfg.BufferSize
	if bufSize <= 0 {
		bufSize = 64 * 1024
	}
	maxToken := cfg.MaxToken
	if maxToken <= 0 {
		maxToken = bufSize
	}
	scanner.Buffer(make([]byte, bufSize), maxToken)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := handler(scanner.Text()); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("text stream error: %w", err)
	}
	return nil
}

// StreamTextToChan sends each scanned token to out.
func StreamTextToChan(ctx context.Context, path string, cfg TextStreamConfig, out chan<- string) error {
	return StreamTextFile(ctx, path, cfg, func(token string) error {
		select {
		case out <- token:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
}

// CopyFile copies src to dst, creating parent directories as needed.
func CopyFile(src, dst string) error {
	if err := EnsureDir(dst); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer func() {
		if err := in.Close(); err != nil {
			logger.CoreError("error closing file: %v", err)
		}
	}()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(dst)
		return fmt.Errorf("copy pipeline stream failure: %w", err)
	}

	if err := out.Sync(); err != nil {
		_ = out.Close()
		return fmt.Errorf("disk sync flush failed: %w", err)
	}

	return out.Close()
}
