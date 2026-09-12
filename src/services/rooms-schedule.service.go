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
	ErrNilScheduleRepository = errors.New("schedule repository is required")
	ErrNilScheduleUpdater    = errors.New("schedule updater is required")
)

// RoomsScheduleService orquesta la consulta y sincronización de horarios de salas.
type RoomsScheduleService struct {
	reader  domain.ScheduleRepository
	updater domain.ScheduleUpdater
}

// NewRoomsScheduleService inicializa el servicio validando sus dependencias sin panic.
func NewRoomsScheduleService(reader domain.ScheduleRepository, updater domain.ScheduleUpdater) (*RoomsScheduleService, error) {
	if reader == nil {
		return nil, ErrNilScheduleRepository
	}
	if updater == nil {
		return nil, ErrNilScheduleUpdater
	}
	return &RoomsScheduleService{
		reader:  reader,
		updater: updater,
	}, nil
}

// GetRoomScheduleStatus calcula la ocupación y la puerta deseada para un instante dado.
func (s *RoomsScheduleService) GetRoomScheduleStatus(ctx context.Context, rawRoomID string, now time.Time) (domain.RoomScheduleStatus, error) {
	roomID, err := domain.NewRoomID(rawRoomID)
	if err != nil {
		return domain.RoomScheduleStatus{}, err
	}

	from := now.Add(-24 * time.Hour)
	to := now.Add(24 * time.Hour)

	events, err := s.reader.GetEventsForRoom(ctx, roomID, from, to)
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

	if err := s.updater.SetEvents(ctx, roomID, events); err != nil {
		slog.Error("failed to set room schedule events", "room_id", roomID, "error", err)
		return fmt.Errorf("persist room events failed: %w", err)
	}

	slog.Info("room schedule synced successfully", "room_id", roomID, "events_count", len(events))
	return nil
}
