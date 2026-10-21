package inspector

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/help"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/input"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

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
		cmds = append(cmds, m.list.SetItems(visibleItems(msg.Fields)))
		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		m.list.SetWidth(m.Width())
		height := m.Height()
		if m.list.Paginator.TotalPages > 1 {
			// reserve one row for the pagination indicator
			height--
		}
		m.list.SetHeight(height)
		if h, ok := m.help.(*help.Model); ok {
			h.SetShortWidth(m.Width())
		}

	case tea.KeyMsg:
		if cmd := m.handleKey(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	if m.help != nil {
		m.help, cmd = m.help.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleKey(msg tea.KeyMsg) tea.Cmd {
	if m.env.Keys.Inspector.Edit.Matches(msg) {
		field, ok := m.SelectedField()
		if !ok || field.Input == nil || field.snapshot == nil {
			return nil
		}
		field.Input.SetValue(*field.snapshot)
		return input.OpenInputDialog(
			field.Input,
			dialog.WithPosition(dialog.Center, dialog.Top),
			dialog.WithOffset(0, 3),
			dialog.WithTitle(field.Name),
		)
	}

	if m.env.Keys.Inspector.Description.Matches(msg) {
		field, ok := m.SelectedField()
		if !ok || field.Input == nil {
			return nil
		}
		title := fmt.Sprintf("Description %s", field.Name)
		return notice.NewNoticeCmd(m.env, title, field.Description, notice.NOTICE_INFO)
	}

	return nil
}
