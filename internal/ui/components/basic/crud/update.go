package crud

import (
	"errors"
	"fmt"

	"bgscan/internal/logger"
	"bgscan/internal/ui/components/basic/confirm"
	"bgscan/internal/ui/components/basic/input"
	"bgscan/internal/ui/components/basic/input/textinput"
	"bgscan/internal/ui/components/basic/notice"
	"bgscan/internal/ui/components/basic/table"
	"bgscan/internal/ui/shared/ui"
	"bgscan/internal/ui/shared/validation"

	tea "charm.land/bubbletea/v2"
)

type (
	msgLoaded[T any] struct{ items []T }
	msgError         struct{ err error }
)

// RefreshCmd loads items from the provider.
func (m *Model[T]) RefreshCmd() tea.Msg {
	items, err := m.provider.Load()
	if err != nil {
		return msgError{err: err}
	}
	return msgLoaded[T]{items: items}
}

func (m *Model[T]) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {

	case MsgRefresh:
		return m, m.RefreshCmd

	case msgLoaded[T]:
		m.items = msg.items
		clear(m.itemsMap)
		for _, item := range msg.items {
			id := m.provider.Identity(item)
			m.itemsMap[id] = item
		}
		m.syncRows()
		return m, nil

	case MsgActionTrigger:
		var cmd tea.Cmd
		switch msg.ActionType {
		case "select":
			cmd = m.handleSelect()
		case "add":
		case "delete":
			cmd = m.requestDeletion()
		case "rename":
			cmd = m.handleRequestRename()
		}
		if cmd != nil {
			return m, cmd
		}

	case msgError:
		logger.UIError("[%s] operation failed: %v", m.name, msg.err)
		return m, notice.NewNoticeCmd(m.layout, "Error", msg.err.Error(), notice.NOTICE_ERROR)
	}

	updatedTable, cmd := m.table.Update(msg)
	m.table = updatedTable.(*table.Model)
	return m, cmd
}

func (m *Model[T]) syncRows() {
	rows := make([]table.Row, 0, len(m.items))
	for _, item := range m.items {
		rows = append(rows, m.provider.RenderRow(item))
	}
	m.table.SetRows(rows)
}

func (m *Model[T]) handleSelect() tea.Cmd {
	item, err := m.getSelected()
	if err != nil {
		return notice.NewNoticeCmd(m.layout, "Selection", err.Error(), notice.NOTICE_INFO)
	}
	if cmd, ok := m.provider.OnSelect(item); ok {
		return cmd
	}
	return nil
}

func (m *Model[T]) requestDeletion() tea.Cmd {
	item, err := m.getSelected()
	if err != nil {
		return notice.NewNoticeCmd(m.layout, "Selection", err.Error(), notice.NOTICE_INFO)
	}
	delCmd, ok := m.provider.OnDelete(item)
	if !ok {
		return nil
	}

	row := m.table.BubbleTable.SelectedRow()
	if len(row) == 0 {
		return nil
	}
	itemID := row[0] // Based on getSelected() logic, row[0] is the ID

	return confirm.ConfirmCmd(
		m.layout,
		fmt.Sprintf("Delete %s '%s'?", m.name, itemID),
		tea.Sequence(delCmd, func() tea.Msg { return MsgRefresh{} }),
		false,
	)
}

func (m *Model[T]) handleRequestRename() tea.Cmd {
	item, err := m.getSelected()
	if err != nil {
		return notice.NewNoticeCmd(m.layout, "Selection", err.Error(), notice.NOTICE_INFO)
	}

	row := m.table.BubbleTable.SelectedRow()
	if len(row) == 0 {
		return nil
	}
	itemID := row[0]

	inp := textinput.New(
		m.layout,
		fmt.Sprintf("Enter new name for %s:", m.name),
		textinput.WithPlaceholder("new name"),
		textinput.WithValue(itemID), // Pre-fill with current ID/name
		textinput.WithValidation(validation.ValidateFilename),
		textinput.WithFocus(),
		textinput.WithOnSubmit(func(newName string) tea.Cmd {
			cmd, ok := m.provider.OnRename(item, newName)
			if !ok {
				return nil
			}
			return tea.Sequence(cmd, func() tea.Msg { return MsgRefresh{} })
		}),
	)

	return input.OpenInputDialog(inp)
}

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
