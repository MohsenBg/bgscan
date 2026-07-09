package logger

var debugLogger *Logger

func InitDebug() error {
	var err error
	debugLogger, err = newLogger("debug.log")
	return err
}

func Debug() *Logger {
	return debugLogger
}

func DebugInfo(msg string, args ...any) {
	debugLogger.write(LevelInfo, msg, args...)
}

func DebugError(msg string, args ...any) {
	debugLogger.write(LevelError, msg, args...)
}

func DebugDump(label string, v any) {
	debugLogger.Dump(label, v)
}
