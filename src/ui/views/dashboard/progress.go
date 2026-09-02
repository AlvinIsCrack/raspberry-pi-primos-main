package dashboard

import (
	"time"

	"primos/usm"
)

// CalculateScheduleProgress retorna el porcentaje transcurrido (0-100) y los minutos transcurridos desde el inicio del período actual.
func CalculateScheduleProgress(state usm.CurrentScheduleState, now time.Time) (int, int) {
	currentSec := now.Hour()*3600 + now.Minute()*60 + now.Second()

	var startSec, endSec int
	switch state.Kind {
	case usm.PeriodLunch:
		startSec = usm.LunchStartMinutes * 60
		endSec = usm.LunchEndMinutes * 60

	case usm.PeriodLecture, usm.PeriodIntermission:
		startSec = (state.Block.StartHour*60 + state.Block.StartMin) * 60
		endSec = (state.Block.EndHour*60 + state.Block.EndMin) * 60
		if state.Kind == usm.PeriodIntermission {
			endSec += usm.BlockIntermissionDurationMinutes * 60
		}

	default:
		return 0, 0
	}

	total := endSec - startSec
	if total <= 0 || currentSec < startSec {
		return 0, 0
	}

	elapsedSec := min(total, currentSec-startSec)
	elapsedMinutes := elapsedSec / 60
	percent := min(100, max(0, (elapsedSec*100)/total))

	return percent, elapsedMinutes
}
