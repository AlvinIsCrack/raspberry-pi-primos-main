package services

import (
	"primos/domain"
	"time"
)

func GetEnergyPolicy(t time.Time) domain.EnergyPolicy {
	weekday := t.Weekday()
	hour := t.Hour()

	if weekday == time.Saturday || weekday == time.Sunday || hour >= 23 || hour < 7 {
		return domain.PolicyUltraEnergySaver
	}
	if hour >= 18 && hour < 23 {
		return domain.PolicyEnergySaver
	}
	return domain.PolicyDefault
}

// GetCurrentEnergyPolicy calcula el modo según la hora y día local
func GetCurrentEnergyPolicy() domain.EnergyPolicy {
	return GetEnergyPolicy(time.Now())
}
