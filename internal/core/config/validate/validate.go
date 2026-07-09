// Package validate provides validation and normalization for all scanner
// configuration sections.
//
// Two behaviors are intentionally separated:
//
//   - Validate*  strict checks that return per-field errors.
//     Used by the UI (OnValidate) and SaveConfig to reject bad values
//     before they reach disk.
//
//   - Normalize* lenient checks that auto-fix bad values to their defaults
//     and return a list of Warnings describing what changed.
//     Used only at TOML load time so the app always starts successfully.
//     After normalizing, the corrected config is written back to disk so
//     the TOML file always reflects what the app is actually running with.
package validate

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"bgscan/internal/core/config"
)

// ============================================================================
// Warning — describes one auto-fix applied during normalization
// ============================================================================

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

// ============================================================================
// Internal helpers — shared by all Validate* and Normalize* functions
// ============================================================================

func checkInt(field string, v, min, max int) error {
	if v < min || v > max {
		return fmt.Errorf("%s must be between %d and %d", field, min, max)
	}
	return nil
}

func checkUint16(field string, v, min, max uint16) error {
	if v < min || v > max {
		return fmt.Errorf("%s must be between %d and %d", field, min, max)
	}
	return nil
}

func checkDuration(field string, v, min, max time.Duration) error {
	const epsilon = time.Millisecond
	if v+epsilon < min || v-epsilon > max {
		return fmt.Errorf("%s must be between %s and %s", field, min, max)
	}
	return nil
}

func checkString(field, v string) error {
	if strings.TrimSpace(v) == "" {
		return fmt.Errorf("%s must not be empty", field)
	}
	return nil
}

func checkEnum(field, v string, allowed []string) error {
	s := strings.ToLower(v)
	if slices.Contains(allowed, s) {
		return nil
	}
	return fmt.Errorf("%s must be one of %s", field, strings.Join(allowed, ", "))
}

func checkStringSlice(field string, v []string) error {
	if len(v) == 0 {
		return fmt.Errorf("%s must not be empty", field)
	}
	return nil
}

// checkEnumOrder verifies that the 'min' value appears before or at the same
// index as the 'max' value in the provided ordered slice.
func checkEnumOrder(minField, maxField, minVal, maxVal string, allowed []string) error {
	minIdx := slices.Index(allowed, minVal)
	maxIdx := slices.Index(allowed, maxVal)

	// If either value is not in the allowed list, we return nil here.
	// The individual checkEnum calls will catch and report the invalid values.
	if minIdx == -1 || maxIdx == -1 {
		return nil
	}

	if minIdx > maxIdx {
		return fmt.Errorf("%s (%s) must be less than or equal to %s (%s)", minField, minVal, maxField, maxVal)
	}

	return nil
}

// illegalFilenameRegex matches characters that are illegal in filenames on most
// operating systems (Windows, Linux, macOS), including ASCII control characters.
var illegalFilenameRegex = regexp.MustCompile(`[\\/:*?"<>|\x00-\x1F]`)

// checkPrefix verifies that the string is safe to use as a filename or filename prefix.
// It rejects empty strings, path separators, reserved characters, and leading/trailing dots/spaces.
func checkPrefix(field, prefix string) error {
	if strings.TrimSpace(prefix) == "" {
		return fmt.Errorf("%s must not be empty", field)
	}

	// Check for illegal filesystem characters
	if illegalFilenameRegex.MatchString(prefix) {
		return fmt.Errorf("%s contains illegal characters for a filename (\\ / : * ? \" < > |)", field)
	}

	// Prevent leading/trailing dots and spaces, which are highly problematic on Windows
	if strings.HasPrefix(prefix, ".") || strings.HasSuffix(prefix, ".") ||
		strings.HasPrefix(prefix, " ") || strings.HasSuffix(prefix, " ") {
		return fmt.Errorf("%s must not start or end with a dot or space", field)
	}

	return nil
}

// checksHost checks if the host is a valid domain, optionally followed by a path and/or port.
// Examples: "google.com", "google.com/path", "example.com:8080/api/v1"
func checkHost(field, host string) error {
	if strings.TrimSpace(host) == "" {
		return fmt.Errorf("%s must not be empty", field)
	}

	// Strip optional scheme if the user accidentally included it
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")

	// Prepend a dummy scheme to allow url.Parse to correctly identify the host and path
	u, err := url.Parse("http://" + host)
	if err != nil {
		return fmt.Errorf("%s is malformed: %v", field, err)
	}

	domain := u.Hostname()
	if !isValidDomain(domain) {
		return fmt.Errorf("%s contains an invalid domain: %s", field, domain)
	}

	return nil
}

