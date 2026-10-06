package app

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/MohsenBg/bgscan/internal/ui/main/splash"
	"github.com/MohsenBg/bgscan/internal/ui/main/startup"
	"github.com/MohsenBg/bgscan/internal/ui/shared/dialog"
	"github.com/MohsenBg/bgscan/internal/ui/shared/env"
	"github.com/MohsenBg/bgscan/internal/ui/shared/ui"

	tea "charm.land/bubbletea/v2"
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ui.SyncState(msg, m.state)
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.state.Layout.Update(msg.Width, msg.Height)

	case tea.KeyPressMsg:
		if msg.String() == env.KeyCtrlC {
			return m, tea.Quit
		}

		if msg.String() == env.KeyCtrlT {
			m.state.Log.Debug.Info("%s", dumpGoroutines())
		}

		if cmd, closed := m.closeTopDialog(msg); closed {
			return m, cmd
		}

	case splash.SplashDoneMsg:
		cmds = append(cmds, m.splash.OnClose(), m.startup.Init())
		m.stage = StageStartUP

	case startup.StartupChecksDoneMsg:
		cmds = append(cmds, m.startup.OnClose(), m.workspace.Init())
		m.stage = StageWorkspace

	case dialog.OpenDialogMsg:
		m.push(msg)
		return m, msg.Component.Init()

	case ui.CloseComponentMsg:
		cmds = append(cmds, m.close(msg.ID))
	}

	if _, themed := msg.(ui.ThemeChangedMsg); themed {
		cmds = append(cmds, m.updateAll(msg)...)
	} else if cmd, active := m.updateTop(msg); active {
		if _, isKey := msg.(tea.KeyMsg); isKey {
			return m, cmd
		}
		cmds = append(cmds, cmd)
	}

	var cmd tea.Cmd
	switch m.stage {
	case StageSplash:
		m.splash, cmd = m.splash.Update(msg)
	case StageStartUP:
		m.startup, cmd = m.startup.Update(msg)
	case StageWorkspace:
		m.workspace, cmd = m.workspace.Update(msg)
	}

	cmds = append(cmds, cmd)
	return m, tea.Sequence(cmds...)
}

// closeTopDialog closes the active dialog for a back or quit key.
func (m *model) closeTopDialog(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	top := m.top()
	if top == nil {
		return nil, false
	}

	if !env.IsBackKey(msg, top.Mode()) && !env.IsQuitKey(msg, top.Mode()) {
		return nil, false
	}

	return m.popTop(), true
}

func dumpGoroutines() string {
	buf := make([]byte, 1<<20)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			buf = buf[:n]
			break
		}
		buf = make([]byte, len(buf)*2)
	}
	suspiciousStates := []string{
		"[chan receive",
		"[chan send",
		"[select",
		"[IO wait",
		"[sleep",
		"[semacquire",
		"[sync.Mutex.Lock",
	}

	blocks := strings.Split(string(buf), "\n\n")
	var matched []string
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		for _, state := range suspiciousStates {
			if strings.Contains(block, state) {
				matched = append(matched, block)
				break
			}
		}
	}

	total := runtime.NumGoroutine()
	header := fmt.Sprintf("=== Goroutine Dump: %d total, %d suspicious ===\n", total, len(matched))

	if len(matched) == 0 {
		return header + "(none)\n"
	}

	return header + strings.Join(matched, "\n\n") + "\n"
}
