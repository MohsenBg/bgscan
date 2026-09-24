package crud

import (
	"errors"
	"fmt"

	"github.com/MohsenBg/bgscan/internal/ui/components/basic/confirm"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input/textinput"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"
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

func (m *Model[T]) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	if ui.HandleTheme(msg, m) {
		return m, nil
	}

	switch msg := msg.(type) {
	case MsgRefresh:
		return m, m.RefreshCmd

	case msgLoaded[T]:
		m.replaceItems(msg.items)
		return m, nil

	case msgError:
		return m, notice.NewNoticeCmd(ui.Deps{Layout: m.layout, Theme: m.theme, Log: m.log}, "Error", msg.err.Error(), notice.NOTICE_ERROR)

	case MsgActionTrigger:
		return m, m.dispatchAction(msg.ActionType)
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

// dispatchAction runs the handler for name, resolving the selected item
// first when the action needs one.
func (m *Model[T]) dispatchAction(name string) tea.Cmd {
	entry, ok := m.actions[name]
	if !ok {
		return nil
	}

	var item T
	if entry.needsItem {
		selected, err := m.getSelected()
		if err != nil {
			return notice.NewNoticeCmd(ui.Deps{Layout: m.layout, Theme: m.theme, Log: m.log}, "Selection", err.Error(), notice.NOTICE_INFO)
		}
		item = selected
	}

	return entry.run(item)
}

func (m *Model[T]) requestDeletion(item T) tea.Cmd {
	return confirm.ConfirmCmd(ui.Deps{Layout: m.layout, Theme: m.theme, Log: m.log},
		fmt.Sprintf("Delete %s '%s'?", m.source.Title(), m.source.Identity(item)),
		tea.Sequence(m.hooks.OnDelete(item), func() tea.Msg { return MsgRefresh{} }),
		false,
	)
}

func (m *Model[T]) requestRename(item T) tea.Cmd {
	current := m.source.Identity(item)
	rename := m.hooks.OnRename

	inp := textinput.New(ui.Deps{Layout: m.layout, Theme: m.theme, Log: m.log},
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
	row := m.table.BubbleTable.SelectedRow()
	if len(row) == 0 {
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
