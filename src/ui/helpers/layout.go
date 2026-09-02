package helpers

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// OverlayTopRight superpone un panel flotante en la esquina superior derecha sin mover el canvas base.
func OverlayTopRight(baseCanvas, floatingPanel string, totalWidth int) string {
	if floatingPanel == "" {
		return baseCanvas
	}

	panelWidth := lipgloss.Width(floatingPanel)
	if totalWidth < panelWidth {
		return baseCanvas
	}

	panelLines := strings.Split(floatingPanel, "\n")
	canvasLines := strings.Split(baseCanvas, "\n")
	leftColWidth := totalWidth - panelWidth

	for i, rLine := range panelLines {
		if i >= len(canvasLines) {
			break
		}
		leftPart := lipgloss.NewStyle().
			MaxWidth(leftColWidth).
			Inline(true).
			Render(canvasLines[i])

		padLen := max(0, leftColWidth-lipgloss.Width(leftPart))
		canvasLines[i] = leftPart + strings.Repeat(" ", padLen) + rLine
	}

	return strings.Join(canvasLines, "\n")
}
