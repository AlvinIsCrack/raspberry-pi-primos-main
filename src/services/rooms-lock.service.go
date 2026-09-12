package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"primos/domain"
)

// RoomsLockService orquesta la ingestión de telemetría y consultas.
type RoomsLockService struct {
	repo domain.DeviceRepository
}

// NewRoomsLockService inicializa el servicio. Si repo es nil, utiliza
// automáticamente el repositorio en memoria precargado con "LDS" y "OFI".
func NewRoomsLockService(repo domain.DeviceRepository) *RoomsLockService {
	if repo == nil {
		panic("device repository is required")
	}
	return &RoomsLockService{repo: repo}
}

// ProcessTelemetry resuelve y canaliza la telemetría entrante hacia apagado o actualización regular.
func (s *RoomsLockService) ProcessTelemetry(ctx context.Context, rawRoomID string, payload domain.TelemetryPayload, defaultShutdownReason string) error {
	roomID, err := domain.NewRoomID(rawRoomID)
	if err != nil {
		return err
	}

	if payload.Shutdown || payload.Door == domain.DoorUnknown {
		reason := payload.Reason
		if reason == "" {
			reason = defaultShutdownReason
		}
		return s.ReportShutdown(ctx, roomID, reason)
	}

	battery := -1
	if payload.BatteryLevel != nil {
		battery = *payload.BatteryLevel
	}

	return s.ReportTelemetry(ctx, roomID, domain.TelemetryReport{
		Door:         payload.Door,
		BatteryLevel: battery,
	})
}

// ReportTelemetry aplica métricas periódicas validadas únicamente a cuartos registrados.
func (s *RoomsLockService) ReportTelemetry(ctx context.Context, roomID domain.RoomID, report domain.TelemetryReport) error {
	if err := report.Validate(); err != nil {
		slog.Warn("telemetry validation failed", "room_id", roomID, "error", err)
		return fmt.Errorf("validación de telemetría falló: %w", err)
	}

	// Un solo viaje a la base de datos/memoria: Get valida la existencia de forma atómica
	device, err := s.repo.Get(ctx, roomID)
	if err != nil {
		slog.Warn("telemetry rejected: room not found", "room_id", roomID, "error", err)
		return err
	}

	device.ApplyTelemetry(report, time.Now())

	if err := s.repo.Update(ctx, device); err != nil {
		return fmt.Errorf("falló persistencia de telemetría: %w", err)
	}

	slog.Info("telemetry updated",
		"room_id", roomID,
		"door", report.Door,
		"battery", report.BatteryLevel,
	)
	return nil
}

// ReportShutdown gestiona el apagado explícito / LWT de un dispositivo existente.
func (s *RoomsLockService) ReportShutdown(ctx context.Context, roomID domain.RoomID, reason string) error {
	device, err := s.repo.Get(ctx, roomID)
	if err != nil {
		slog.Warn("sensor shutdown ignored: unregistered", "room_id", roomID)
		return err
	}

	device.ApplyShutdown(reason)

	if err := s.repo.Update(ctx, device); err != nil {
		return fmt.Errorf("falló registro de shutdown: %w", err)
	}

	slog.Info("sensor shutdown recorded", "room_id", roomID, "reason", reason)
	return nil
}

// GetSnapshot retorna el estado instantáneo proyectado del cuarto.
func (s *RoomsLockService) GetSnapshot(ctx context.Context, rawRoomID string) (domain.SensorSnapshot, bool) {
	roomID, err := domain.NewRoomID(rawRoomID)
	if err != nil {
		return domain.SensorSnapshot{}, false
	}

	device, err := s.repo.Get(ctx, roomID)
	if err != nil {
		return domain.SensorSnapshot{
			RoomID:       roomID,
			Door:         domain.DoorUnknown,
			Connectivity: domain.StatusOffline,
			BatteryLevel: -1,
		}, false
	}

	return device.Snapshot(time.Now()), true
}

// GetAllSnapshots devuelve la foto global del ecosistema ordenada de forma determinista.
func (s *RoomsLockService) GetAllSnapshots(ctx context.Context) ([]domain.SensorSnapshot, error) {
	devices, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al listar snapshots: %w", err)
	}

	now := time.Now()
	snapshots := make([]domain.SensorSnapshot, 0, len(devices))
	for _, dev := range devices {
		snapshots = append(snapshots, dev.Snapshot(now))
	}

	return snapshots, nil
}
