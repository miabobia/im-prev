package styles

import (
	lipgloss "github.com/charmbracelet/lipgloss"
)

func ImageBorderStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(width).
		Height(height).
		// Width(config.C.ImageBorderWidth).
		// Height(config.C.ImageBorderHeight).
		Align(lipgloss.Center, lipgloss.Center)
}

func SelectorBorderStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(width).
		Height(height).
		Align(lipgloss.Left, lipgloss.Center)
}

func BoldStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

func GreenStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#6cce42"))
}

func GreenBoldStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#6cce42"))
}
