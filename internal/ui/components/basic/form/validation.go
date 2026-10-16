package form

import (
	"fmt"
	"sort"
	"strings"
)

// FormatValidationErrors renders field errors as a stable bulleted list.
func FormatValidationErrors[K ~string](errs map[K]error) string {
	if len(errs) == 0 {
		return ""
	}

	names := make([]string, 0, len(errs))
	for field := range errs {
		names = append(names, string(field))
	}
	sort.Strings(names)

	var b strings.Builder
	for i, name := range names {
		if i > 0 {
			b.WriteString("\n\n")
		}
		fmt.Fprintf(&b, "- %s\n  %s", toTitleCase(name), formatError(errs[K(name)]))
	}

	return b.String()
}

func formatError(err error) string {
	msg := strings.TrimSpace(err.Error())
	return strings.ReplaceAll(msg, "\n", "\n  ")
}

func toTitleCase(s string) string {
	if s == "" {
		return ""
	}

	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-'
	})

	for i, word := range words {
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}

	return strings.Join(words, " ")
}
