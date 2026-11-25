// Package picker provides a file picker overlay for dialogs.
// Model wraps the bubbles file picker, limits files by extension
// when the caller passes allowed types, and reports the chosen
// path through the OnSelect callback before it closes.
package picker

import (
	"os"
	"slices"

	"charm.land/bubbles/v2/filepicker"
	tea "charm.land/bubbletea/v2"

	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

// Model is a file picker overlay.
type Model struct {
	id   ui.ComponentID
	name string
	env  *ui.Env
	// FilePicker is the inner bubbles file picker.
	FilePicker filepicker.Model
	// OnSelect runs with the chosen path when the user selects a file.
	OnSelect OnSelect
}

// New creates a file picker overlay.
// baseDir defaults to the user's home directory when empty.
func New(env *ui.Env, title, baseDir string, allowedTypes []string, onSelect OnSelect) *Model {
	p := filepicker.New()
	p.KeyMap = pickerKeyMap(env)

	if baseDir == "" {
		baseDir, _ = os.UserHomeDir()
	}
	p.CurrentDirectory = baseDir

	if len(allowedTypes) > 0 {
		p.AllowedTypes = allowedTypes
	}
	if onSelect == nil {
		onSelect = func(string) tea.Cmd { return nil }
	}

	p.ShowPermissions = true
	p.AutoHeight = false
	p.SetHeight(pickerHeight(env))

	if title == "" {
		title = "Pick File"
	}

	return &Model{
		id:         ui.NewComponentID(),
		name:       title,
		env:        env,
		FilePicker: p,
		OnSelect:   onSelect,
	}
}

// Init starts the inner file picker.
func (m *Model) Init() tea.Cmd { return m.FilePicker.Init() }

// ID returns the component ID.
func (m *Model) ID() ui.ComponentID { return m.id }

// Name returns the picker title, which is also the component name.
func (m *Model) Name() string { return m.name }

// OnClose returns no command. The picker holds no resources to free.
func (m *Model) OnClose() tea.Cmd { return nil }

// Mode returns the key mode the component uses.
func (m *Model) Mode() ui.KeyMode { return ui.KeyModeNormal }

// DialogHints describes picker keys for the dialog frame.
func (m *Model) DialogHints() keys.Hints {
	k := m.env.Keys.Picker

	return keys.Hints{
		keys.Combine("move", k.Up, k.Down),
		keys.Combine("dir", lastKeyHint(k.Back), lastKeyHint(k.Open)),
		keys.Combine("open", k.Open, k.Select),
		keys.Combine("back", k.Back),
	}
}

// CloseCmd emits CloseComponentMsg so the router removes the overlay.
func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.id}
	}
}

// pickerKeyMap maps shared picker bindings to the bubbles filepicker keymap.
func pickerKeyMap(env *ui.Env) filepicker.KeyMap {
	k := env.Keys.Picker
	km := filepicker.DefaultKeyMap()

	km.GoToTop.SetKeys(k.GotoTop.Keys...)
	km.GoToLast.SetKeys(k.GotoLast.Keys...)
	km.Up.SetKeys(k.Up.Keys...)
	km.Down.SetKeys(k.Down.Keys...)
	km.PageUp.SetKeys(k.PageUp.Keys...)
	km.PageDown.SetKeys(k.PageDown.Keys...)
	km.Back.SetKeys(k.Back.Keys...)
	km.Open.SetKeys(k.Open.Keys...)
	km.Select.SetKeys(k.Select.Keys...)

	return km
}

// lastKeyHint shows the last bound key as the primary hint key.
func lastKeyHint(b keys.Binding) keys.Binding {
	if len(b.Keys) == 0 {
		return b
	}

	out := b
	out.Keys = slices.Clone(b.Keys)
	out.Keys[0] = b.Keys[len(b.Keys)-1]

	return out
}
