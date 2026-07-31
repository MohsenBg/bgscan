package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var invalidCharsRe = regexp.MustCompile(`[<>:"/\\|?*\x00]`)

var reservedNames = []string{
	"CON", "PRN", "AUX", "NUL",
	"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
	"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
}

// ValidateFilename checks whether filename is usable across major operating
// systems. It rejects empty input, names longer than 255 characters,
// filesystem-invalid characters, Windows reserved names, and trailing dots.
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
