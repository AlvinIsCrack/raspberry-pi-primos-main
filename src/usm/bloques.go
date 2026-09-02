package usm

import (
	"time"
)

const (
	BlockDurationMinutes             = 35
	BlockIntermissionDurationMinutes = 15
	LunchStartMinutes                = 13*60 + 40
	LunchEndMinutes                  = 14*60 + 30

	DayStartHour   = 8
	DayStartMinute = 15
	TotalBlocks    = 10
)

// PeriodKind identifica el tipo de momento académico actual.
type PeriodKind int

const (
	PeriodOffHours PeriodKind = iota
	PeriodLunch
	PeriodLecture
	PeriodIntermission
)

// AcademicBlock defines the temporal bounds and numeric indices for a double-period lecture slot.
type AcademicBlock struct {
	FirstIndex  int
	SecondIndex int
	StartHour   int
	StartMin    int
	EndHour     int
	EndMin      int
}

// CurrentScheduleState representa el estado académico calculado sin ataduras a presentación visual.
type CurrentScheduleState struct {
	Kind           PeriodKind
	Block          AcademicBlock
	ActiveSubBlock int
}

// buildStandardBlocks genera los bloques lectivos en tiempo de inicio de forma determinista.
func buildStandardBlocks() []AcademicBlock {
	blocks := make([]AcademicBlock, TotalBlocks)
	cursor := DayStartHour*60 + DayStartMinute

	for i := range TotalBlocks {
		firstIdx := (i * 2) + 1
		secondIdx := firstIdx + 1

		// Si el cursor cae dentro del almuerzo, salta al término del almuerzo
		if cursor >= LunchStartMinutes && cursor < LunchEndMinutes {
			cursor = LunchEndMinutes
		}

		endCursor := cursor + (BlockDurationMinutes * 2)

		blocks[i] = AcademicBlock{
			FirstIndex:  firstIdx,
			SecondIndex: secondIdx,
			StartHour:   cursor / 60,
			StartMin:    cursor % 60,
			EndHour:     endCursor / 60,
			EndMin:      endCursor % 60,
		}

		// Avanza el descanso estándar para el siguiente bloque
		cursor = endCursor + BlockIntermissionDurationMinutes
	}

	return blocks
}

var standardBlocks = buildStandardBlocks()

// GetCurrentAcademicSchedule evalúa la hora contra los bloques lectivos, descansos y almuerzo.
func GetCurrentAcademicSchedule(t time.Time) CurrentScheduleState {
	current := t.Hour()*60 + t.Minute()

	if current >= LunchStartMinutes && current < LunchEndMinutes {
		return CurrentScheduleState{Kind: PeriodLunch}
	}

	for i, block := range standardBlocks {
		start := block.StartHour*60 + block.StartMin
		end := block.EndHour*60 + block.EndMin

		if current >= start && current <= end {
			midpoint := start + BlockDurationMinutes
			activeSub := 1
			if current >= midpoint {
				activeSub = 2
			}
			return CurrentScheduleState{
				Kind:           PeriodLecture,
				Block:          block,
				ActiveSubBlock: activeSub,
			}
		}

		if i < len(standardBlocks)-1 {
			nextStart := standardBlocks[i+1].StartHour*60 + standardBlocks[i+1].StartMin
			if current > end && current < nextStart {
				return CurrentScheduleState{
					Kind:           PeriodIntermission,
					Block:          block,
					ActiveSubBlock: 0,
				}
			}
		}
	}

	return CurrentScheduleState{Kind: PeriodOffHours}
}
