package domain

import "context"

// DeviceRepository desacopla el almacenamiento del ciclo de vida del servicio.
type DeviceRepository interface {
	// Get obtiene el estado persistido del dispositivo.
	Get(ctx context.Context, id RoomID) (SensorDevice, error)
	// Update muta un dispositivo registrado.
	Update(ctx context.Context, device SensorDevice) error
	// GetAll obtiene la lista completa de dispositivos registrados.
	GetAll(ctx context.Context) ([]SensorDevice, error)
}
