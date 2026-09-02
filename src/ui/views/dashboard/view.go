package dashboard

import (
	"primos/ui/helpers"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return ""
	}

	// Render de submódulos
	cardContent := m.Clock.View(m.CardStyle)
	roomsContent := m.RenderRoomsPanel()

	// Composición geométrica central
	baseCanvas := lipgloss.Place(
		m.Width,
		m.Height,
		lipgloss.Center,
		lipgloss.Center,
		cardContent,
	)

	// Superposición flotante declarativa
	return helpers.OverlayTopRight(baseCanvas, roomsContent, m.Width)
}