// checkSNI checks if the SNI is a strict, valid domain without any paths, ports, or prefixes.
// Examples: "google.com", "example.com"
func checkSNI(field, sni string) error {
	if strings.TrimSpace(sni) == "" {
		return nil
	}

	// Strip optional scheme if the user accidentally included it
	sni = strings.TrimPrefix(sni, "http://")
	sni = strings.TrimPrefix(sni, "https://")

	// SNI must not contain paths or ports
	if strings.Contains(sni, "/") || strings.Contains(sni, ":") {
		return fmt.Errorf("%s must be a strict domain without paths or ports: %s", field, sni)
	}

	if !isValidDomain(sni) {
		return fmt.Errorf("%s must be a valid domain: %s", field, sni)
	}

	return nil
}

// checkStatusCodes validates that all status codes are in the valid HTTP range
func checkStatusCodes(fieldName string, codes []int) error {
	for _, c := range codes {
		if !isValidHTTPStatusCode(c) {
			return fmt.Errorf("%s: invalid status code %d (must be 100-599)", fieldName, c)
		}
	}
	return nil
}

// pubKeyRegex matches a 64-character hexadecimal string (common for Curve25519/Ed25519 keys).
var pubKeyRegex = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

// checkPubKey verifies that the string is a valid 64-character hexadecimal public key.
func checkPubKey(field, key string) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("%s must not be empty", field)
	}

	if !pubKeyRegex.MatchString(key) {
		return fmt.Errorf("%s must be a 64-character hexadecimal string", field)
	}

	return nil
}

// ============================================================================
// Internal helpers for Normalize* functions
// ============================================================================

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

// I use DurationMS
// func fixDuration(field string, v *time.Duration, min, max, def time.Duration, warns *[]Warning) {
// 	if err := checkDuration(field, *v, min, max); err != nil {
// 		old := *v
// 		*v = def
// 		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
// 	}
// }

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

// fixDurationMS is like fixDuration but operates directly on a DurationMS
// field, avoiding the need to convert back and forth in every caller.
func fixDurationMS(field string, v *config.DurationMS, min, max time.Duration, def config.DurationMS, warns *[]Warning) {
	if err := checkDuration(field, v.Duration(), min, max); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

// fixHost auto-corrects an invalid host to the default value.
func fixHost(field string, v *string, def string, warns *[]Warning) {
	if err := checkHost(field, *v); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

// fixSNI auto-corrects an invalid SNI to the default value.
func fixSNI(field string, v *string, def string, warns *[]Warning) {
	if err := checkSNI(field, *v); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

// fixHTTPStatusCodes Code
func fixHTTPStatusCodes(field string, v *[]int, def []int, warns *[]Warning) {
	if err := checkStatusCodes(field, *v); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

// fixEnumOrder verifies that the 'min' value appears before or at the same
// index as the 'max' value in the provided ordered slice. If the order is
// invalid, it resets both fields to their respective defaults.
func fixEnumOrder(minField, maxField string, minVal, maxVal *string, minDef, maxDef string, allowed []string, warns *[]Warning) {
	if err := checkEnumOrder(minField, maxField, *minVal, *maxVal, allowed); err != nil {
		oldMin, oldMax := *minVal, *maxVal

		// Auto-fix by resetting both to their defaults
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

// fixPubKey auto-corrects an invalid public key to the default value.
func fixPubKey(field string, v *string, def string, warns *[]Warning) {
	if err := checkPubKey(field, *v); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

// fixPrefix auto-corrects an invalid prefix to the default value.
func fixPrefix(field string, v *string, def string, warns *[]Warning) {
	if err := checkPrefix(field, *v); err != nil {
		old := *v
		*v = def
		*warns = append(*warns, Warning{field, old, def, err.Error() + " → default"})
	}
}

// ============================================================================
// Domain Helper
// ============================================================================

var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$`)

func isValidDomain(host string) bool {
	if host == "" || len(host) > 253 {
		return false
	}

	if net.ParseIP(host) != nil {
		return true
	}

	// 2. Explicitly allow "localhost"
	if host == "localhost" {
		return true
	}

	// 3. Validate as a domain name using the regex
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

// isValidHTTPStatusCode checks if a status code is a valid HTTP status code (100-599)
func isValidHTTPStatusCode(code int) bool {
	return code >= 100 && code <= 599
}
