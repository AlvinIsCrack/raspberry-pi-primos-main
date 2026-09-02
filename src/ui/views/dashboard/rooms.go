package dashboard

import (
	"fmt"
	"strings"
	"time"

	"primos/domain"
	"primos/ui/helpers"

	"github.com/charmbracelet/lipgloss"
)

// RenderRoomsPanel formatea la lista vertical de salas monitoreadas.
func (m Model) RenderRoomsPanel() string {
	snapshots := m.RoomsSnapshots
	if len(snapshots) == 0 && m.LockService != nil {
		snapshots = m.LockService.GetAllSnapshots()
	}
	if len(snapshots) == 0 {
		return ""
	}

	// Parpadea cada segundo alternando el color de UNKN
	isEven := m.CurrentTime.Second()%2 == 0

	lines := make([]string, 0, len(snapshots))
	for _, snap := range snapshots {
		style := lipgloss.NewStyle().Foreground(m.Theme.Muted)
		switch snap.Door {
		case domain.DoorOpen:
			style = style.Foreground(m.Theme.Primary)
		case domain.DoorClosed:
			style = style.Foreground(m.Theme.Secondary)
		default:
			style = style.Faint(isEven)
		}

		// Indicador de actividad
		activityIndicator := " "
		if !snap.LastSeenAt.IsZero() && m.CurrentTime.Sub(snap.LastSeenAt) < time.Second {
			activityIndicator = lipgloss.NewStyle().Faint(true).Render("*")
		}

		// Batería
		battery := "   "
		if snap.BatteryLevel >= 15 {
			c1 := m.Theme.Error
			c2 := m.Theme.Muted
			c3 := m.Theme.Muted

			if snap.BatteryLevel > 33 {
				c2 = m.Theme.Secondary
			}
			if snap.BatteryLevel > 66 {
				c3 = m.Theme.Primary
			}

			p1 := lipgloss.NewStyle().Foreground(c1).Render("|")
			p2 := lipgloss.NewStyle().Foreground(c2).Render("|")
			p3 := lipgloss.NewStyle().Foreground(c3).Render("|")

			battery = fmt.Sprintf("%s%s%s", p1, p2, p3)
		} else if snap.BatteryLevel >= 0 {
			battery = lipgloss.NewStyle().Foreground(m.Theme.Critical).Faint(!isEven).Render("LOW")
		}

		lines = append(lines, fmt.Sprintf("%s %s%s %s", battery, snap.RoomID, activityIndicator, style.Render(string(snap.Door))))
	}

	content := strings.Join(lines, "\n")
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.Theme.Border).
		Padding(0, 1)

	titleStyle := lipgloss.NewStyle().
		Foreground(m.Theme.Muted)

	panel := helpers.RenderTitledBox(" Puertas ", content, boxStyle, titleStyle)

	return lipgloss.NewStyle().
		Margin(1, 2).
		Render(panel)
}
