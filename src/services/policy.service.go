package services

import "time"

// Modos de política de energía para los microcontroladores:
// 1 = Normal / Defecto (Lunes a Viernes de 07:00 a 20:00)
// 2 = Energy Saver (Lunes a Viernes de 20:00 a 23:00)
// 3 = Ultra Energy Saver (Noches 23:00 a 07:00 y Fines de semana)
const (
	PolicyDefault          = 1
	PolicyEnergySaver      = 2
	PolicyUltraEnergySaver = 3
)

// GetCurrentEnergyPolicy calcula el modo según la hora y día local
func GetCurrentEnergyPolicy() int {
	now := time.Now()
	weekday := now.Weekday()
	hour := now.Hour()

	// Fin de semana completo: Sábado o Domingo -> Ultra Energy Saver
	if weekday == time.Saturday || weekday == time.Sunday {
		return PolicyUltraEnergySaver
	}

	// Madrugada / Noche profunda -> Ultra Energy Saver
	if hour >= 23 || hour < 7 {
		return PolicyUltraEnergySaver
	}

	// Tarde / Noche -> Energy Saver
	if hour >= 18 && hour < 23 {
		return PolicyEnergySaver
	}

	// Horario laboral activo -> Defecto
	return PolicyDefault
}
