package outbounds

import (
	"os"
	"slices"

	"github.com/MohsenBg/bgscan/internal/core/xray"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

type source struct {
	session *ui.Session
}

func newSource(session *ui.Session) crud.Source[xray.OutboundFile] {
	return &source{
		session: session,
	}
}

func (s *source) Title() string { return "Outbound Templates" }

func (s *source) Columns() []table.Column {
	return []table.Column{
		{Title: "Name", Width: 40},
		{Title: "Protocol", Width: 15},
		{Title: "Network", Width: 15},
		{Title: "Security", Width: 10},
		{Title: "Created Time", Width: 20},
	}
}

func (s *source) Load() ([]xray.OutboundFile, error) {
	coreLog := s.session.CoreLog()
	outbounds, err := xray.ListOutboundTemplates(&coreLog)
	if err != nil {
		s.session.Env().Log.Error("Failed to load outbounds: %v", err)
		return nil, err
	}

	s.session.Env().Log.Info("Loaded %d Outbounds", len(outbounds))

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

// newHooks wires the mutation behaviors. OnAdd only registers the "add"
// keybinding: adding an outbound goes through the custom picker workflow
// intercepted in Update via crud.MsgActionTrigger, so its body never runs.
func newHooks(session *ui.Session, onSelect func(*xray.OutboundFile) tea.Cmd) crud.Hooks[xray.OutboundFile] {
	hooks := crud.Hooks[xray.OutboundFile]{
		OnDelete: func(item xray.OutboundFile) tea.Cmd {
			if err := os.Remove(item.Path); err != nil && !os.IsNotExist(err) {
				session.Env().Log.Error("Failed to delete outbound: %v", err)

				return notice.NewNoticeCmd(session.Env(),
					"Delete Failed",
					err.Error(),
					notice.NOTICE_ERROR,
				)
			}

			return nil
		},
		OnRename: func(item xray.OutboundFile, newName string) tea.Cmd {
			if _, err := xray.RenameOutboundTemplate(item.Name, newName); err != nil {
				session.Env().Log.Error("Rename failed: %v", err)

				return notice.NewNoticeCmd(session.Env(),
					"Rename Failed",
					err.Error(),
					notice.NOTICE_ERROR,
				)
			}

			return nil
		},
		OnAdd: func() tea.Cmd { return nil },
	}

	if onSelect != nil {
		hooks.OnSelect = func(item xray.OutboundFile) tea.Cmd {
			return onSelect(&item)
		}
	}

	return hooks
}
