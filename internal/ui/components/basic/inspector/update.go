package inspector

import (
	"fmt"

	"bgscan/internal/ui/components/basic/input"
	"bgscan/internal/ui/components/basic/notice"
	"bgscan/internal/ui/shared/dialog"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Update processes incoming Bubble Tea messages, delegating first to the
// tabs component, then handling group switches, then forwarding everything
// else to the field list.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmds []tea.Cmd

	if m.tabs != nil {
		updated, cmd := m.tabs.Update(msg)
		m.tabs = updated
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tabChangeMsg:
		m.Title = msg.Group
		cmd := m.list.SetItems(visibleItems(msg.Fields))
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	case tea.WindowSizeMsg:
		m.list.SetWidth(m.Width())
		m.list.SetHeight(m.Height())
	case tea.KeyPressMsg:
		if msg.Code == tea.KeyEnter {
			filed, ok := m.SelectedField()
			if ok && filed.Input != nil && filed.snapshot != nil {
				filed.Input.SetValue(*filed.snapshot)
				cmds = append(
					cmds, input.OpenInputDialog(
						filed.Input,
						dialog.WithPosition(dialog.Center, dialog.Top),
						dialog.WithOffset(0, 3),
					),
				)
			}
		}

		if msg.String() == "d" {
			filed, ok := m.SelectedField()
			if ok && filed.Input != nil {
				title := fmt.Sprintf("Description %s", filed.Name)
				return m, notice.NewNoticeCmd(m.layout, title, filed.Description, notice.NOTICE_INFO)
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
