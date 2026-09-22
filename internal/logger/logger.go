// Package logger provides file loggers for bgscan.
//
// Design: three file loggers (core.log, ui.log, debug.log) are created once
// in main via NewSet and threaded through the app explicitly — there are no
// package-level globals. Internally each Logger records through log/slog;
// lumberjack handles rotation and a fan-out handler feeds live subscribers
// (the TUI log viewer).
//
// Call style is printf-based (Info("failed: %v", err)) to keep messages
// compact; use With(key, value) to create a scoped child logger that tags
// every line (all children of one file share its level dial).
//
// A nil *Logger is valid and discards everything, so zero-value structs and
// un-wired tests never panic.
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	// LogDir is the subdirectory (next to the executable) holding log files.
	LogDir = "logs"

	sessionStart = "=== Log session started ==="
	sessionEnd   = "=== Log session ended ==="
)

// sink is the shared state of one log file: the rotating writer, the level
// dial, and the live-view subscribers. All children created via With share
// the same sink.
type sink struct {
	file  io.Writer // *lumberjack.Logger, or nil for discarded loggers
	path  string    // file path (for tailing); "" for discarded loggers
	level *slog.LevelVar

	subClosed bool
	subsMu    sync.RWMutex
	subs      []chan string
}

// broadcast fans one formatted line out to every live subscriber without
// blocking (full buffers drop the line, matching the previous behavior).
func (s *sink) broadcast(line string) {
	if s.file == nil {
		return
	}
	s.subsMu.RLock()
	defer s.subsMu.RUnlock()
	for _, ch := range s.subs {
		select {
		case ch <- line:
		default:
		}
	}
}

// raw writes content verbatim (no formatting) and fans it out.
func (s *sink) raw(content string) {
	if s.file == nil {
		return
	}
	_, _ = io.WriteString(s.file, content)
	s.broadcast(content)
}

// Logger writes leveled, formatted lines to one log file.
type Logger struct {
	slog *slog.Logger
	sink *sink
}

// Option configures a Logger created by New/NewSet.
type Option func(*options)

type options struct {
	dir   string
	level slog.Level
}

// WithDir writes the log file into dir instead of <exe dir>/logs.
func WithDir(dir string) Option {
	return func(o *options) { o.dir = dir }
}

// WithLevel sets the minimum severity recorded (default: Info).
func WithLevel(level slog.Level) Option {
	return func(o *options) { o.level = level }
}

// New creates a named log file logger (e.g. New("core.log")).
func New(name string, opts ...Option) (*Logger, error) {
	cfg := options{level: slog.LevelInfo}
	for _, opt := range opts {
		opt(&cfg)
	}
	dir := cfg.dir
	if dir == "" {
		var err error
		if dir, err = defaultDir(); err != nil {
			return nil, fmt.Errorf("get base path: %w", err)
		}
	}
	return newFileLogger(name, dir, cfg.level)
}

// Set is the bundle of the app's three log files.
type Set struct {
	Core  *Logger
	UI    *Logger
	Debug *Logger
}

// NewSet creates core.log, ui.log and debug.log in one directory.
func NewSet(opts ...Option) (Set, error) {
	core, err := New("core.log", opts...)
	if err != nil {
		return Set{}, err
	}
	ui, err := New("ui.log", opts...)
	if err != nil {
		core.Close()
		return Set{}, err
	}
	debug, err := New("debug.log", opts...)
	if err != nil {
		core.Close()
		ui.Close()
		return Set{}, err
	}
	return Set{Core: core, UI: ui, Debug: debug}, nil
}

// DiscardSet returns three loggers that write nothing. Used as a safe
// fallback when file loggers cannot be created.
func DiscardSet() Set {
	return Set{Core: Discard(), UI: Discard(), Debug: Discard()}
}

// Close closes every logger in the set (idempotent per logger).
func (s Set) Close() {
	s.Core.Close()
	s.UI.Close()
	s.Debug.Close()
}

// SetLevel moves the minimum severity dial on every logger in the set
// (each file has its own dial; children of each share it).
func (s Set) SetLevel(level slog.Level) {
	s.Core.SetLevel(level)
	s.UI.SetLevel(level)
	s.Debug.SetLevel(level)
}

// ParseLevel maps a level name ("debug", "info", "warn" or "error",
// case-insensitive) to slog.Level. Unknown names return an error.
func ParseLevel(name string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown log level %q (want debug, info, warn or error)", name)
	}
}

// Discard returns a logger that writes nothing, ever.
func Discard() *Logger {
	s := &sink{level: new(slog.LevelVar)}
	s.level.Set(slog.LevelInfo)
	return &Logger{slog: slog.New(&handler{sink: s}), sink: s}
}

