package notice

import "unicode"

// IsNonLatin reports whether typed text contains characters outside ASCII
func IsNonLatin(text string) bool {
	for _, r := range text {
		switch {
		case r <= unicode.MaxASCII:
			continue
		case unicode.Is(unicode.Latin, r):
			continue
		case unicode.Is(unicode.ASCII_Hex_Digit, r):
			continue
		default:
			return true
		}
	}

	return false
}
