package services

import (
	"fmt"

	"primos/config"
	"primos/db/repository/memory"
	"primos/domain"
)

// Container agrupa los servicios de negocio instanciados.
type Container struct {
	RoomsLock *RoomsLockService
	Schedule  *RoomsScheduleService
}

// Bootstrap construye todos los repositorios y servicios centrales en un solo paso.
func Bootstrap(cfg *config.AppConfig) (*Container, error) {
	// Repositorios
	roomIDs := make([]domain.RoomID, len(cfg.Rooms))
	for i, r := range cfg.Rooms {
		roomIDs[i] = domain.RoomID(r)
	}
	deviceRepo := memory.NewInMemoryDeviceRepository(roomIDs...)
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
