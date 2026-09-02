package clock

import (
	"fmt"
	"time"

	"primos/ui"
	"primos/ui/components"
	"primos/ui/helpers"
	"primos/usm"

	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Theme         ui.Theme
	CurrentTime   time.Time
	ScheduleState usm.CurrentScheduleState
	Progress      int
	Minutes       int
	IsTickEven    bool
}

func New(theme ui.Theme) Model {
	return Model{Theme: theme}
}

func (m Model) Update(t time.Time, state usm.CurrentScheduleState, progress, minutes int) Model {
	m.CurrentTime = t
	m.ScheduleState = state
	m.Progress = progress
	m.Minutes = minutes
	m.IsTickEven = t.Second()%2 == 0
	return m
}

// FormatScheduleBlock construye la etiqueta textual del bloque o receso activo.
func FormatScheduleBlock(state usm.CurrentScheduleState, theme ui.Theme) string {
	switch state.Kind {
	case usm.PeriodLunch:
		return "Almuerzo"
	case usm.PeriodOffHours:
		return "Fuera de horario"

	case usm.PeriodLecture, usm.PeriodIntermission:
		firstStr := fmt.Sprintf("%d", state.Block.FirstIndex)
		secondStr := fmt.Sprintf("%d", state.Block.SecondIndex)
		extraStr := ""

		activeStyle := lipgloss.NewStyle().Foreground(theme.Secondary).Bold(true)

		switch state.Kind {
		case usm.PeriodLecture:
			switch state.ActiveSubBlock {
			case 1:
				firstStr = activeStyle.Render(firstStr)
			case 2:
				secondStr = activeStyle.Render(secondStr)
			}
		case usm.PeriodIntermission:
			extraStr = " " + activeStyle.Render("Recreo")
		}

		return fmt.Sprintf("Bloque %s-%s%s", firstStr, secondStr, extraStr)
	default:
		return ""
	}
}

func (m Model) View(cardStyle lipgloss.Style) string {
	dateStyle := lipgloss.NewStyle().Foreground(m.Theme.Muted)
	timeStyle := lipgloss.NewStyle().Bold(true).Foreground(m.Theme.Primary).Padding(0, 3)
	blockStyle := lipgloss.NewStyle().Foreground(lipgloss.NoColor{})

	hours := timeStyle.Render(helpers.RenderBigTime(m.CurrentTime.Format("15")))
	colon := timeStyle.Faint(m.IsTickEven).Render(helpers.RenderBigTime(":"))
	minutes := timeStyle.Render(helpers.RenderBigTime(m.CurrentTime.Format("04")))

	renderedTime := lipgloss.JoinHorizontal(lipgloss.Top, hours, colon, minutes)
	clockWidth := lipgloss.Width(renderedTime)

	dateLabel := helpers.FormatSpanishDate(m.CurrentTime)
	centeredDate := lipgloss.NewStyle().Width(clockWidth).Align(lipgloss.Center).Render(dateStyle.Render(dateLabel))

	blockLabel := FormatScheduleBlock(m.ScheduleState, m.Theme)
	centeredBlock := lipgloss.NewStyle().Width(clockWidth).Align(lipgloss.Center).Render(blockStyle.Render(blockLabel))

	var label string
	switch m.ScheduleState.Kind {
	case usm.PeriodLunch:
		label = "%p en almuerzo"
	case usm.PeriodOffHours:
		label = ""
	default:
		if m.Progress >= 98 {
			label = lipgloss.NewStyle().Foreground(m.Theme.Primary).Faint(!m.IsTickEven).Render("Turno finalizado")
		} else if m.IsTickEven {
			label = fmt.Sprintf("%dm en turno", m.Minutes)
		} else {
			label = "%p en turno"
		}
	}

	centeredProgressBar := lipgloss.NewStyle().
		Width(clockWidth).
		Align(lipgloss.Center).
		Render(components.RenderProgressBar(components.ProgressBarConfig{
			Percent:     m.Progress,
			TotalWidth:  clockWidth,
			FgColor:     m.Theme.Primary,
			EmptyColor:  m.Theme.Muted,
			LabelFormat: label,
			FilledChar:  "=",
			EmptyChar:   "-",
		}))

	return cardStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Center,
			centeredDate,
			renderedTime,
			" ",
			centeredBlock,
			centeredProgressBar,
		),
	)
}
