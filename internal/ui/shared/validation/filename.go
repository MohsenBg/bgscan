package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// invalidCharsRe matches characters that are invalid in filenames
// across common operating systems (Windows, macOS, Linux).
//
// The pattern includes:
//
//	< > : " / \ | ? *
//
// and the null byte (\x00), which is disallowed in all filesystems.
//
// The regex is compiled once at package initialization for performance.
var invalidCharsRe = regexp.MustCompile(`[<>:"/\\|?*\x00]`)

// reservedNames defines Windows‑reserved filenames that cannot be used
// as regular file names.
//
// Windows disallows these names regardless of extension. For example:
//
//	CON
//	CON.txt
//	PRN.log
//
// These names are checked case‑insensitively.
var reservedNames = []string{
	"CON", "PRN", "AUX", "NUL",
	"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
	"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
}

// ValidateFilename verifies that a filename is safe and valid across
// major operating systems.
//
// The validation rules enforce cross‑platform compatibility by checking:
//
//   - filename is not empty
//   - maximum length is 255 characters
//   - invalid filesystem characters are not present
//   - Windows reserved filenames are not used
//   - filename does not end with a dot
//
// The function returns:
//
//	(true, "") if the filename is valid
//	(false, reason) if validation fails
//
// The returned error message is intended for direct display in UI
// components such as input validation prompts.
func ValidateFilename(filename string) error {
	filename = strings.TrimSpace(filename)

	if filename == "" {
		return errors.New("filename cannot be empty")
	}

	if len(filename) > 255 {
		return errors.New("filename too long (max 255 characters)")
	}

	if invalidCharsRe.MatchString(filename) {
		return errors.New(`filename contains invalid characters: < > : " / \ | ? *`)
	}

	upper := strings.ToUpper(filename)
	for _, reserved := range reservedNames {
		if upper == reserved || strings.HasPrefix(upper, reserved+".") {
			msg := fmt.Sprintf("'%s' is a reserved filename", reserved)
			return errors.New(msg)
		}
	}

	if strings.HasSuffix(filename, ".") {
		return errors.New("filename cannot end with a dot")
	}

	return nil
}
