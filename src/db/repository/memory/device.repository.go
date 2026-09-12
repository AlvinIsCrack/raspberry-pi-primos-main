package memory

import (
	"context"
	"sort"
	"sync"

	"primos/domain"
)

type InMemoryDeviceRepository struct {
	mu      sync.RWMutex
	devices map[domain.RoomID]domain.SensorDevice
}

// NewInMemoryDeviceRepository permite inicializar los IDs predeterminados.
func NewInMemoryDeviceRepository(initialRoomIDs ...domain.RoomID) *InMemoryDeviceRepository {
	repo := &InMemoryDeviceRepository{
		devices: make(map[domain.RoomID]domain.SensorDevice),
	}
	for _, id := range initialRoomIDs {
		repo.devices[id] = domain.NewSensorDevice(id)
	}
	return repo
}

func (r *InMemoryDeviceRepository) Get(_ context.Context, id domain.RoomID) (domain.SensorDevice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	device, exists := r.devices[id]
	if !exists {
		return domain.SensorDevice{}, domain.ErrRoomNotFound
	}
	return device, nil
}

func (r *InMemoryDeviceRepository) Update(_ context.Context, device domain.SensorDevice) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.devices[device.RoomID]; !exists {
		return domain.ErrRoomNotFound
	}
	r.devices[device.RoomID] = device
	return nil
}

func (r *InMemoryDeviceRepository) GetAll(_ context.Context) ([]domain.SensorDevice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.SensorDevice, 0, len(r.devices))
	for _, dev := range r.devices {
		result = append(result, dev)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].RoomID < result[j].RoomID
	})

	return result, nil
}
