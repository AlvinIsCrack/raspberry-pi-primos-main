package dashboard

import (
	"time"

	"primos/domain"
	"primos/services"
	"primos/ui"
	"primos/ui/views/dashboard/clock"
	"primos/usm"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TickMsg conveys clock pulse events.
type TickMsg time.Time

// Model contains minimal state to display the current academic block, time, and room availability.
type Model struct {
	Theme          ui.Theme
	LockService    *services.RoomsLockService
	CardStyle      lipgloss.Style
	CurrentTime    time.Time
	ScheduleState  usm.CurrentScheduleState
	Clock          clock.Model
	RoomsSnapshots []domain.SensorSnapshot
	Width          int
	Height         int
}

// NewModel instantiates a minimal dashboard model with default room availability.
func NewModel(theme ui.Theme, lockService *services.RoomsLockService) Model {
	return Model{
		Theme:       theme,
		LockService: lockService,
		Clock:       clock.New(theme),
		CardStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border).
			Padding(0).
			Align(lipgloss.Center),
	}
}

// tickCmd schedules a UI update every second.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Init triggers initial data fetching and starts the ticker.
func (m Model) Init() tea.Cmd {
	return tickCmd()
}
