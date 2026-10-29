package logview

// LogAppendMsg carries the log lines buffered since the previous message.
type LogAppendMsg struct {
	Lines []string
}
