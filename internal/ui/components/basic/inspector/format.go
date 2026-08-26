package inspector

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"bgscan/internal/core/speedtest"
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

func FormatDataSpeed(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}

	kpbs, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		logger.UIError("Error while casting type to uint64")
		return ""
	}

	speed := speedtest.BitsPerSec(kpbs * 1000)
	return speed.String()
}

func FormatHex(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}

	m := min(len(s), 15)

	return "0x" + s[:m] + "..."
}

func FormatPublicKey(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}

	s = strings.TrimSpace(s)
	if len(s) <= 10 {
		return s
	}

	return s[:6] + "..." + s[len(s)-4:]
}

// FormatPrivateKey renders an SSH private key as a safe, compact preview.
// The actual key material is never displayed.
func FormatPrivateKey(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}

	s = strings.TrimSpace(s)
	if s == "" {
		return "<empty>"
	}

	keyType := "SSH private key"
	switch {
	case strings.Contains(s, "BEGIN OPENSSH PRIVATE KEY"):
		keyType = "OpenSSH private key"
	case strings.Contains(s, "BEGIN RSA PRIVATE KEY"):
		keyType = "RSA private key"
	case strings.Contains(s, "BEGIN EC PRIVATE KEY"):
		keyType = "EC private key"
	case strings.Contains(s, "BEGIN PRIVATE KEY"):
		keyType = "Private key"
	}

	return "🔑 " + keyType
}

func FormatZeroAsAuto(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}

	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return s
	}

	if f == 0 {
		return "Auto"
	}

	return s
}

func FormatUTLS(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}

	if s == "" {
		return "Native TLS"
	}
	return s
}

func FormatEmptyString(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}

	if s == "" {
		return "<empty>"
	}
	return s
}