func newFileLogger(name, dir string, level slog.Level) (*Logger, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, name)

	file := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    50,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}

	s := &sink{file: file, path: path, level: new(slog.LevelVar)}
	s.level.Set(level)

	l := &Logger{
		slog: slog.New(&handler{sink: s}),
		sink: s,
	}

	l.Info(sessionStart)

	return l, nil
}

// Info records a normal event. Printf-style formatting is applied when
// args are present.
func (l *Logger) Info(msg string, args ...any) {
	if l == nil || l.sink.level.Level() > slog.LevelInfo {
		return
	}
	l.slog.Info(format(msg, args...))
}

// Debug records a detailed event; hidden unless the dial is at Debug.
func (l *Logger) Debug(msg string, args ...any) {
	if l == nil || l.sink.level.Level() > slog.LevelDebug {
		return
	}
	l.slog.Debug(format(msg, args...))
}

// Warn records a noteworthy but non-fatal condition.
func (l *Logger) Warn(msg string, args ...any) {
	if l == nil || l.sink.level.Level() > slog.LevelWarn {
		return
	}
	l.slog.Warn(format(msg, args...))
}

// Error records a failure.
func (l *Logger) Error(msg string, args ...any) {
	if l == nil || l.sink.level.Level() > slog.LevelError {
		return
	}
	l.slog.Error(format(msg, args...))
}

// Dump writes an indented JSON block for deep inspection. Visible at the
// Info dial and below (i.e. hidden when only Warn/Error pass).
func (l *Logger) Dump(label string, v any) {
	if l == nil || l.sink.level.Level() > slog.LevelInfo {
		return
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		l.Error("DUMP FAILED %s: %v", label, err)
		return
	}
	l.sink.raw(fmt.Sprintf(
		"\n========== DUMP %s ==========\n%s\n========== END DUMP ==========\n",
		label,
		string(data),
	))
}

// With returns a child logger that tags every line with the given
// key/value pairs. Children share the parent's file and level dial.
// Example: log.With("probe", "slipstream").
func (l *Logger) With(args ...any) *Logger {
	if l == nil {
		return nil
	}
	return &Logger{slog: l.slog.With(args...), sink: l.sink}
}

// SetLevel moves the minimum severity dial for this log file (children
// created by With are affected too).
func (l *Logger) SetLevel(level slog.Level) {
	if l == nil {
		return
	}
	l.sink.level.Set(level)
}

// Level reports the current minimum severity dial.
func (l *Logger) Level() slog.Level {
	if l == nil {
		return slog.LevelInfo + 1 // effectively "off" for a nil logger
	}
	return l.sink.level.Level()
}

// Slog exposes the underlying slog.Logger for stdlib interop.
func (l *Logger) Slog() *slog.Logger {
	if l == nil {
		return slog.New(&handler{sink: &sink{}})
	}
	return l.slog
}

// Subscribe returns a channel of live log lines, optionally primed with
// the last tailLast lines of the file. Call Unsubscribe to release it.
func (l *Logger) Subscribe(buffer int, tailLast int) chan string {
	ch := make(chan string, buffer)

	if l == nil || l.sink.file == nil {
		close(ch)
		return ch
	}

	if tailLast > 0 {
		if lines, err := tailFile(l.sink.path, tailLast); err == nil {
			for _, line := range lines {
				select {
				case ch <- line:
				default:
				}
			}
		}
	}

	l.sink.subsMu.Lock()
	if l.sink.subClosed {
		l.sink.subsMu.Unlock()
		close(ch)
		return ch
	}
	l.sink.subs = append(l.sink.subs, ch)
	l.sink.subsMu.Unlock()

	return ch
}

// Unsubscribe removes ch from the subscriber list and closes it.
func (l *Logger) Unsubscribe(ch chan string) {
	if l == nil || l.sink == nil {
		return
	}
	l.sink.subsMu.Lock()
	defer l.sink.subsMu.Unlock()
	for i, sub := range l.sink.subs {
		if sub == ch {
			l.sink.subs[i] = l.sink.subs[len(l.sink.subs)-1]
			l.sink.subs = l.sink.subs[:len(l.sink.subs)-1]
			close(ch)
			return
		}
	}
}

// Close writes a session-end marker and closes subscriber channels.
// Safe to call more than once.
func (l *Logger) Close() {
	if l == nil || l.sink == nil {
		return
	}
	l.Info(sessionEnd)

	l.sink.subsMu.Lock()
	if !l.sink.subClosed {
		l.sink.subClosed = true
		for _, ch := range l.sink.subs {
			close(ch)
		}
		l.sink.subs = nil
	}
	l.sink.subsMu.Unlock()

	if file, ok := l.sink.file.(*lumberjack.Logger); ok {
		_ = file.Close()
	}
	l.sink.file = nil // further writes are discarded
}

// format renders the printf-style call into a plain string.
func format(msg string, args ...any) string {
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}
