package ipviewer

import (
	"fmt"

	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == env.KeyEnter {
		if cmd := m.copySelectedIP(); cmd != nil {
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *Model) copySelectedIP() tea.Cmd {
	t, ok := m.table.(*table.Model)
	if !ok || t == nil {
		return nil
	}

	row := t.BubbleTable.SelectedRow()
	if len(row) == 0 {
		return nil
	}

	if err := clipboard.WriteAll(row[0]); err != nil {
		return m.errorCmd("Error Copying IP", err.Error())
	}
	return m.infoCmd("IP Copied", fmt.Sprintf("IP copied to clipboard:%s", row[0]))
}

func (m *Model) errorCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.layout, title, message, notice.NOTICE_ERROR)
}

func (m *Model) infoCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.layout, title, message, notice.NOTICE_INFO)
}
