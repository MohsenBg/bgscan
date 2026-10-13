package crud

import (
	"errors"
	"fmt"

	"github.com/MohsenBg/bgscan/internal/ui/components/basic/confirm"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"
	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
	"github.com/MohsenBg/bgscan/internal/ui/shared/validation"

	tea "charm.land/bubbletea/v2"
)

type (
	msgLoaded[T any] struct{ items []T }
	msgError         struct{ err error }
)

// RefreshCmd loads items from the source.
func (m *Model[T]) RefreshCmd() tea.Msg {
	items, err := m.source.Load()
	if err != nil {
		return msgError{err: err}
	}
	return msgLoaded[T]{items: items}
}

// notice opens an informational dialog using the shared environment.
func (m *Model[T]) notice(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.env,
		title, message, notice.NOTICE_INFO)
}

func (m *Model[T]) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case MsgRefresh:
		return m, m.RefreshCmd

	case msgLoaded[T]:
		m.replaceItems(msg.items)
		return m, nil

	case msgError:
		return m, notice.NewNoticeCmd(m.env, "Error", msg.err.Error(), notice.NOTICE_ERROR)

	case MsgActionTrigger:
		return m, m.dispatchAction(msg.ActionType)
	}

	// Verbs win over the table's own keys, so a provider-registered verb
	// cannot be shadowed by navigation or sorting. While the filter is
	// being typed, though, the table owns every key -- including the ones
	// verbs use, such as enter confirming a filter.
	if key, ok := msg.(tea.KeyMsg); ok && !m.table.Filtering() {
		if cmd, handled := m.handleAction(key); handled {
			return m, cmd
		}
	}

	updated, cmd := m.table.Update(msg)
	m.table = updated.(*table.Model)
	return m, cmd
}

// replaceItems rebuilds the table rows and identity index from items.
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
	return confirm.ConfirmCmd(m.env,
		fmt.Sprintf("Delete %s '%s'?", m.source.Title(), m.source.Identity(item)),
		tea.Sequence(m.hooks.OnDelete(item), func() tea.Msg { return MsgRefresh{} }),
		false,
		dialog.WithTitle("⚠ Delete "+m.source.Title()),
		dialog.WithDanger(),
	)
}

func (m *Model[T]) requestRename(item T) tea.Cmd {
	current := m.source.Identity(item)
	rename := m.hooks.OnRename

	inp := textinput.New(m.env,
		fmt.Sprintf("Enter new name for %s:", m.source.Title()),
		textinput.WithPlaceholder("new name"),
		textinput.WithValue(current),
		textinput.WithValidation(validation.ValidateFilename),
		textinput.WithFocus(),
		textinput.WithOnSubmit(func(newName string) tea.Cmd {
			return tea.Sequence(rename(item, newName), func() tea.Msg { return MsgRefresh{} })
		}),
	)

	return input.OpenInputDialog(inp)
}

// getSelected resolves the highlighted row to its item via the identity in
// the first column (see Source.RenderRow).
func (m *Model[T]) getSelected() (T, error) {
	row, ok := m.table.SelectedRow()
	if !ok {
		var zero T
		return zero, errors.New("no row selected")
	}
	item, ok := m.itemsMap[row[0]]
	if !ok {
		var zero T
		return zero, fmt.Errorf("item '%s' not found", row[0])
	}
	return item, nil
}
