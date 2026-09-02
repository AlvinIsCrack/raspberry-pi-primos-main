package components

import (
	"github.com/charmbracelet/lipgloss"
)

// RenderCard wraps arbitrary text content inside a styled container.
func RenderCard(title, content string, width int, borderStyle lipgloss.Style) string {
	titleHeader := lipgloss.NewStyle().Bold(true).MarginBottom(1).Render(title)
	body := lipgloss.JoinVertical(lipgloss.Left, titleHeader, content)

	return borderStyle.Width(width).Render(body)
}
