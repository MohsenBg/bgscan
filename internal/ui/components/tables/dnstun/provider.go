package dnstun

import (
	"os"
	"slices"

	"bgscan/internal/core/dns"
	"bgscan/internal/logger"
	"bgscan/internal/ui/components/basic/crud"
	"bgscan/internal/ui/components/basic/notice"
	"bgscan/internal/ui/components/basic/table"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

type provider struct {
	state    *ui.AppState
	onSelect func(*dns.DNSTunConfigFile) tea.Cmd
}

func newProvider(
	state *ui.AppState,
	onSelect func(*dns.DNSTunConfigFile) tea.Cmd,
) crud.Provider[dns.DNSTunConfigFile] {
	return &provider{
		state:    state,
		onSelect: onSelect,
	}
}

func (p *provider) Title() string {
	return "DNS Tunnels"
}

func (p *provider) Columns() []table.Column {
	return []table.Column{
		{Title: "Name", Width: 40},
		{Title: "Protocol", Width: 15},
		{Title: "Auth", Width: 15},
		{Title: "Created Time", Width: 20},
	}
}

func (p *provider) Load() ([]dns.DNSTunConfigFile, error) {
	configs, err := dns.GetAllDNSTunsFile()
	if err != nil {
		logger.UIError("Failed to load DNS tunnel configs: %s", err)
		return nil, err
	}

	logger.UIInfo("Loaded %d DNS tunnel configs", len(configs))

	slices.SortFunc(configs, func(i, j dns.DNSTunConfigFile) int {
		return j.CreatedAt.Compare(i.CreatedAt)
	})

	return configs, nil
}

func (p *provider) RenderRow(item dns.DNSTunConfigFile) table.Row {
	return table.Row{
		item.Name,
		string(item.Protocol),
		item.Proxy,
		item.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (p *provider) Identity(item dns.DNSTunConfigFile) string {
	return item.Name
}

func (p *provider) OnSelect(item dns.DNSTunConfigFile) (tea.Cmd, bool) {
	if p.onSelect != nil {
		return p.onSelect(&item), true
	}

	return nil, false
}

func (p *provider) OnDelete(item dns.DNSTunConfigFile) (tea.Cmd, bool) {
	if err := os.Remove(item.Path); err != nil && !os.IsNotExist(err) {
		logger.UIError("Failed to delete DNS tunnel config: %s", err)

		return notice.NewNoticeCmd(
			p.state.Layout,
			"Delete Failed",
			err.Error(),
			notice.NOTICE_ERROR,
		), true
	}

	return nil, true
}

func (p *provider) OnRename(item dns.DNSTunConfigFile, newName string) (tea.Cmd, bool) {
	if err := dns.RenameDNSTunConfigFile(item, newName); err != nil {
		logger.UIError("Rename failed: %v", err)

		return notice.NewNoticeCmd(
			p.state.Layout,
			"Rename Failed",
			err.Error(),
			notice.NOTICE_ERROR,
		), true
	}

	return nil, true
}

func (p *provider) OnAdd(item dns.DNSTunConfigFile) (tea.Cmd, bool) {
	return nil, true
}
