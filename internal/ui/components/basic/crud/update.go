package crud

import (
	"errors"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/confirm"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"
	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/shared/validation"
)

type (
	loadedMsg[T any] struct{ items []T }
	errorMsg         struct{ err error }
)

// refreshCmd loads items from the source.
func (m *Model[T]) refreshCmd() tea.Msg {
	items, err := m.source.Load()
	if err != nil {
		return errorMsg{err: err}
	}
	return loadedMsg[T]{items: items}
}

func (m *Model[T]) showNotice(title, msg string, level notice.LEVEL) tea.Cmd {
	return notice.NewNoticeCmd(m.env, title, msg, level)
}

func (m *Model[T]) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case RefreshMsg:
		m.table.SetLoading(true)
		return m, m.refreshCmd

	case loadedMsg[T]:
		m.table.SetLoading(false)
		m.replaceItems(msg.items)
		return m, nil
	case errorMsg:
		m.table.SetLoading(false)
		return m, m.showNotice("Error", msg.err.Error(), notice.NOTICE_ERROR)
	case TriggerActionMsg:
		return m, m.dispatchAction(msg.Action)
	}

	// Verbs override table navigation keys, unless the user is actively typing in the filter.
	if key, ok := msg.(tea.KeyMsg); ok && !m.table.Filtering() {
		if cmd, handled := m.handleAction(key); handled {
			return m, cmd
		}
	}

	updated, cmd := m.table.Update(msg)
	m.table = updated.(*table.Model)
	return m, cmd
}

// replaceItems rebuilds the table rows and identity index.
func (m *Model[T]) replaceItems(items []T) {
	m.items = items
	clear(m.itemsMap)

	rows := make([]table.Row, 0, len(items))
	for _, item := range items {
		id := m.source.Identity(item)
		m.itemsMap[id] = item
		rows = append(rows, m.source.RenderRow(item))
	}
	m.table.SetRows(rows)
}

func (m *Model[T]) requestDeletion(item T) tea.Cmd {
	title := m.source.Title()
	id := m.source.Identity(item)

	return confirm.ConfirmCmd(m.env,
		fmt.Sprintf("Delete %s %q?", title, id),
		tea.Sequence(m.hooks.OnDelete(item), func() tea.Msg { return RefreshMsg{} }),
		false,
		dialog.WithTitle("⚠ Delete "+title),
		dialog.WithDanger(),
	)
}

func (m *Model[T]) requestRename(item T) tea.Cmd {
	current := m.source.Identity(item)
	rename := m.hooks.OnRename

	inp := textinput.New(m.env,
		fmt.Sprintf("Enter new name for %s:", m.source.Name()),
		textinput.WithPlaceholder("new name"),
		textinput.WithValue(current),
		textinput.WithValidation(validation.ValidateFilename),
		textinput.WithFocus(),
		textinput.WithOnSubmit(func(newName string) tea.Cmd {
			return tea.Sequence(rename(item, newName), func() tea.Msg { return RefreshMsg{} })
		}),
	)
	return input.OpenInputDialog(inp, dialog.WithTitle(fmt.Sprintf("Rename %s", m.source.Name())))
}

// getSelected resolves the highlighted row to its underlying item.
// Assumes the first cell of the rendered row is the identity string.
func (m *Model[T]) getSelected() (T, error) {
	var zero T

	row, ok := m.table.SelectedRow()
	if !ok {
		return zero, errors.New("no row selected")
	}

	item, ok := m.itemsMap[row[0]]
	if !ok {
		return zero, fmt.Errorf("item %q not found", row[0])
	}
	return item, nil
}
