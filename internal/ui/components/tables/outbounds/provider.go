package outbounds

import (
	"bgscan/internal/core/xray"
	"bgscan/internal/logger"
	"bgscan/internal/ui/components/basic/crud"
	"bgscan/internal/ui/components/basic/notice"
	"bgscan/internal/ui/components/basic/table"
	"bgscan/internal/ui/shared/layout"
	"os"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
)

type provider struct {
	layout   *layout.Layout
	onSelect func(*xray.XrayOutboundsFile) tea.Cmd
}

func newProvider(layout *layout.Layout, onSelect func(*xray.XrayOutboundsFile) tea.Cmd) crud.Provider[xray.XrayOutboundsFile] {
	return &provider{
		layout:   layout,
		onSelect: onSelect,
	}
}

func (p *provider) Title() string { return "Outbound Templates" }

func (p *provider) Columns() []table.Column {
	return []table.Column{
		{Title: "Name", Width: 50},
		{Title: "Created Time", Width: 50},
	}
}

func (p *provider) Load() ([]xray.XrayOutboundsFile, error) {
	outbounds, err := xray.ListOutboundTemplates()
	if err != nil {
		logger.UIError("Failed to load outbounds: %s", err.Error())
		return nil, err
	}
	logger.UIInfo("Loaded %d Outbounds", len(outbounds))
	slices.SortFunc(outbounds, func(i, j xray.XrayOutboundsFile) int {
		return j.CreatedTime.Compare(i.CreatedTime)
	})
	return outbounds, nil
}

func (p *provider) RenderRow(item xray.XrayOutboundsFile) table.Row {
	return table.Row{
		item.Name,
		item.CreatedTime.Format("2006-01-02 15:04:05"),
	}
}

func (p *provider) Identity(item xray.XrayOutboundsFile) string {
	return item.Name
}

func (p *provider) OnSelect(item xray.XrayOutboundsFile) (tea.Cmd, bool) {
	if p.onSelect != nil {
		return p.onSelect(&item), true
	}
	return nil, false
}

func (p *provider) OnDelete(item xray.XrayOutboundsFile) (tea.Cmd, bool) {
	if err := os.Remove(item.Path); err != nil && !os.IsNotExist(err) {
		logger.UIError("Failed to delete outbound: %s", err.Error())
		return notice.NewNoticeCmd(p.layout, "Delete Failed", err.Error(), notice.NOTICE_ERROR), true
	}
	return nil, true
}

func (p *provider) OnRename(item xray.XrayOutboundsFile, newName string) (tea.Cmd, bool) {
	_, err := xray.RenameOutboundTemplate(item.Name, newName)
	if err != nil {
		logger.UIError("Rename failed: %v", err)
		return notice.NewNoticeCmd(p.layout, "Rename Failed", err.Error(), notice.NOTICE_ERROR), true
	}
	return nil, true
}

// We satisfy the provider interface, but since we use a custom picker workflow via `AddFunc`,
// this hook isn't directly needed for triggering updates.
func (p *provider) OnAdd(item xray.XrayOutboundsFile) (tea.Cmd, bool) {
	return nil, true
}
