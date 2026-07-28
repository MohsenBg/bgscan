// Package validate validates scanner configuration and normalizes invalid values.
//
// Two behaviors are intentionally separated:
//
//   - Validate* performs strict checks and returns per-field errors.
//
//   - Normalize* replaces invalid values with defaults and reports the changes.
package validate

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"bgscan/internal/core/config"
)

var (
	// ErrOutOfRange indicates a value outside its allowed range.
	ErrOutOfRange = errors.New("value out of range")
	// ErrEmpty indicates a required value is empty.
	ErrEmpty = errors.New("value must not be empty")
	// ErrInvalidEnum indicates a value is not one of the allowed values.
	ErrInvalidEnum = errors.New("value not in allowed set")
	// ErrEnumOrder indicates a minimum value exceeds its maximum.
	ErrEnumOrder = errors.New("min value exceeds max value")
	// ErrIllegalChars indicates a value contains unsupported filename characters.
	ErrIllegalChars = errors.New("value contains illegal characters")
	// ErrInvalidDomain indicates a domain is malformed.
	ErrInvalidDomain = errors.New("invalid domain")
	// ErrMalformedURL indicates a URL cannot be parsed.
	ErrMalformedURL = errors.New("malformed URL")
	// ErrPathOrPort indicates a domain includes a path or port.
	ErrPathOrPort = errors.New("domain must not contain path or port")
	// ErrInvalidStatusCode indicates a status code is outside the HTTP range.
	ErrInvalidStatusCode = errors.New("invalid HTTP status code")
	// ErrInvalidPubKey indicates a public key is not a 64-character hexadecimal value.
	ErrInvalidPubKey = errors.New("invalid public key")
	// ErrLeadingTrailing indicates a value starts or ends with a dot or space.
	ErrLeadingTrailing = errors.New("value must not start or end with a dot or space")
	// ErrDotDot indicates a value contains two consecutive dots.
	ErrDotDot = errors.New("value cannot contain '..'")
)

// Warning describes a single field that was auto-fixed during normalization.
type Warning struct {
	Field  string
	OldVal any
	NewVal any
	Reason string
}

// String returns a human-readable one-line description of the warning.
func (w Warning) String() string {
	return fmt.Sprintf("%s: %v → %v (%s)", w.Field, w.OldVal, w.NewVal, w.Reason)
}

func checkInt(field string, v, min, max int) error {
	if v < min || v > max {
		return fmt.Errorf("%s must be between %d and %d: %w", field, min, max, ErrOutOfRange)
	}
	return nil
}

func checkUint16(field string, v, min, max uint16) error {
	if v < min || v > max {
		return fmt.Errorf("%s must be between %d and %d: %w", field, min, max, ErrOutOfRange)
	}
	return nil
}

func checkDuration(field string, v, min, max time.Duration) error {
	const epsilon = time.Millisecond
	if v+epsilon < min || v-epsilon > max {
		return fmt.Errorf("%s must be between %s and %s: %w", field, min, max, ErrOutOfRange)
	}
	return nil
}

func checkString(field, v string) error {
	if strings.TrimSpace(v) == "" {
		return fmt.Errorf("%s must not be empty: %w", field, ErrEmpty)
	}
	return nil
}

func checkEnum(field, v string, allowed []string) error {
	s := strings.ToLower(v)
	if slices.Contains(allowed, s) {
		return nil
	}
	return fmt.Errorf("%s must be one of %s: %w", field, strings.Join(allowed, ", "), ErrInvalidEnum)
}

func checkStringSlice(field string, v []string) error {
	if len(v) == 0 {
		return fmt.Errorf("%s must not be empty: %w", field, ErrEmpty)
	}
	return nil
}

func checkEnumOrder(minField, maxField, minVal, maxVal string, allowed []string) error {
	minIdx := slices.Index(allowed, minVal)
	maxIdx := slices.Index(allowed, maxVal)

	// Individual enum checks report values not present in allowed.
	if minIdx == -1 || maxIdx == -1 {
		return nil
	}

	if minIdx > maxIdx {
		return fmt.Errorf("%s (%s) must be less than or equal to %s (%s): %w",
			minField, minVal, maxField, maxVal, ErrEnumOrder)
	}

	return nil
}

var illegalFilenameRegex = regexp.MustCompile(`[\\/:*?"<>|\x00-\x1F]`)

func checkPrefix(field, prefix string) error {
	if strings.TrimSpace(prefix) == "" {
		return fmt.Errorf("%s must not be empty: %w", field, ErrEmpty)
	}

	if illegalFilenameRegex.MatchString(prefix) {
		return fmt.Errorf("%s contains illegal characters for a filename (\\ / : * ? \" < > |): %w",
			field, ErrIllegalChars)
	}

	if strings.HasPrefix(prefix, ".") || strings.HasSuffix(prefix, ".") ||
		strings.HasPrefix(prefix, " ") || strings.HasSuffix(prefix, " ") {
		return fmt.Errorf("%s must not start or end with a dot or space: %w", field, ErrLeadingTrailing)
	}

	return nil
}

