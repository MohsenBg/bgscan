package form

import (
	"fmt"
	"strings"
)

// FormatValidationErrors formats a map of validation errors into a
// user-friendly string with bullet points and indented messages.
//
// Field names are converted from snake_case to Title Case for display.
func FormatValidationErrors(errs map[string]error) string {
	if len(errs) == 0 {
		return ""
	}

	var b strings.Builder
	for field, e := range errs {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		fmt.Fprintf(&b, "- %s\n  %s", toTitleCase(field), e)
	}
	return b.String()
}

// toTitleCase converts a snake_case string to Title Case.
// Examples: "domain" -> "Domain", "pub_key" -> "Pub Key"
func toTitleCase(s string) string {
	if s == "" {
		return s
	}

	words := strings.Split(s, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}
