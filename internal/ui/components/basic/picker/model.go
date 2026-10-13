package picker

import (
	"os"

	"github.com/MohsenBg/bgscan/internal/ui/shared/keys"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	"charm.land/bubbles/v2/filepicker"
	tea "charm.land/bubbletea/v2"
)

// Model is a file picker overlay wrapping BubbleTea's filepicker.Model and
// integrating it with the component/overlay system.
type Model struct {
	id   ui.ComponentID
	name string

	// Overlay title displayed by the layout
	Title string

	env *ui.Env

	// Underlying BubbleTea file picker
	FilePicker filepicker.Model

	// Callback triggered when a file is selected
	OnSelect OnSelect

	// inDialog is true while the picker is hosted inside a dialog
	// frame: the title and key hints then live in the frame's borders.
	inDialog bool
}

// Init initializes the underlying file picker component.
func (m *Model) Init() tea.Cmd {
	return m.FilePicker.Init()
}

func (m *Model) ID() ui.ComponentID {
	return m.id
}

func (m *Model) Name() string {
	return m.name
}

// SetInDialog moves the picker's title and key hints into the dialog
// frame's border lines.
func (m *Model) SetInDialog(inDialog bool) { m.inDialog = inDialog }

// Hints returns the picker's key hints as raw data, excluding Cancel.
func Hints() []keys.Hint {
	return []keys.Hint{
		{Key: "← →", Label: "dir"},
		{Key: "↑ ↓", Label: "move"},
		{Key: "⏎", Label: "select"},
	}
}

// CloseCmd emits CloseComponentMsg; the router removes the overlay.
func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}

// New creates a file picker overlay. baseDir defaults to the user's home
// directory when empty; allowType restricts selectable extensions; onSelect
// runs after a file is selected (a no-op is used if nil).
func New(env *ui.Env, title string, baseDir string, allowType []string, onSelect OnSelect) *Model {
	p := filepicker.New()

	if baseDir != "" {
		p.CurrentDirectory = baseDir
	} else {
		p.CurrentDirectory, _ = os.UserHomeDir()
	}

	if len(allowType) > 0 {
		p.AllowedTypes = allowType
	}

	// Ensure callback is never nil
	if onSelect == nil {
		onSelect = func(path string) tea.Cmd { return nil }
	}

	p.ShowPermissions = true
	p.AutoHeight = false
	p.SetHeight(pickerHeight(env))

	return &Model{
		id:         ui.NewComponentID(),
		name:       "Pick File",
		Title:      title,
		env:        env,
		FilePicker: p,
		OnSelect:   onSelect,
	}
}

func (m *Model) OnClose() tea.Cmd {
	return nil
}

// Mode returns KeyModeNormal for standard keyboard navigation.
func (m *Model) Mode() ui.KeyMode {
	return ui.KeyModeNormal
}