func checkDirectoryName(field, name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%s must not be empty: %w", field, ErrEmpty)
	}

	if illegalFilenameRegex.MatchString(name) {
		return fmt.Errorf("%s contains illegal characters for a directory name: %w", field, ErrIllegalChars)
	}

	if name == "." || name == ".." {
		return fmt.Errorf("%s cannot be '.' or '..': %w", field, ErrIllegalChars)
	}

	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") ||
		strings.HasPrefix(name, " ") || strings.HasSuffix(name, " ") {
		return fmt.Errorf("%s must not start or end with a dot or space: %w", field, ErrLeadingTrailing)
	}

	if strings.Contains(name, "..") {
		return fmt.Errorf("%s cannot contain '..': %w", field, ErrDotDot)
	}

	return nil
}

func checkHost(field, host string) error {
	if strings.TrimSpace(host) == "" {
		return fmt.Errorf("%s must not be empty: %w", field, ErrEmpty)
	}

	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")

	u, err := url.Parse("http://" + host)
	if err != nil {
		return fmt.Errorf("%s is malformed: %v: %w", field, err, ErrMalformedURL)
	}

	domain := u.Hostname()
	if !isValidDomain(domain) {
		return fmt.Errorf("%s contains an invalid domain: %s: %w", field, domain, ErrInvalidDomain)
	}

	return nil
}

func checkDomain(field, domain string) error {
	if err := checkString(field, domain); err != nil {
		return err
	}
	if err := checkSNI(field, domain); err != nil {
		return err
	}
	return nil
}

func checkSNI(field, sni string) error {
	if strings.TrimSpace(sni) == "" {
		return nil
	}

	sni = strings.TrimPrefix(sni, "http://")
	sni = strings.TrimPrefix(sni, "https://")

	if strings.Contains(sni, "/") || strings.Contains(sni, ":") {
		return fmt.Errorf("%s must be a strict domain without paths or ports: %s: %w",
			field, sni, ErrPathOrPort)
	}

	if !isValidDomain(sni) {
		return fmt.Errorf("%s must be a valid domain: %s: %w", field, sni, ErrInvalidDomain)
	}

	return nil
}

func checkStatusCodes(fieldName string, codes []int) error {
	for _, c := range codes {
		if !isValidHTTPStatusCode(c) {
			return fmt.Errorf("%s: invalid status code %d (must be 100-599): %w",
				fieldName, c, ErrInvalidStatusCode)
		}
	}
	return nil
}

var pubKeyRegex = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

func checkPubKey(field, key string) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("%s must not be empty: %w", field, ErrEmpty)
	}

	if !pubKeyRegex.MatchString(key) {
		return fmt.Errorf("%s must be a 64-character hexadecimal string: %w", field, ErrInvalidPubKey)
	}

	return nil
}

func fixInt(field string, v *int, min, max, def int, warns *[]Warning) {
	if err := checkInt(field, *v, min, max); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

func fixUint16(field string, v *uint16, min, max, def uint16, warns *[]Warning) {
	if err := checkUint16(field, *v, min, max); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

func fixString(field string, v *string, def string, warns *[]Warning) {
	if err := checkString(field, *v); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

func fixEnum(field string, v *string, allowed []string, def string, warns *[]Warning) {
	if err := checkEnum(field, *v, allowed); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

func fixStringSlice(field string, v *[]string, def []string, warns *[]Warning) {
	if err := checkStringSlice(field, *v); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

func fixDurationMS(field string, v *config.DurationMS, min, max time.Duration, def config.DurationMS, warns *[]Warning) {
	if err := checkDuration(field, v.Duration(), min, max); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

func fixHost(field string, v *string, def string, warns *[]Warning) {
	if err := checkHost(field, *v); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

func fixSNI(field string, v *string, def string, warns *[]Warning) {
	if err := checkSNI(field, *v); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

func fixDomain(field string, v *string, def string, warns *[]Warning) {
	if err := checkDomain(field, *v); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

func fixHTTPStatusCodes(field string, v *[]int, def []int, warns *[]Warning) {
	if err := checkStatusCodes(field, *v); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

func fixEnumOrder(minField, maxField string, minVal, maxVal *string, minDef, maxDef string, allowed []string, warns *[]Warning) {
	if err := checkEnumOrder(minField, maxField, *minVal, *maxVal, allowed); err != nil {
		oldMin, oldMax := *minVal, *maxVal
		*minVal = minDef
		*maxVal = maxDef
		*warns = append(*warns, Warning{
			Field:  fmt.Sprintf("%s and %s", minField, maxField),
			OldVal: fmt.Sprintf("%s, %s", oldMin, oldMax),
			NewVal: fmt.Sprintf("%s, %s", minDef, maxDef),
			Reason: err.Error() + " → defaults",
		})
	}
}

func fixPubKey(field string, v *string, def string, warns *[]Warning) {
	if err := checkPubKey(field, *v); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

func fixPrefix(field string, v *string, def string, warns *[]Warning) {
	if err := checkPrefix(field, *v); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

func fixDirectoryName(field string, v *string, def string, warns *[]Warning) {
	if err := checkDirectoryName(field, *v); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$`)

func isValidDomain(host string) bool {
	if host == "" || len(host) > 253 {
		return false
	}

	if net.ParseIP(host) != nil {
		return true
	}

	if host == "localhost" {
		return true
	}

	if !domainRegex.MatchString(host) {
		return false
	}

	parts := strings.Split(host, ".")
	if len(parts) > 0 {
		tld := parts[len(parts)-1]
		isAllNumeric := true
		for _, c := range tld {
			if c < '0' || c > '9' {
				isAllNumeric = false
				break
			}
		}
		if isAllNumeric {
			return false
		}
	}

	return true
}

func isValidHTTPStatusCode(code int) bool {
	return code >= 100 && code <= 599
}
