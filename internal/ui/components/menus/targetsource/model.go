// Package targetsource lets the user choose where scan targets come from
// before a scan starts. The menu offers a saved target list file or the
// results of a previous scan. Picking a source opens the matching file
// picker, and the picked file path goes to the scan type menu next.
package targetsource

import (
	"github.com/MohsenBg/bgscan/internal/core/result"
	"github.com/MohsenBg/bgscan/internal/core/target"
	"github.com/MohsenBg/bgscan/internal/ui/components/basic/menu"
	"github.com/MohsenBg/bgscan/internal/ui/components/menus/scantype"
	resultlistTable "github.com/MohsenBg/bgscan/internal/ui/components/tables/resultlist"
	targetlistTable "github.com/MohsenBg/bgscan/internal/ui/components/tables/targetlist"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

// Model is the target source selection component. It wraps a basic menu
// and forwards messages and rendering to it.
type Model struct {
	id      ui.ComponentID
	name    string
	session *ui.Session
	menu    ui.Component
}

// New creates the target source menu bound to session.
func New(session *ui.Session) *Model {
	m := &Model{
		session: session,
		id:      ui.NewComponentID(),
		name:    "Target Source",
	}
	items := []menu.MenuItem{
		menu.NewMenuItem("▣", "Target List", "i", func() tea.Msg {
			return m.OpenIPList(func(i *target.FileInfo) tea.Cmd {
				return ui.OpenComponentCmd(scantype.New(session, i.Path))
			})
		}).WithDescription("Pick a list of IPs or domains to scan"),
		menu.NewMenuItem("▤", "Result List", "r", func() tea.Msg {
			return m.OpenResultIPList(func(r *result.ResultFile) tea.Cmd {
				return ui.OpenComponentCmd(scantype.New(session, r.Path))
			})
		}).WithDescription("Reuse IPs from a previous scan result"),
	}

	m.menu = menu.New(items, "Select Target Source", session.Env())
	return m
}

// OpenIPList returns a message that opens the target file picker overlay.
// The picker calls onSelect once the user picks a file.
func (m *Model) OpenIPList(onSelect func(*target.FileInfo) tea.Cmd) tea.Msg {
	return ui.OpenComponentMsg{Component: targetlistTable.New(m.session, "Select Target File", onSelect)}
}

// OpenResultIPList returns a message that opens the result file picker
// overlay. The picker calls onSelect once the user picks a file.
func (m *Model) OpenResultIPList(onSelect func(*result.ResultFile) tea.Cmd) tea.Msg {
	// The cap keeps a very large result file from slowing down rendering.
	var maxRenderIP uint32 = 10_000
	return ui.OpenComponentMsg{Component: resultlistTable.New(m.session, "Select IP Result File", maxRenderIP, onSelect)}
}

// Init starts the underlying menu and returns its first command.
func (m *Model) Init() tea.Cmd {
	return m.menu.Init()
}

// ID returns the component ID assigned at creation.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the component display name.
func (m *Model) Name() string {
	return m.name
}

// OnClose runs when the component closes. It has nothing to clean up and
// returns nil.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// Mode returns the key mode the component uses, which is normal mode.
func (m *Model) Mode() ui.KeyMode {
	return ui.KeyModeNormal
}
