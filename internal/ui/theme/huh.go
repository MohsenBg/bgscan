package theme

import (
	"charm.land/huh/v2"
)

// NewHuhTheme builds a huh.Theme backed by the app's centralized palette
// (theme.Dark / theme.Light), so huh fields (select, confirm, text input,
// etc.) match the rest of the UI instead of huh's built-in themes.
func NewHuhTheme() huh.Theme {
	return huh.ThemeFunc(func(_ bool) *huh.Styles {
		t := Current()

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
