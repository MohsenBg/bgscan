package fileutil

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
)

type TextStreamConfig struct {
	SplitFunc  bufio.SplitFunc
	BufferSize int
	MaxToken   int
}

func WriteTextFile(path string, content string) error {
	if err := EnsureDir(path); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write text file: %w", err)
	}
	return nil
}

func WriteTextFileIfNotExist(path string, content string) error {
	if CheckFileExists(path) {
		return nil
	}
	return WriteTextFile(path, content)
}

func GetTextFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read text file: %w", err)
	}
	return string(data), nil
}

func AppendTextFile(path string, content string) error {
	if err := EnsureDir(path); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open text file append mode: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("append text file: %w", err)
	}
	return nil
}

func StreamTextFile(ctx context.Context, path string, cfg TextStreamConfig, handler func(string) error) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open text file stream: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	// Default configurations handle allocations reactively
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

// StreamTextToChan streams tokens cleanly into an unbuffered or buffered conduit channel thread.
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

func CopyFile(src, dst string) error {
	if err := EnsureDir(dst); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
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
