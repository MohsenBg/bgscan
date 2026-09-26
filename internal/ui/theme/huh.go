package theme

import (
	"charm.land/huh/v2"
)

// NewHuhTheme derives huh styles from the given palette so form fields
// match the rest of the TUI.
func NewHuhTheme(theme *Theme) huh.Theme {
	var t Theme
	if theme == nil || theme.Name == "" {
		if terminalLooksDark() {
			t = BGScanDark
		} else {
			t = BGScanLight
		}
	} else {
		t = *theme
	}

	return huh.ThemeFunc(func(_ bool) *huh.Styles {
		th := huh.ThemeBase(false)

		focused := &th.Focused
		focused.Title = focused.Title.Foreground(t.Info)
		focused.Description = focused.Description.Foreground(t.Muted)
		focused.ErrorIndicator = focused.ErrorIndicator.Foreground(t.Error)
		focused.ErrorMessage = focused.ErrorMessage.Foreground(t.Error)

		focused.SelectSelector = focused.SelectSelector.Foreground(t.Primary)
		focused.SelectedOption = focused.SelectedOption.Foreground(t.Primary)
		focused.UnselectedOption = focused.UnselectedOption.Foreground(t.Text)

		focused.TextInput.Cursor = focused.TextInput.Cursor.Foreground(t.Primary)
		focused.TextInput.Prompt = focused.TextInput.Prompt.Foreground(t.Secondary)
		focused.TextInput.Text = focused.TextInput.Text.Foreground(t.Text)
		focused.TextInput.Placeholder = focused.TextInput.Placeholder.Foreground(t.Muted)

		blurred := &th.Blurred
		blurred.Title = blurred.Title.Foreground(t.Secondary)
		blurred.Description = blurred.Description.Foreground(t.Muted)
		blurred.SelectSelector = blurred.SelectSelector.Foreground(t.Secondary)
		blurred.SelectedOption = blurred.SelectedOption.Foreground(t.Secondary)
		blurred.UnselectedOption = blurred.UnselectedOption.Foreground(t.Muted)

		return th
	})
}
