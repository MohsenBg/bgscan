package logger

import (
	"os"
	"path/filepath"
	"strings"
)

// defaultDir resolves <exe dir>/logs (falling back to the working
// directory under `go run` temp binaries).
func defaultDir() (string, error) {
	base, err := basePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, LogDir), nil
}

// tailFile reads the last n lines of a file efficiently (byte-wise scan
// backwards from the end).
func tailFile(path string, n int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	filesize := stat.Size()

	var lines []string
	var currentLine []byte

	for i := filesize - 1; i >= 0; i-- {
		if _, err := file.Seek(i, 0); err != nil {
			return nil, err
		}
		char := make([]byte, 1)
		if _, err := file.Read(char); err != nil {
			return nil, err
		}
		if char[0] == '\n' {
			if len(currentLine) > 0 {
				lines = append([]string{string(reverse(currentLine))}, lines...)
				currentLine = nil
				if len(lines) == n {
					break
				}
			}
		} else {
			currentLine = append(currentLine, char[0])
		}
	}
	if len(lines) < n && len(currentLine) > 0 {
		lines = append([]string{string(reverse(currentLine))}, lines...)
	}

	return lines, nil
}

func reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func basePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	resolved, err := filepath.EvalSymlinks(exe)
	if err == nil {
		exe = resolved
	}

	if isGoRunTempBinary(exe) {
		return os.Getwd()
	}

	return filepath.Dir(exe), nil
}

func isGoRunTempBinary(path string) bool {
	dir := filepath.Dir(path)
	base := filepath.Base(dir)

	return strings.Contains(path, "go-build") ||
		strings.HasPrefix(base, "exe") && strings.Contains(dir, string(os.PathSeparator)+"b0")
}
