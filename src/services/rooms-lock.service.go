package services

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"primos/domain"
)

// SensorDevice mantiene el estado interno y telemetría de un nodo IoT.
type SensorDevice struct {
	RoomID        string
	LastReport    domain.TelemetryReport
	LastSeenAt    time.Time
	IsExplicitOff bool
	ShutdownNote  string
}

// RoomsLockService gestiona la telemetría y ciclo de vida de los sensores de salas.
type RoomsLockService struct {
	mu      sync.RWMutex
	devices map[string]SensorDevice
}

// NewRoomsLockService inicializa el servicio con los cuartos por defecto.
func NewRoomsLockService() *RoomsLockService {
	s := &RoomsLockService{
		devices: make(map[string]SensorDevice),
	}

	// Sensores predeterminados
	for _, id := range []string{"LDS", "OFI"} {
		s.devices[id] = SensorDevice{
			RoomID: id,
			LastReport: domain.TelemetryReport{
				Door:         domain.DoorUnknown,
				BatteryLevel: -1,
			},
			LastSeenAt:    time.Time{},
			IsExplicitOff: true,
			ShutdownNote:  "Inicialización por defecto",
		}
	}

	return s
}

// ResetTelemetry retorna un reporte limpio en estado desconocido y sin métricas.
func ResetTelemetry() domain.TelemetryReport {
	return domain.TelemetryReport{
		Door:         domain.DoorUnknown,
		BatteryLevel: -1,
		IsCharging:   false,
		RSSI:         0,
	}
}

func (s *RoomsLockService) ProcessTelemetry(roomID string, payload domain.TelemetryPayload, defaultShutdownReason string) error {
	if len(roomID) != 3 {
		return errors.New("el ID del cuarto debe tener exactamente 3 caracteres")
	}

	if payload.Shutdown || payload.Door == domain.DoorUnknown {
		reason := payload.Reason
		if reason == "" {
			reason = defaultShutdownReason
		}
		return s.ReportShutdown(roomID, reason)
	}

	battery := -1
	if payload.BatteryLevel != nil {
		battery = *payload.BatteryLevel
	}

	return s.ReportTelemetry(roomID, domain.TelemetryReport{
		Door:         payload.Door,
		BatteryLevel: battery,
		IsCharging:   payload.IsCharging,
		RSSI:         payload.RSSI,
	})
}

// ReportTelemetry procesa una ráfaga periódica de datos del sensor.
func (s *RoomsLockService) ReportTelemetry(roomID string, report domain.TelemetryReport) error {
	if len(roomID) != 3 {
		return errors.New("el ID del cuarto debe tener exactamente 3 caracteres")
	}
	if err := report.Validate(); err != nil {
		return fmt.Errorf("validación de telemetría falló: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.devices[roomID] = SensorDevice{
		RoomID:        roomID,
		LastReport:    report,
		LastSeenAt:    time.Now(),
		IsExplicitOff: false,
		ShutdownNote:  "",
	}
	return nil
}

// ReportShutdown atiende eventos LWT (Last Will and Testament) o apagado voluntario del sensor.
// Fuerza el estado a desconocido de inmediato sin esperar el timeout de 5 minutos.
func (s *RoomsLockService) ReportShutdown(roomID string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	device, exists := s.devices[roomID]
	if !exists {
		return errors.New("sensor no registrado")
	}

	device.IsExplicitOff = true
	device.ShutdownNote = reason
	device.LastReport = ResetTelemetry()
	s.devices[roomID] = device

	return nil
}

// GetSnapshot calcula el estado actual del dispositivo y valida si expiró la ventana de gracia.
func (s *RoomsLockService) GetSnapshot(roomID string) (domain.SensorSnapshot, bool) {
	s.mu.RLock()
	device, exists := s.devices[roomID]
	s.mu.RUnlock()

	if !exists {
		return domain.SensorSnapshot{
			RoomID:       roomID,
			Door:         domain.DoorUnknown,
			Connectivity: domain.StatusOffline,
			BatteryLevel: -1,
		}, false
	}

	now := time.Now()
	isExpired := now.Sub(device.LastSeenAt) > domain.HeartbeatGracePeriod
	isOffline := device.IsExplicitOff || isExpired

	doorState := device.LastReport.Door
	connectivity := domain.StatusOnline
	batteryLevel := device.LastReport.BatteryLevel

	if isOffline || doorState == domain.DoorUnknown {
		doorState = domain.DoorUnknown
		connectivity = domain.StatusOffline
		batteryLevel = -1
	}

	return domain.SensorSnapshot{
		RoomID:       device.RoomID,
		Door:         doorState,
		Connectivity: connectivity,
		BatteryLevel: batteryLevel,
		LastSeenAt:   device.LastSeenAt,
		IsStale:      isExpired,
	}, true
}

// GetAllSnapshots devuelve todos los snapshots calculados y ordenados por RoomID.
func (s *RoomsLockService) GetAllSnapshots() []domain.SensorSnapshot {
	s.mu.RLock()
	ids := make([]string, 0, len(s.devices))
	for id := range s.devices {
		ids = append(ids, id)
	}
	s.mu.RUnlock()

	sort.Strings(ids)

	snapshots := make([]domain.SensorSnapshot, 0, len(ids))
	for _, id := range ids {
		snap, _ := s.GetSnapshot(id)
		snapshots = append(snapshots, snap)
	}

	return snapshots
}
