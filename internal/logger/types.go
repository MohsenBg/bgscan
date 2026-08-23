package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarning
	LevelError
)

func (l LogLevel) String() string {
	return [...]string{"DEBUG", "INFO", "WARN", "ERROR"}[l]
}

type Logger struct {
	name       string
	fileLogger *log.Logger
	fileWriter *lumberjack.Logger
	enabled    bool

	subsMu sync.RWMutex
	subs   []chan string
}

// write a log entry
func (l *Logger) write(level LogLevel, msg string, args ...any) {
	if !l.enabled || l.fileLogger == nil {
		return
	}

	text := fmt.Sprintf(msg, args...)
	line := fmt.Sprintf("[%s] %s", level, text)

	l.fileLogger.Println(line)

	ts := time.Now().Format("2006/01/02 15:04:05")
	fullLine := fmt.Sprintf("%s %s", ts, line)

	l.subsMu.RLock()
	for _, ch := range l.subs {
		select {
		case ch <- fullLine:
		default:
		}
	}
	l.subsMu.RUnlock()
}

// write raw string
func (l *Logger) writeRaw(content string) {
	if !l.enabled || l.fileLogger == nil {
		return
	}

	l.fileLogger.Print(content)

	l.subsMu.RLock()
	for _, ch := range l.subs {
		select {
		case ch <- content:
		default:
		}
	}
	l.subsMu.RUnlock()
}

// Unsubscribe removes a specific channel from the subscriber list.
func (l *Logger) Unsubscribe(ch chan string) {
	l.subsMu.Lock()
	defer l.subsMu.Unlock()

	for i, subCh := range l.subs {
		if subCh == ch {
			l.subs[i] = l.subs[len(l.subs)-1]
			l.subs = l.subs[:len(l.subs)-1]
			close(ch)
			return
		}
	}
}

// subscribe returns a channel for live updates and optionally tails last N lines
func (l *Logger) Subscribe(buffer int, tailLast int) chan string {
	ch := make(chan string, buffer)

	// tail the last N lines
	if tailLast > 0 {
		lines, err := tailFile(l.fileWriter.Filename, tailLast)
		if err == nil {
			for _, line := range lines {
				select {
				case ch <- line:
				default:
				}
			}
		}
	}

	// register subscriber
	l.subsMu.Lock()
	l.subs = append(l.subs, ch)
	l.subsMu.Unlock()

	return ch
}

// Close logger
func (l *Logger) Close() {
	l.write(LevelInfo, "=== Log session ended ===")

	l.subsMu.Lock()
	for _, ch := range l.subs {
		close(ch)
	}
	l.subs = nil
	l.subsMu.Unlock()
}

func (l *Logger) Enable()  { l.enabled = true }
func (l *Logger) Disable() { l.enabled = false }
func (l *Logger) IsEnabled() bool {
	return l.enabled
}

// Dump object to log
func (l *Logger) Dump(label string, v any) {
	if !l.enabled || l.fileLogger == nil {
		return
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		l.write(LevelError, "DUMP FAILED %s: %v", label, err)
		return
	}

	l.writeRaw(fmt.Sprintf(
		"\n========== DUMP %s ==========\n%s\n========== END DUMP ==========\n",
		label,
		string(data),
	))
}

// ----------------- helpers -----------------

// tailFile reads the last N lines efficiently from a file
func tailFile(path string, n int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			CoreError("error closing file: %v", err)
		}
	}()

	stat, _ := file.Stat()
	filesize := stat.Size()

	var lines []string
	var currentLine []byte

	// 1. Move the cursor backward from the end, one byte at a time
	for i := filesize - 1; i >= 0; i-- {
		_, err := file.Seek(i, 0) // Jump to this specific byte
		if err != nil {
			return nil, err
		}

		char := make([]byte, 1)

		_, err = file.Read(char)
		if err != nil {
			return nil, err
		}

		if char[0] == '\n' {
			if len(currentLine) > 0 {
				// Reverse the bytes we collected (because we read backwards)
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

	// Catch the very first line if we didn't hit 'n' lines yet
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
