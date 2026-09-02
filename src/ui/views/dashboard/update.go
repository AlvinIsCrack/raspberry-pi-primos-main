package dashboard

import (
	"primos/usm"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case TickMsg:
		now := time.Time(msg)
		m.CurrentTime = now

		state := usm.GetCurrentAcademicSchedule(now)
		progress, minutes := CalculateScheduleProgress(state, now)

		// El submódulo de reloj recibe el estado ya digerido
		m.Clock = m.Clock.Update(now, state, progress, minutes)

		// Actualizar snapshots de salas sin acoplar I/O a la vista
		if m.LockService != nil {
			m.RoomsSnapshots = m.LockService.GetAllSnapshots()
		}

		return m, tickCmd()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}

	return m, nil
}
