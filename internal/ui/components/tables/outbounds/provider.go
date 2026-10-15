package outbounds

import (
	"os"
	"slices"

	tea "charm.land/bubbletea/v2"
	"github.com/MohsenBg/bgscan/internal/core/xray"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"
)

type source struct {
	session *ui.Session
}

func newSource(session *ui.Session) crud.Source[xray.OutboundFile] {
	return &source{session: session}
}

func (s *source) Title() string { return "Xray Outbounds" }

func (s *source) Name() string {
	return "outbound"
}

func (s *source) Columns() []table.Column {
	return []table.Column{
		{Title: "Name", Width: 40},
		{Title: "Protocol", Width: 15},
		{Title: "Network", Width: 15},
		{Title: "Security", Width: 10},
		{Title: "Created", Width: 20},
	}
}

func (s *source) Load() ([]xray.OutboundFile, error) {
	coreLog := s.session.CoreLog()
	outbounds, err := xray.ListOutboundTemplates(&coreLog)
	if err != nil {
		s.session.Env().Log.Error("Failed to load outbounds: %v", err)
		return nil, err
	}

	// Sort newest first
	slices.SortFunc(outbounds, func(i, j xray.OutboundFile) int {
		return j.CreatedTime.Compare(i.CreatedTime)
	})
	return outbounds, nil
}

func (s *source) RenderRow(item xray.OutboundFile) table.Row {
	return table.Row{
		item.Name,
		string(item.Protocol),
		string(item.Network),
		string(item.Security),
		item.CreatedTime.Format("2006-01-02 15:04:05"),
	}
}

func (s *source) Identity(item xray.OutboundFile) string {
	return item.Name
}

func newHooks(session *ui.Session, onSelect func(*xray.OutboundFile) tea.Cmd, add func() tea.Cmd) crud.Hooks[xray.OutboundFile] {
	hooks := crud.Hooks[xray.OutboundFile]{
		OnDelete: func(item xray.OutboundFile) tea.Cmd {
			return func() tea.Msg {
				if err := os.Remove(item.Path); err != nil && !os.IsNotExist(err) {
					session.Env().Log.Error("Failed to delete outbound: %v", err)
					cmd := notice.NewNoticeCmd(session.Env(), "Delete Failed", err.Error(), notice.NOTICE_ERROR)
					if cmd != nil {
						return cmd()
					}
				}
				return nil
			}
		},
		OnRename: func(item xray.OutboundFile, newName string) tea.Cmd {
			return func() tea.Msg {
				if _, err := xray.RenameOutboundTemplate(item.Name, newName); err != nil {
					session.Env().Log.Error("Rename failed: %v", err)
					cmd := notice.NewNoticeCmd(session.Env(), "Rename Failed", err.Error(), notice.NOTICE_ERROR)
					if cmd != nil {
						return cmd()
					}
				}
				return nil
			}
		},
		OnAdd: add,
	}

	if onSelect != nil {
		hooks.OnSelect = func(item xray.OutboundFile) tea.Cmd {
			return onSelect(&item)
		}
	}
	return hooks
}
