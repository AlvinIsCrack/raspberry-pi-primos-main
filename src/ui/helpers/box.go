package helpers

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderTitledBox renderiza un bloque con un borde y su título incrustado en la línea superior.
func RenderTitledBox(title string, content string, boxStyle lipgloss.Style, titleStyle lipgloss.Style) string {
	rendered := boxStyle.Render(content)
	lines := strings.Split(rendered, "\n")
	if len(lines) == 0 {
		return ""
	}

	border := boxStyle.GetBorderStyle()
	borderFg := boxStyle.GetBorderTopForeground()
	borderStyle := lipgloss.NewStyle().Foreground(borderFg)

	leftCorner := borderStyle.Render(border.TopLeft + border.Top)
	titleRendered := titleStyle.Render(title)
	rightCorner := borderStyle.Render(border.TopRight)

	lineWidth := lipgloss.Width(lines[0])
	usedWidth := lipgloss.Width(leftCorner) + lipgloss.Width(titleRendered) + lipgloss.Width(rightCorner)

	if lineWidth >= usedWidth {
		fill := borderStyle.Render(strings.Repeat(border.Top, lineWidth-usedWidth))
		lines[0] = leftCorner + titleRendered + fill + rightCorner
	}

	return strings.Join(lines, "\n")
}
