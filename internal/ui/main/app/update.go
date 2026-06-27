package app

import (
	"bytes"
	"runtime"
	"runtime/pprof"
	"strings"

	"bgscan/internal/logger"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// Update is the central message router for the application.
// It processes BubbleTea messages, manages overlay layers,
// and dispatches updates to UI components.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// Handle terminal resize
	case tea.WindowSizeMsg:
		m.layout.Update(msg.Width, msg.Height)

	// Handle keyboard input
	case tea.KeyMsg:

		// Immediate application quit
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// dump Goroutine for DebugInfo
		if msg.String() == tea.KeyCtrlT.String() {
			dumpGoroutines()
		}

		// Overlay back/quit handling
		if len(m.layers) > 0 {
			lastIdx := len(m.layers) - 1
			top := m.layers[lastIdx]

			if env.IsBackKey(msg, top.Mode()) || env.IsQuitKey(msg, top.Mode()) {

				// Execute overlay cleanup command
				cmds = append(cmds, top.OnClose())

				// Remove overlay placement metadata
				delete(m.overlayPlacements, top.ID())

				// Remove overlay from stack
				m.layers[lastIdx] = nil
				m.layers = m.layers[:lastIdx]

				return m, tea.Batch(cmds...)
			}
		}

	// Add a new overlay component
	case ui.AddOverlayMsg:
		m.layers = append(m.layers, msg.Component)

		m.overlayPlacements[msg.Component.ID()] = &OverlayPlacement{
			XPos:    msg.XPos,
			YPos:    msg.YPos,
			XOffset: msg.XOffset,
			YOffset: msg.YOffset,
		}

		return m, msg.Component.Init()

	// Close an existing overlay component
	case ui.CloseComponentMsg:
		for i, ov := range m.layers {
			if ov.ID() == msg.ID {

				cmds = append(cmds, ov.OnClose())

				// Remove overlay safely from slice
				m.layers = append(m.layers[:i], m.layers[i+1:]...)

				// Remove placement metadata
				delete(m.overlayPlacements, msg.ID)

				break
			}
		}
	}

	// --- Overlay Input Routing ---

	// If overlays exist, the top overlay consumes all input.
	if len(m.layers) > 0 {
		lastIdx := len(m.layers) - 1

		newLayer, cmd := m.layers[lastIdx].Update(msg)
		m.layers[lastIdx] = newLayer

		// Block key input from reaching background components
		if _, ok := msg.(tea.KeyMsg); ok {
			return m, cmd
		}

		cmds = append(cmds, cmd)
	}

	// --- Background Component Updates ---

	var hCmd, bCmd, fCmd tea.Cmd

	m.header, hCmd = m.header.Update(msg)
	m.body, bCmd = m.body.Update(msg)
	m.footer, fCmd = m.footer.Update(msg)

	cmds = append(cmds, hCmd, bCmd, fCmd)

	return m, tea.Batch(cmds...)
}

// DumpGoroutines logs a filtered view of all goroutines,
// focusing on potentially blocked states. It also dumps
// the full goroutine profile at debug level for offline analysis.
//
// Call this when you suspect worker leaks or stuck probes.
func dumpGoroutines() {
	var buf bytes.Buffer

	count := runtime.NumGoroutine()
	logger.DebugInfo("=== Goroutine Dump (count=%d) ===", count)

	// Write full goroutine profile into buffer
	if err := pprof.Lookup("goroutine").WriteTo(&buf, 2); err != nil {
		logger.DebugError("failed to write goroutine profile: %v", err)
		return
	}

	dump := buf.String()
	lines := strings.SplitSeq(dump, "\n")

	// Filter lines with suspicious wait states
	for line := range lines {
		if strings.Contains(line, "[chan receive]") ||
			strings.Contains(line, "[chan send]") ||
			strings.Contains(line, "[select]") ||
			strings.Contains(line, "[IO wait]") ||
			strings.Contains(line, "[sleep]") {
			logger.DebugInfo("%s", line)
		}
	}
}
