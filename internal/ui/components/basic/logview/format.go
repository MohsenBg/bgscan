package logview

import (
	"regexp"
	"strings"

	"github.com/MohsenBg/bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

var logPattern = regexp.MustCompile(`^(\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2})\s+\[(\w+)\]\s+(.*)$`)

type logParts struct {
	Timestamp string
	Level     string
	Message   string
}

func parseLogLine(line string) *logParts {
	matches := logPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	return &logParts{
		Timestamp: matches[1],
		Level:     matches[2],
		Message:   matches[3],
	}
}

func levelStyle(th *theme.Theme, level string) lipgloss.Style {
	switch strings.ToUpper(level) {
	case "ERROR":
		return lipgloss.NewStyle().Foreground(th.Error).Bold(true)
	case "WARN":
		return lipgloss.NewStyle().Foreground(th.Yellow).Bold(true)
	case "INFO":
		return lipgloss.NewStyle().Foreground(th.Info)
	case "DEBUG":
		return lipgloss.NewStyle().Foreground(th.Orange)
	default:
		return lipgloss.NewStyle().Foreground(th.Text)
	}
}

func timestampStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(th.Timestamp)
}

func messageStyle(th *theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(th.Text)
}

// formatLogLine colorizes a single log line with styled timestamp,
// level, and message. Returns the original line if parsing fails.
func formatLogLine(th *theme.Theme, line string) string {
	parts := parseLogLine(line)
	if parts == nil {
		return line
	}

	ts := timestampStyle(th).Render(parts.Timestamp)
	lvl := levelStyle(th, parts.Level).Render("[" + parts.Level + "]")
	msg := messageStyle(th).Render(parts.Message)

	return ts + " " + lvl + " " + msg
}

// formatLogLines colorizes multiple log lines with spacing between them.
func formatLogLines(th *theme.Theme, lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	var b strings.Builder
	for i, line := range lines {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(formatLogLine(th, line))
	}

	return b.String()
}
