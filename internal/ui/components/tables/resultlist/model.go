package resultlist

import (
	"fmt"

	"bgscan/internal/core/result"
	"bgscan/internal/ui/components/basic/crud"
	"bgscan/internal/ui/components/basic/notice"
	ipviewer "bgscan/internal/ui/components/tables/ipviewer"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

type Model struct {
	id        ui.ComponentID
	name      string
	state     *ui.AppState
	maxIPs    uint32
	crudTable *crud.Model[result.ResultFile]
}

func New(state *ui.AppState, title string, maxIPs uint32, onSelect func(*result.ResultFile) tea.Cmd) *Model {
	m := &Model{
		id:     ui.NewComponentID(),
		name:   "Result Files",
		state:  state,
		maxIPs: maxIPs,
	}

	if onSelect == nil {
		onSelect = m.defaultSelectHandler
	}

	m.crudTable = crud.New(title, state.Layout, newProvider(state, title, onSelect), 100, false)

	return m
}

func (m *Model) Init() tea.Cmd      { return m.crudTable.Init() }
func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return m.crudTable.OnClose() }
func (m *Model) Mode() env.Mode     { return m.crudTable.Mode() }

// defaultSelectHandler automatically opens the matching IP details viewer component
func (m *Model) defaultSelectHandler(file *result.ResultFile) tea.Cmd {
	ips, err := result.ReadResultFile(file.Path, file.Schema)
	if err != nil {
		return notice.NewNoticeCmd(m.state.Layout, "Selection", err.Error(), notice.NOTICE_ERROR)
	}
	return m.OpenResultIP(ips)
}

// OpenResultIP loads IP results from a result file and opens the IP viewer.
func (m *Model) OpenResultIP(file result.ResultFile) tea.Cmd {
	ips, err := result.LoadAll(file.Path, file.Schema, m.maxIPs)
	if err != nil {
		return notice.NewNoticeCmd(
			m.state.Layout,
			"Result File Error",
			fmt.Sprintf("Error while reading result file: %v", err),
			notice.NOTICE_ERROR,
		)
	}

	return func() tea.Msg {
		return ui.OpenComponentMsg{
			Component: ipviewer.New(
				m.state.Layout,
				fmt.Sprintf("IP Scan [%s]", file.Schema.Name),
				ips,
				file.Schema,
			),
		}
	}
}
