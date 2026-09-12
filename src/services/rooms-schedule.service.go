package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"primos/domain"
)

var (
	ErrNilScheduleStore = errors.New("schedule store is required")
)

// RoomsScheduleService orquesta la consulta y sincronización de horarios de salas.
type RoomsScheduleService struct {
	store domain.ScheduleStore
}

// NewRoomsScheduleService inicializa el servicio validando sus dependencias sin recurrir a panic.
func NewRoomsScheduleService(store domain.ScheduleStore) (*RoomsScheduleService, error) {
	if store == nil {
		return nil, ErrNilScheduleStore
	}
	return &RoomsScheduleService{store: store}, nil
}

// GetRoomScheduleStatus calcula la ocupación y la puerta deseada para un instante dado.
func (s *RoomsScheduleService) GetRoomScheduleStatus(ctx context.Context, rawRoomID string, now time.Time) (domain.RoomScheduleStatus, error) {
	roomID, err := domain.NewRoomID(rawRoomID)
	if err != nil {
		return domain.RoomScheduleStatus{}, err
	}

	// Ventana segura: 24h atrás para eventos continuos extensos y 24h adelante para detectar el siguiente bloque.
	from := now.Add(-24 * time.Hour)
	to := now.Add(24 * time.Hour)

	events, err := s.store.GetEventsForRoom(ctx, roomID, from, to)
	if err != nil {
		slog.Error("failed to retrieve room schedule events", "room_id", roomID, "error", err)
		return domain.RoomScheduleStatus{}, fmt.Errorf("retrieve room events failed: %w", err)
	}

	return domain.EvaluateRoomSchedule(roomID, events, now), nil
}

// SyncRoomEvents actualiza y persiste de manera segura las reservas asociadas a una sala.
func (s *RoomsScheduleService) SyncRoomEvents(ctx context.Context, rawRoomID string, events []domain.ScheduleEvent) error {
	roomID, err := domain.NewRoomID(rawRoomID)
	if err != nil {
		return err
	}

	if err := s.store.SetEvents(ctx, roomID, events); err != nil {
		slog.Error("failed to set room schedule events", "room_id", roomID, "error", err)
		return fmt.Errorf("persist room events failed: %w", err)
	}

	slog.Info("room schedule synced successfully", "room_id", roomID, "events_count", len(events))
	return nil
}
