package services_test

import (
	"testing"
	"time"

	"primos/domain"
	"primos/services"
)

func TestGetEnergyPolicy(t *testing.T) {
	loc := time.UTC
	// Lunes 14 de Septiembre de 2026
	monday := time.Date(2026, 9, 14, 0, 0, 0, 0, loc)
	// Sábado 19 de Septiembre de 2026
	saturday := time.Date(2026, 9, 19, 12, 0, 0, 0, loc)
	// Domingo 20 de Septiembre de 2026
	sunday := time.Date(2026, 9, 20, 12, 0, 0, 0, loc)

	tests := []struct {
		name     string
		dateTime time.Time
		expected domain.EnergyPolicy
	}{
		{
			name:     "Día laboral mañana (10:00) -> DEFAULT",
			dateTime: monday.Add(10 * time.Hour),
			expected: domain.PolicyDefault,
		},
		{
			name:     "Día laboral tarde (18:00) -> ENERGY_SAVER",
			dateTime: monday.Add(18 * time.Hour),
			expected: domain.PolicyEnergySaver,
		},
		{
			name:     "Día laboral noche temprana (22:59) -> ENERGY_SAVER",
			dateTime: monday.Add(22*time.Hour + 59*time.Minute),
			expected: domain.PolicyEnergySaver,
		},
		{
			name:     "Día laboral noche tardía (23:00) -> ULTRA_ENERGY_SAVER",
			dateTime: monday.Add(23 * time.Hour),
			expected: domain.PolicyUltraEnergySaver,
		},
		{
			name:     "Día laboral madrugada (04:00) -> ULTRA_ENERGY_SAVER",
			dateTime: monday.Add(4 * time.Hour),
			expected: domain.PolicyUltraEnergySaver,
		},
		{
			name:     "Sábado a mediodía -> ULTRA_ENERGY_SAVER",
			dateTime: saturday,
			expected: domain.PolicyUltraEnergySaver,
		},
		{
			name:     "Domingo a mediodía -> ULTRA_ENERGY_SAVER",
			dateTime: sunday,
			expected: domain.PolicyUltraEnergySaver,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := services.GetEnergyPolicy(tt.dateTime)
			if got != tt.expected {
				t.Errorf("para %s esperaba política %v, obtuvo %v", tt.name, tt.expected, got)
			}
		})
	}
}
