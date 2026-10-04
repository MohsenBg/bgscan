package form

import (
	"fmt"
	"strings"
)

// FormatValidationErrors formats validation errors as a bulleted list.
// Field names are converted from snake_case to Title Case.
func FormatValidationErrors[K ~string](errs map[K]error) string {
	if len(errs) == 0 {
		return ""
	}

	var b strings.Builder
	for field, err := range errs {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}

		fmt.Fprintf(&b, "- %s\n  %s", toTitleCase(string(field)), err)
	}

	return b.String()
}

func toTitleCase(s string) string {
	if s == "" {
		return ""
	}

	words := strings.Split(s, "_")
	for i, word := range words {
		if word != "" {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}

	return strings.Join(words, " ")
}
