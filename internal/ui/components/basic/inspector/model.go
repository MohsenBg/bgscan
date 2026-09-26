package inspector

import (
	"fmt"
	"io"
	"strings"

	"github.com/MohsenBg/bgscan/internal/logger"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/tabs"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/layout"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/theme"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type FieldInput interface {
	input.Dialog
	Value() any
	SetValue(any)
}

// Field is one inspectable property, rendered as one row in the list.
type Field struct {
	Name        string
	Description string
	Group       string
	Input       FieldInput
	Visible     func() bool
	Format      func(any) string
	snapshot    *any
}

// value returns the last committed display value. Input.Value() is not read
// here because it reflects live, uncommitted edits.
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

func (f Field) visible() bool {
	if f.Visible == nil {
		return true
	}
	return f.Visible()
}

type FieldItem struct {
	Field Field
}

func (i FieldItem) FilterValue() string { return i.Field.Name }

// fieldDelegate renders "Name  value" rows with an optional description line.
type fieldDelegate struct {
	theme *theme.Theme
}

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
		name = selectedFieldNameStyle(d.theme).Render("▶ " + name)
	} else {
		name = fieldNameStyle(d.theme).Render(name)
	}
	leftSection := name
	rightSection := valueStyle(d.theme).Render(f.value())

	gap := max(m.Width()-lipgloss.Width(leftSection)-lipgloss.Width(rightSection), 1)

	line := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftSection,
		strings.Repeat(" ", gap),
		rightSection,
	)

	_, _ = fmt.Fprint(w, PaddingCell().Render(line))
}

// Model is a tabbed list of fields grouped by Field.Group.
type Model struct {
	id     ui.ComponentID
	name   string
	layout *layout.Layout
	theme  *theme.Theme
	log    logger.Set

	Title string

	groups map[string][]Field
	order  []string

	tabs     ui.Component
	list     list.Model
	maxWidth int

	widthOverride  int
	heightOverride int
}

type Option func(*Model)

func WithWidth(width int) Option {
	return func(m *Model) {
		if width > 0 {
			m.widthOverride = width
		}
	}
}

func WithHeight(height int) Option {
	return func(m *Model) {
		if height > 0 {
			m.heightOverride = height
		}
	}
}

// New builds an inspector from a flat field list, grouped by Field.Group
// into tabs. Groups are ordered alphabetically for a stable tab order
// across runs (map iteration order is not stable in Go).
func New(deps ui.Deps, name string, fields []Field, opts ...Option) *Model {
	m := &Model{
		id:       ui.NewComponentID(),
		name:     name,
		layout:   deps.Layout,
		theme:    deps.Theme,
		log:      deps.Log,
		maxWidth: 60,
	}

	for _, opt := range opts {
		opt(m)
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

	tb := tabs.New(deps, tbs, func(_ int, tab tabs.Tab[[]Field]) tea.Cmd {
		return func() tea.Msg {
			return tabChangeMsg{Group: tab.Label, Fields: tab.Value}
		}
	})
	tb.SetMaxWidth(m.maxWidth)
	m.tabs = tb

	if len(tbs) > 0 {
		m.Title = tbs[0].Label
		m.list = newFieldList(tbs[0].Value, m.theme, m.Width(), m.Height())
	}

	return m
}

func newFieldList(fields []Field, th *theme.Theme, width, height int) list.Model {
	lm := list.New(visibleItems(fields), fieldDelegate{theme: th}, width, height)
	lm.DisableQuitKeybindings()
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
				key.WithHelp(env.KeyEnter, "edit selected field"),
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

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Theme() *theme.Theme { return m.theme }

func (m *Model) SetTheme(th *theme.Theme) {
	m.theme = th
	if tb, ok := m.tabs.(*tabs.Model[[]Field]); ok {
		tb.SetTheme(th)
	}
	m.list.SetDelegate(fieldDelegate{theme: th})
	// Field.Inputs cache their own styles and open as dialogs later through
	// these same pointers, so theme them here too. (Groups hold Field
	// copies, but Input is shared by pointer.)
	for _, fields := range m.groups {
		for _, f := range fields {
			if t, ok := f.Input.(ui.Themed); ok && t != nil {
				t.SetTheme(th)
			}
		}
	}
}
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return nil }
func (m *Model) Mode() env.Mode     { return env.NormalMode }

func (m *Model) Width() int {
	if m.widthOverride > 0 {
		return m.widthOverride
	}
	if m.layout == nil {
		return m.maxWidth
	}
	return min(m.maxWidth, m.layout.BodyContentWidth())
}

func (m *Model) Height() int {
	padding := 4

	takenHeight := 0
	if len(m.groups) > 0 {
		takenHeight += lipgloss.Height(m.tabs.View())
	}

	if m.Title != "" {
		takenHeight += lipgloss.Height(m.Title)
	}

	total := m.heightOverride
	if total <= 0 {
		if m.layout == nil {
			return 15
		}
		total = m.layout.BodyContentHeight()
	}

	available := total - takenHeight - padding
	return max(15, available)
}

func (m *Model) SetWidth(width int) {
	m.widthOverride = width
	if tb, ok := m.tabs.(*tabs.Model[[]Field]); ok {
		tb.SetMaxWidth(width)
	}
	m.list.SetWidth(width)
}

func (m *Model) SetHeight(height int) {
	m.heightOverride = height
	m.list.SetHeight(m.Height())
}

func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}

func (m *Model) SelectedField() (Field, bool) {
	item, ok := m.list.SelectedItem().(FieldItem)
	if !ok {
		return Field{}, false
	}
	return item.Field, true
}

func (m *Model) Fields() []Field {
	return m.groups[m.Title]
}

// Refresh re-pulls Value() from every Input and re-renders, used after an
// external edit changes underlying state.
func (m *Model) Refresh() tea.Cmd {
	for _, f := range m.Fields() {
		*f.snapshot = f.Input.Value()
	}
	return m.list.SetItems(visibleItems(m.groups[m.Title]))
}
