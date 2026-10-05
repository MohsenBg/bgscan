// Package keys provides key bindings and help hints.
package keys

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

type Binding struct {
	Action Action
	Keys   []string // Keys matched by Bubble Tea, e.g. ["up", "k"].

	HelpKey   string // Optional key shown in help.
	HelpText  string // Full help label.
	ShortText string // Short help label. Falls back to HelpText.

	Hidden   bool // Omitted from short help.
	Disabled bool // Omitted from all help.

	// CaseSensitive controls whether key matching is case-sensitive.
	CaseSensitive bool
}

type Hint struct {
	Key   string
	Label string
}

type Hints []Hint

// Matches reports whether the pressed key triggers this binding.
func (b Binding) Matches(k tea.KeyMsg) bool {
	if b.Disabled || len(b.Keys) == 0 {
		return false
	}

	pressed := k.String()
	if !b.CaseSensitive {
		pressed = strings.ToLower(pressed)
	}

	return slices.ContainsFunc(b.Keys, func(key string) bool {
		if !b.CaseSensitive {
			key = strings.ToLower(key)
		}

		return key == pressed
	})
}

// LongItem is the full help entry for this binding.
// It shows every active key bound to the action.
func (b Binding) LongItem() Hint {
	return Hint{
		Key:   b.longKey(),
		Label: b.HelpText,
	}
}

// ShortItem is the compact help entry for this binding.
func (b Binding) ShortItem() Hint {
	label := b.ShortText
	if label == "" {
		label = b.HelpText
	}

	return Hint{
		Key:   b.shortKey(),
		Label: label,
	}
}

func (b Binding) shortKey() string {
	if b.HelpKey != "" {
		return DisplayKey(b.HelpKey)
	}

	for _, key := range b.Keys {
		if key != "" {
			return DisplayKey(key)
		}
	}

	return ""
}

func (b Binding) longKey() string {
	seen := make(map[string]struct{}, len(b.Keys))
	displayed := make([]string, 0, len(b.Keys))

	for _, key := range b.Keys {
		if key == "" {
			continue
		}

		display := DisplayKey(key)

		id := display
		if !b.CaseSensitive {
			id = strings.ToLower(id)
		}

		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		displayed = append(displayed, display)
	}

	if len(displayed) == 0 && b.HelpKey != "" {
		return DisplayKey(b.HelpKey)
	}

	return strings.Join(displayed, "/")
}

// Combine merges several bindings into one short help entry.
// Example: "←/→ move".
func Combine(label string, bs ...Binding) Hint {
	seen := make(map[string]struct{}, len(bs))
	keys := make([]string, 0, len(bs))

	for _, b := range bs {
		if b.Disabled {
			continue
		}

		key := b.shortKey()
		if key == "" {
			continue
		}

		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		keys = append(keys, key)
	}

	return Hint{
		Key:   strings.Join(keys, "/"),
		Label: label,
	}
}

// Short returns compact help hints, skipping hidden and disabled bindings.
func Short(bs ...Binding) Hints {
	out := make(Hints, 0, len(bs))

	for _, b := range bs {
		if b.Hidden || b.Disabled {
			continue
		}

		item := b.ShortItem()
		if item.Key == "" {
			continue
		}

		out = append(out, item)
	}

	return out
}

// Long returns full help hints, including hidden bindings.
// Disabled bindings are skipped.
func Long(bs ...Binding) Hints {
	out := make(Hints, 0, len(bs))

	for _, b := range bs {
		if b.Disabled {
			continue
		}

		item := b.LongItem()
		if item.Key == "" {
			continue
		}

		out = append(out, item)
	}

	return out
}

var keyAliases = map[string]string{
	"up":        "↑",
	"down":      "↓",
	"left":      "←",
	"right":     "→",
	"enter":     "⏎",
	"return":    "⏎",
	"esc":       "esc",
	"escape":    "esc",
	"space":     "Space",
	"tab":       "tab",
	"pgup":      "PgUp",
	"pageup":    "PgUp",
	"pgdown":    "PgDn",
	"pagedown":  "PgDn",
	"home":      "Home",
	"end":       "End",
	"backspace": "⌫",
	"delete":    "Del",
	"insert":    "Ins",
}

var modifierNames = map[string]string{
	"ctrl":  "Ctrl",
	"ctl":   "Ctrl",
	"shift": "Shift",
	"alt":   "Alt",
	"meta":  "Meta",
	"cmd":   "Cmd",
	"super": "Super",
	"hyper": "Hyper",
}

// DisplayKey converts a Bubble Tea key string into a nicer help label.
func DisplayKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}

	lower := strings.ToLower(key)

	if display, ok := keyAliases[lower]; ok {
		return display
	}

	if strings.Contains(lower, "+") {
		return displayCombo(lower)
	}

	// Preserve printable single characters such as ?, q, /, [.
	if utf8.RuneCountInString(key) == 1 {
		return key
	}

	return titleCase(lower)
}

func displayCombo(combo string) string {
	parts := strings.Split(combo, "+")
	out := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if name, ok := modifierNames[part]; ok {
			out = append(out, name)
			continue
		}

		if display, ok := keyAliases[part]; ok {
			out = append(out, display)
			continue
		}

		if utf8.RuneCountInString(part) == 1 {
			out = append(out, strings.ToUpper(part))
			continue
		}

		out = append(out, titleCase(part))
	}

	return strings.Join(out, "+")
}

func titleCase(s string) string {
	if s == "" {
		return ""
	}

	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])

	return string(r)
}
