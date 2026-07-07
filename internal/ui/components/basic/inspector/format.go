package inspector

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"bgscan/internal/logger"
)

func FormatInt(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		return ""
	}

	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}

	str := strconv.Itoa(n)

	var b strings.Builder
	b.Grow(len(str) + len(str)/3)

	for i, r := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}

	return sign + b.String()
}

func FormatIntOrUnlimited(v any) string {
	format := FormatInt(v)
	if format == "0" {
		return "unlimited"
	}
	return format
}

func FormatDurationMS(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}

	ms, err := strconv.Atoi(s)
	if err != nil {
		return ""
	}

	return (time.Duration(ms) * time.Millisecond).String()
}

func FormatBool(v any) string {
	b, ok := v.(bool)
	if !ok {
		return ""
	}

	if b {
		return "Active"
	}

	return "Disabled"
}

func FormatStringList(value any) string {
	const width = 20

	items, ok := value.([]string)
	if !ok {
		logger.UIError("Error while casting type to []string")
		return ""
	}

	if len(items) == 0 {
		return "-"
	}

	var b strings.Builder

	for i, item := range items {
		part := item
		if i > 0 {
			part = ", " + part
		}

		if b.Len()+len(part) > width {
			fmt.Fprintf(&b, " (+%d)", len(items)-i)
			break
		}

		b.WriteString(part)
	}

	return b.String()
}

func FormatIntList(value any) string {
	items, ok := value.([]int)
	if !ok {
		logger.UIError("Error while casting type to []int")
		return ""
	}

	l := make([]string, 0, len(items))
	for _, item := range items {
		l = append(l, strconv.Itoa(item))
	}

	return FormatStringList(l)
}
