package inspector

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MohsenBg/bgscan/internal/core/speedtest"
)

// FormatInt formats a string that holds an integer with comma separators.
// FormatInt returns an empty string when the value is not a numeric string.
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

// FormatIntOrUnlimited formats a numeric string like FormatInt.
// FormatIntOrUnlimited returns "unlimited" when the value is zero.
func FormatIntOrUnlimited(v any) string {
	if format := FormatInt(v); format == "0" {
		return "unlimited"
	} else {
		return format
	}
}

// FormatDurationMS formats a numeric string that holds milliseconds as a duration.
// FormatDurationMS returns an empty string when the value is not a numeric string.
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

// FormatBool formats a bool as "Active" or "Disabled".
// FormatBool returns an empty string when the value is not a bool.
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

// FormatStringList joins a string list with commas and cuts it at 20 columns.
// FormatStringList adds a count of the hidden items when the list is cut.
// FormatStringList returns "-" for an empty list and an empty string for other types.
func FormatStringList(value any) string {
	const width = 20
	items, ok := value.([]string)
	if !ok {
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

// FormatIntList formats an int list like FormatStringList.
// FormatIntList returns an empty string when the value is not an int list.
func FormatIntList(value any) string {
	items, ok := value.([]int)
	if !ok {
		return ""
	}
	l := make([]string, 0, len(items))
	for _, item := range items {
		l = append(l, strconv.Itoa(item))
	}
	return FormatStringList(l)
}

// FormatDataSpeed formats a numeric string that holds kilobits per second as a data speed.
// FormatDataSpeed returns an empty string when the value is not a numeric string.
func FormatDataSpeed(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}
	kbps, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return ""
	}
	return speedtest.BitsPerSec(kbps * 1000).String()
}

// FormatHex formats a string with a 0x prefix and cuts it after 15 characters.
// FormatHex returns an empty string when the value is not a string.
func FormatHex(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return "0x" + s[:min(len(s), 15)] + "..."
}

// FormatPublicKey shortens a public key string to its first and last characters.
// FormatPublicKey returns short values unchanged and an empty string for other types.
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

// FormatZeroAsAuto returns "Auto" when a numeric string is zero.
// FormatZeroAsAuto returns the input string unchanged in all other cases.
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

// FormatEmptyStringAuto returns "Auto" for an empty string.
// FormatEmptyStringAuto returns an empty string when the value is not a string.
func FormatEmptyStringAuto(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}
	if s == "" {
		return "Auto"
	}
	return s
}

// FormatUTLS returns "Native TLS" for an empty string, which means no uTLS fingerprint is set.
// FormatUTLS returns an empty string when the value is not a string.
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

// FormatEmptyString returns "<empty>" for an empty string and the string itself otherwise.
// FormatEmptyString uses String for fmt.Stringer values and returns an empty string for other types.
func FormatEmptyString(v any) string {
	switch val := v.(type) {
	case string:
		if val == "" {
			return "<empty>"
		}
		return val
	case fmt.Stringer:
		return val.String()
	default:
		return ""
	}
}

// FormatSeconds formats a numeric string that holds seconds as a duration.
// FormatSeconds returns "<empty>" for an empty string and an empty string for bad or negative values.
func FormatSeconds(v any) string {
	s, ok := v.(string)
	if !ok || s == "" {
		return "<empty>"
	}
	seconds, err := strconv.ParseFloat(s, 64)
	if err != nil || seconds < 0 {
		return ""
	}
	d := time.Duration(seconds * float64(time.Second))
	return d.Round(time.Millisecond).String()
}
