package domain

import (
	"context"
	"time"
)

// DeviceRepository desacopla el almacenamiento del ciclo de vida del servicio.
type DeviceRepository interface {
	// Get obtiene el estado persistido del dispositivo.
	Get(ctx context.Context, id RoomID) (SensorDevice, error)
	// Update muta un dispositivo registrado.
	Update(ctx context.Context, device SensorDevice) error
	// GetAll obtiene la lista completa de dispositivos registrados.
	GetAll(ctx context.Context) ([]SensorDevice, error)
}

// ScheduleRepository define la consulta de reservas para una sala.
type ScheduleRepository interface {
	// GetEventsForRoom retorna los eventos de una sala dentro de una ventana de tiempo [from, to].
	GetEventsForRoom(ctx context.Context, roomID RoomID, from, to time.Time) ([]ScheduleEvent, error)
}

// ScheduleUpdater define la capacidad de reemplazar las reservas de una sala.
type ScheduleUpdater interface {
	SetEvents(ctx context.Context, roomID RoomID, events []ScheduleEvent) error
}

// ScheduleStore compone lectura y escritura para implementaciones que soporten ambas.
type ScheduleStore interface {
	ScheduleRepository
	ScheduleUpdater
}
