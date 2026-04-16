package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/miabobia/im-prev/config"
)

func handleKeyPress(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "down", "j":
		return handleNavigation(m, msg)
	case "i", "I", "f", "F", "?", "esc":
		return handleModeToggle(m, msg)
	case "C", "c":
		return handleClipboard(m, msg)
	case "ctrl+c", "q", "Q":
		return m, tea.Quit
	}
	return m, nil
}

func handleClipboard(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "C":
		m.clipboardManager.CopyImage(m.entries[m.cursor].name)

	case "c":
		m.clipboardManager.CopyString(m.entries[m.cursor].name)
	}
	return m, nil
}

func handleModeToggle(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "i", "I":
		if m.mode == Info {
			m.mode = Default
		} else {
			m.mode = Info
		}

	case "f", "F":
		if m.mode == FullScreen {
			m.mode = Default
		} else {
			m.mode = FullScreen
		}

	case "?":
		if m.mode == Help {
			m.mode = Default
		} else {
			m.mode = Help
		}

	case "esc":
		m.mode = Default
	}
	return m, nil
}

func handleNavigation(m model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

		start, end := getWindowRange(m)
		var cmds []tea.Cmd

		for i := start; i <= end; i++ {
			if m.entries[i].loadState == Unloaded {
				m.entries[i].loadState = Loading
				cmds = append(cmds, loadImageCmd(m.chafaPath, m.layout, m.entries[i], i))
			}
		}

		// evict the entry that fell off the bottom of the window
		evict := end + 1
		if evict < len(m.entries) {
			m.entries[evict].data = "NOT\nLOADED"
			m.entries[evict].loadState = Unloaded
		}

		return m, tea.Batch(cmds...)

	case "down", "j":
		if m.cursor < len(m.entries)-1 {
			m.cursor++
		}

		start, end := getWindowRange(m)
		var cmds []tea.Cmd

		for i := start; i <= end; i++ {
			if m.entries[i].loadState == Unloaded {
				m.entries[i].loadState = Loading
				cmds = append(cmds, loadImageCmd(m.chafaPath, m.layout, m.entries[i], i))
			}
		}

		// evict the entry that fell off the top of the window
		evict := start - 1
		if evict >= 0 {
			m.entries[evict].data = "NOT\nLOADED"
			m.entries[evict].loadState = Unloaded
		}

		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func handleWindowResize(m model, msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	cmds = append(cmds, tea.ClearScreen)

	w, h := msg.Width, msg.Height
	m.layout = computeLayout(w, h, config.C)

	if w <= config.C.MinimumScreenWidth || h <= config.C.MinimumScreenHeight {
		m.mode = TooSmall
	} else {
		if m.mode == TooSmall {
			m.mode = Default
		}
	}

	winStart, winEnd := getWindowRange(m)
	// recreate all chafa images
	// wipe all current chafa images
	for index, entry := range m.entries {
		entry.data = "NOT\nLOADED"
		entry.loadState = Unloaded
		if index >= winStart && index <= winEnd {
			cmds = append(cmds, loadImageCmd(m.chafaPath, m.layout, entry, index))
		}
	}
	return m, tea.Batch(cmds...)
}
