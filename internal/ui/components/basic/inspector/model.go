package inspector

import (
	"fmt"
	"io"
	"strings"

	"bgscan/internal/ui/components/basic/input"
	"bgscan/internal/ui/components/basic/tabs"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type FiledInput interface {
	input.Dialog
	Value() any
	SetValue(any)
}

// Field describes a single inspectable property, rendered as one row in
// the field list.
type Field struct {
	Name        string
	Description string
	Group       string
	Input       FiledInput
	Visible     func() bool
	Format      func(any) string
	snapshot    *any
}

// value returns the field's last committed display value, using Format if
// set, otherwise falling back to fmt.Sprint. This intentionally does not
// read Input.Value() directly, since that reflects live, uncommitted edits.
func (f Field) value() string {
	if f.snapshot == nil {
		return ""
	}
	v := *f.snapshot
	if f.Format != nil {
		return f.Format(v)
	}
	return fmt.Sprint(v)
}

// visible reports whether the field should currently be shown.
func (f Field) visible() bool {
	if f.Visible == nil {
		return true
	}
	return f.Visible()
}

// --- list item / delegate ---------------------------------------------------

// FieldItem adapts a Field to list.Item.
type FieldItem struct {
	Field Field
}

func (i FieldItem) FilterValue() string { return i.Field.Name }

// fieldDelegate renders a Field as: "Name  value" on the left, the field's
// Key as a shortcut hint on the right, with an optional description line.
type fieldDelegate struct{}

func (d fieldDelegate) Height() int                             { return 2 }
func (d fieldDelegate) Spacing() int                            { return 0 }
func (d fieldDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d fieldDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(FieldItem)
	if !ok {
		return
	}
	f := item.Field

	name := f.Name
	if index == m.Index() {
		name = selectedFiledNameStyle().Render("▶ " + name)
	} else {
		name = filedNameStyle().Render(name)
	}
	leftSection := name
	rightSection := vlaueStyle().Render(f.value())

	gap := max(m.Width()-lipgloss.Width(leftSection)-lipgloss.Width(rightSection), 1)

	line := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftSection,
		strings.Repeat(" ", gap),
		rightSection,
	)

	fmt.Fprint(w, PaddingCell().Render(line))
}

// --- Model -----------------------------------------------------------------

// Model is the inspector component: a tabbed list of fields grouped by
// Field.Group, where each row shows the field's name, current value, and
// key shortcut.
type Model struct {
	id     ui.ComponentID
	name   string
	layout *layout.Layout

	Title string

	groups map[string][]Field
	order  []string

	tabs     ui.Component
	list     list.Model
	maxWidth int
}

// New creates a new inspector from a flat list of fields, grouped by
// Field.Group into tabs. Groups are ordered alphabetically for a stable
// tab order across runs (map iteration order is not stable in Go).
func New(l *layout.Layout, name string, fields []Field) *Model {
	m := &Model{
		id:       ui.NewComponentID(),
		name:     name,
		layout:   l,
		maxWidth: 60,
	}

	for i := range fields {
		f := &fields[i]
		if f.Input == nil {
			continue
		}
		v := f.Input.Value()
		f.snapshot = &v
		f.Input.AppendOnSubmit(func() tea.Cmd {
			m.Refresh()
			return nil
		})
	}

	groups := make(map[string][]Field)
	order := make([]string, 0)
	for _, f := range fields {
		if _, seen := groups[f.Group]; !seen {
			order = append(order, f.Group)
		}
		groups[f.Group] = append(groups[f.Group], f)
	}

	tbs := make([]tabs.Tab[[]Field], 0, len(order))
	for _, group := range order {
		tbs = append(tbs, tabs.Tab[[]Field]{
			Label: group,
			Value: groups[group],
		})
	}

	m.groups = groups
	m.order = order

	tb := tabs.New(l, tbs, func(_ int, tab tabs.Tab[[]Field]) tea.Cmd {
		return func() tea.Msg {
			return tabChangeMsg{Group: tab.Label, Fields: tab.Value}
		}
	})
	tb.SetMaxWidth(m.maxWidth + 5)
	m.tabs = tb

	if len(tbs) > 0 {
		m.Title = tbs[0].Label
		m.list = newFieldList(tbs[0].Value, m.Width(), m.Height())
	}

	return m
}

func newFieldList(fields []Field, width, height int) list.Model {
	lm := list.New(visibleItems(fields), fieldDelegate{}, width, height)
	lm.SetShowStatusBar(false)
	lm.SetShowTitle(false)
	lm.SetFilteringEnabled(true)
	lm.SetShowHelp(true)
	lm.SetFilteringEnabled(false)
	lm.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys(env.KeyEnter),
				key.WithHelp(env.KeyEnter, "edit"),
			),
			key.NewBinding(
				key.WithKeys("d"),
				key.WithHelp("d", "description"),
			),
		}
	}

	lm.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys(env.KeyEnter),
				key.WithHelp(env.KeyEnter, "edit selected filed"),
			),
			key.NewBinding(
				key.WithKeys("d"),
				key.WithHelp("d", "show description"),
			),
		}
	}
	return lm
}

func visibleItems(fields []Field) []list.Item {
	items := make([]list.Item, 0, len(fields))
	for _, f := range fields {
		if !f.visible() {
			continue
		}
		items = append(items, FieldItem{Field: f})
	}
	return items
}

// --- ui.Component ------------------------------------------------------------

func (m *Model) Init() tea.Cmd      { return nil }
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return nil }
func (m *Model) Mode() env.Mode     { return env.NormalMode }

// Width calculates the maximum width of the inspector.
func (m *Model) Width() int {
	if m.layout == nil {
		return m.maxWidth
	}
	return min(m.maxWidth, m.layout.Body.Width)
}

// Height calculates the height of the inspector
func (m *Model) Height() int {
	if m.layout == nil {
		return 30
	}

	takenHeight := 0
	if len(m.groups) > 0 {
		takenHeight += lipgloss.Height(m.tabs.View())
	}

	if m.Title != "" {
		takenHeight += lipgloss.Height(m.Title)
	}

	available := m.layout.Body.Height - takenHeight
	maxHeight := min(30, available)

	return maxHeight
}

// CloseCmd returns a command that closes this component.
func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}

// --- accessors ---------------------------------------------------------------

// SelectedField returns the field currently highlighted in the list.
func (m *Model) SelectedField() (Field, bool) {
	item, ok := m.list.SelectedItem().(FieldItem)
	if !ok {
		return Field{}, false
	}
	return item.Field, true
}

// Fields returns the fields belonging to the currently active group/tab.
func (m *Model) Fields() []Field {
	return m.groups[m.Title]
}

// Refresh re-pulls Value() from every field's Input and re-renders the
// list, useful after an external edit changes underlying state.
func (m *Model) Refresh() tea.Cmd {
	for _, f := range m.Fields() {
		*f.snapshot = f.Input.Value()
	}
	return m.list.SetItems(visibleItems(m.groups[m.Title]))
}
