package usm_test

import (
	"testing"
	"time"

	"primos/usm"
)

func TestGetCurrentAcademicSchedule(t *testing.T) {
	loc, err := time.LoadLocation("America/Santiago")
	if err != nil {
		t.Fatalf("failed to load location: %v", err)
	}
	baseDate := time.Date(2026, 9, 14, 0, 0, 0, 0, loc)

	tests := []struct {
		name              string
		hour              int
		minute            int
		expectedKind      usm.PeriodKind
		expectedFirstIdx  int
		expectedSecondIdx int
		expectedSubBlock  int
	}{
		{
			name:         "Antes del inicio de la jornada (OffHours)",
			hour:         7,
			minute:       45,
			expectedKind: usm.PeriodOffHours,
		},
		{
			name:              "Primer bloque lectivo - Sub-bloque 1 (08:15)",
			hour:              8,
			minute:            15,
			expectedKind:      usm.PeriodLecture,
			expectedFirstIdx:  1,
			expectedSecondIdx: 2,
			expectedSubBlock:  1,
		},
		{
			name:              "Primer bloque lectivo - Sub-bloque 2 (08:50)",
			hour:              8,
			minute:            50,
			expectedKind:      usm.PeriodLecture,
			expectedFirstIdx:  1,
			expectedSecondIdx: 2,
			expectedSubBlock:  2,
		},
		{
			name:              "Receso entre bloque 1-2 y bloque 3-4 (09:30)",
			hour:              9,
			minute:            30,
			expectedKind:      usm.PeriodIntermission,
			expectedFirstIdx:  1,
			expectedSecondIdx: 2,
			expectedSubBlock:  0,
		},
		{
			name:         "Almuerzo USM (13:40 - 14:30)",
			hour:         13,
			minute:       50,
			expectedKind: usm.PeriodLunch,
		},
		{
			name:             "Bloque lectivo posalmuerzo (14:30)",
			hour:             14,
			minute:           30,
			expectedKind:     usm.PeriodLecture,
			expectedSubBlock: 1,
		},
		{
			name:         "Noche / Fuera de horario",
			hour:         23,
			minute:       0,
			expectedKind: usm.PeriodOffHours,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkTime := baseDate.Add(time.Duration(tt.hour)*time.Hour + time.Duration(tt.minute)*time.Minute)
			got := usm.GetCurrentAcademicSchedule(checkTime)

			if got.Kind != tt.expectedKind {
				t.Fatalf("esperaba Kind %v, obtuvo %v", tt.expectedKind, got.Kind)
			}

			if tt.expectedKind == usm.PeriodLecture {
				if got.ActiveSubBlock != tt.expectedSubBlock {
					t.Errorf("esperaba ActiveSubBlock %d, obtuvo %d", tt.expectedSubBlock, got.ActiveSubBlock)
				}
				if tt.expectedFirstIdx != 0 && got.Block.FirstIndex != tt.expectedFirstIdx {
					t.Errorf("esperaba FirstIndex %d, obtuvo %d", tt.expectedFirstIdx, got.Block.FirstIndex)
				}
				if tt.expectedSecondIdx != 0 && got.Block.SecondIndex != tt.expectedSecondIdx {
					t.Errorf("esperaba SecondIndex %d, obtuvo %d", tt.expectedSecondIdx, got.Block.SecondIndex)
				}
			}
		})
	}
}
