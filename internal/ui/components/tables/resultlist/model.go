package resultlist

import (
	"bgscan/internal/core/result"
	"bgscan/internal/ui/components/basic/crud"
	"bgscan/internal/ui/components/basic/notice"
	ipviewer "bgscan/internal/ui/components/tables/ipviewer"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	id        ui.ComponentID
	name      string
	layout    *layout.Layout
	maxIPs    uint32
	crudTable *crud.Model[result.ResultFile]
}

func New(l *layout.Layout, name string, maxIPs uint32, onSelect func(*result.ResultFile) tea.Cmd) *Model {
	m := &Model{
		id:     ui.NewComponentID(),
		name:   name,
		layout: l,
		maxIPs: maxIPs,
	}

	if onSelect == nil {
		onSelect = m.defaultSelectHandler
	}

	m.crudTable = crud.New(name, l, newProvider(l, onSelect), false)

	return m
}

func (m *Model) Init() tea.Cmd      { return m.crudTable.Init() }
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return m.crudTable.OnClose() }
func (m *Model) Mode() env.Mode     { return m.crudTable.Mode() }

// defaultSelectHandler automatically opens the matching IP details viewer component
func (m *Model) defaultSelectHandler(file *result.ResultFile) tea.Cmd {
	ips, err := result.ReadResultFileIPs(file.Path)
	if err != nil {
		return notice.NewNoticeCmd(m.layout, "Selection", err.Error(), notice.NOTICE_ERROR)
	}
	return m.OpenResultIP(ips)
}

// OpenResultIP loads IP results from a result file and opens the IP viewer.
func (m *Model) OpenResultIP(file result.ResultFile) tea.Cmd {
	ips, err := result.LoadAll(file.Path, int64(m.maxIPs))
	if err != nil {
		return notice.NewNoticeCmd(
			m.layout,
			"Result File Error",
			fmt.Sprintf("Error while reading result file: %v", err),
			notice.NOTICE_ERROR,
		)
	}

	return func() tea.Msg {
		mode := ipviewer.ShortView

		if file.Type == result.ResultXRAY {
			mode = ipviewer.FullView
		}

		return ui.OpenComponentMsg{
			Component: ipviewer.New(
				m.layout,
				fmt.Sprintf("IP Scan [%s]", file.Type.String()),
				ips,
				mode,
			),
		}
	}
}
