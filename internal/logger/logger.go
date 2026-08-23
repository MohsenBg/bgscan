package logger

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

const LogDir = "logs"

func newLogger(name string) (*Logger, error) {
	return newLoggerToDir(name, LogDir)
}

func newLoggerToDir(name string, dir string) (*Logger, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	path := filepath.Join(dir, name)

	writer := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    50,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}

	l := &Logger{
		name:       name,
		fileWriter: writer,
		fileLogger: log.New(writer, "", log.LstdFlags),
		enabled:    true,
	}

	l.write(LevelInfo, "=== Log session started ===")

	return l, nil
}
