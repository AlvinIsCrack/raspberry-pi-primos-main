package services

import (
	"errors"
	"fmt"

	"primos/config"
	"primos/db/repository/memory"
)

// Container agrupa los servicios de negocio instanciados.
type Container struct {
	RoomsLock *RoomsLockService
	Schedule  *RoomsScheduleService
}

// Bootstrap construye todos los repositorios y servicios centrales en un solo paso.
func Bootstrap(cfg *config.AppConfig) (*Container, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}

	// Repositorios: se pasan directamente las salas tipadas y validadas
	deviceRepo := memory.NewInMemoryDeviceRepository(cfg.Rooms...)
	scheduleRepo := memory.NewInMemoryScheduleRepository()

	// Servicios
	lockService := NewRoomsLockService(deviceRepo)
	scheduleService, err := NewRoomsScheduleServiceFromStore(scheduleRepo)
	if err != nil {
		return nil, fmt.Errorf("rooms schedule: %w", err)
	}

	return &Container{
		RoomsLock: lockService,
		Schedule:  scheduleService,
	}, nil
}
