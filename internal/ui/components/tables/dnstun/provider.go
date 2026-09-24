package dnstun

import (
	"os"
	"slices"

	"github.com/MohsenBg/bgscan/internal/core/dns"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/crud"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/notice"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/table"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

type source struct {
	state *ui.AppState
}

func newSource(state *ui.AppState) crud.Source[dns.DNSTunConfigFile] {
	return &source{
		state: state,
	}
}

func (s *source) Title() string {
	return "DNS Tunnels"
}

func (s *source) Columns() []table.Column {
	return []table.Column{
		{Title: "Name", Width: 40},
		{Title: "Protocol", Width: 15},
		{Title: "Auth", Width: 15},
		{Title: "Created Time", Width: 20},
	}
}

func (s *source) Load() ([]dns.DNSTunConfigFile, error) {
	configs, err := dns.GetAllDNSTunsFile()
	if err != nil {
		s.state.Log.UI.Error("Failed to load DNS tunnel configs: %s", err)
		return nil, err
	}

	s.state.Log.UI.Info("Loaded %d DNS tunnel configs", len(configs))

	slices.SortFunc(configs, func(i, j dns.DNSTunConfigFile) int {
		return j.CreatedAt.Compare(i.CreatedAt)
	})

	return configs, nil
}

func (s *source) RenderRow(item dns.DNSTunConfigFile) table.Row {
	return table.Row{
		item.Name,
		string(item.Protocol),
		item.Proxy,
		item.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (s *source) Identity(item dns.DNSTunConfigFile) string {
	return item.Name
}

// newHooks wires the mutation behaviors. OnAdd only registers the "add"
// keybinding: adding a tunnel goes through the protocol menu intercepted
// in Update via crud.MsgActionTrigger, so its body never runs.
func newHooks(state *ui.AppState, onSelect func(*dns.DNSTunConfigFile) tea.Cmd) crud.Hooks[dns.DNSTunConfigFile] {
	hooks := crud.Hooks[dns.DNSTunConfigFile]{
		OnDelete: func(item dns.DNSTunConfigFile) tea.Cmd {
			if err := os.Remove(item.Path); err != nil && !os.IsNotExist(err) {
				state.Log.UI.Error("Failed to delete DNS tunnel config: %s", err)

				return notice.NewNoticeCmd(state.Deps(),
					"Delete Failed",
					err.Error(),
					notice.NOTICE_ERROR,
				)
			}

			return nil
		},
		OnRename: func(item dns.DNSTunConfigFile, newName string) tea.Cmd {
			if err := dns.RenameDNSTunConfigFile(item, newName); err != nil {
				state.Log.UI.Error("Rename failed: %v", err)

				return notice.NewNoticeCmd(state.Deps(),
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
		hooks.OnSelect = func(item dns.DNSTunConfigFile) tea.Cmd {
			return onSelect(&item)
		}
	}

	return hooks
}
