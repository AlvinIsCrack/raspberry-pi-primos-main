package dashboard

import (
	"fmt"
	"strings"

	"primos/domain"
	"primos/ui/helpers"

	"github.com/charmbracelet/lipgloss"
)

const (
	SensorBatteryIndicatorFilled  = "="
	SensorBatteryIndicatorEmpty   = "-"
	SensorActivityIndicator       = "'"
	LastSeenActionDurationSeconds = 10
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
		if !snap.LastSeenAt.IsZero() {
			elapsed := m.CurrentTime.Sub(snap.LastSeenAt).Seconds()
			if elapsed >= 0 && elapsed < LastSeenActionDurationSeconds {
				color := m.Theme.Primary
				if elapsed > 1 {
					t := elapsed / LastSeenActionDurationSeconds // 0.0 (reciente) a 1.0 (límite)
					r := int(255 - t*(255-74))
					g := int(255 - t*(255-74))
					b := int(255 - t*(255-74))
					color = lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
				}

				activityIndicator = lipgloss.NewStyle().
					Foreground(color).
					Render(SensorActivityIndicator)
			}
		}

		// Batería
		battery := "   "
		if snap.BatteryLevel >= 15 {
			renderSegment := func(filled bool, activeColor lipgloss.TerminalColor) string {
				if filled {
					return lipgloss.NewStyle().Foreground(activeColor).Render(SensorBatteryIndicatorFilled)
				}
				return lipgloss.NewStyle().Foreground(m.Theme.Muted).Render(SensorBatteryIndicatorEmpty)
			}

			p1 := renderSegment(true, m.Theme.Error)
			p2 := renderSegment(snap.BatteryLevel > 33, m.Theme.Secondary)
			p3 := renderSegment(snap.BatteryLevel > 66, m.Theme.Primary)

			battery = fmt.Sprintf("%s%s%s", p1, p2, p3)
		} else if snap.BatteryLevel >= 0 {
			battery = lipgloss.NewStyle().Foreground(m.Theme.Critical).Faint(!isEven).Render("LOW")
		}

		lines = append(lines, fmt.Sprintf("%s %s%s %s", battery, snap.RoomID, activityIndicator, style.Render(snap.Door.String())))
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
