package domain_test

import (
	"testing"
	"time"

	"primos/domain"
)

func TestEvaluateRoomSchedule(t *testing.T) {
	roomID := domain.RoomID("LPA")
	refTime := time.Date(2026, 9, 14, 10, 0, 0, 0, time.UTC)

	t.Run("Sin eventos registrados debe estar libre y puerta cerrada", func(t *testing.T) {
		status := domain.EvaluateRoomSchedule(roomID, nil, refTime)

		if status.Occupancy != domain.OccupancyFree {
			t.Errorf("esperaba ocupación %s, obtuvo %s", domain.OccupancyFree, status.Occupancy)
		}
		if status.DesiredDoor != domain.DoorClosed {
			t.Errorf("esperaba puerta %s, obtuvo %s", domain.DoorClosed, status.DesiredDoor)
		}
		if status.ActiveEvent != nil {
			t.Errorf("no esperaba evento activo, obtuvo %+v", status.ActiveEvent)
		}
	})

	t.Run("Con evento en curso debe estar ocupada y puerta abierta", func(t *testing.T) {
		events := []domain.ScheduleEvent{
			{
				ID:        "ev-1",
				RoomID:    roomID,
				Summary:   "Clase Redes",
				StartTime: refTime.Add(-30 * time.Minute),
				EndTime:   refTime.Add(30 * time.Minute),
			},
			{
				ID:        "ev-2",
				RoomID:    roomID,
				Summary:   "Clase Sistemas",
				StartTime: refTime.Add(1 * time.Hour),
				EndTime:   refTime.Add(2 * time.Hour),
			},
		}

		status := domain.EvaluateRoomSchedule(roomID, events, refTime)

		if status.Occupancy != domain.OccupancyBusy {
			t.Errorf("esperaba ocupación %s, obtuvo %s", domain.OccupancyBusy, status.Occupancy)
		}
		if status.DesiredDoor != domain.DoorOpen {
			t.Errorf("esperaba puerta %s, obtuvo %s", domain.DoorOpen, status.DesiredDoor)
		}
		if status.ActiveEvent == nil || status.ActiveEvent.ID != "ev-1" {
			t.Errorf("esperaba evento activo ev-1, obtuvo %+v", status.ActiveEvent)
		}
		if status.NextEvent == nil || status.NextEvent.ID != "ev-2" {
			t.Errorf("esperaba próximo evento ev-2, obtuvo %+v", status.NextEvent)
		}
	})

	t.Run("Entre eventos sin evento en curso debe estar libre y retener NextEvent", func(t *testing.T) {
		events := []domain.ScheduleEvent{
			{
				ID:        "ev-past",
				RoomID:    roomID,
				Summary:   "Clase Pasada",
				StartTime: refTime.Add(-2 * time.Hour),
				EndTime:   refTime.Add(-1 * time.Hour),
			},
			{
				ID:        "ev-future",
				RoomID:    roomID,
				Summary:   "Clase Futura",
				StartTime: refTime.Add(30 * time.Minute),
				EndTime:   refTime.Add(90 * time.Minute),
			},
		}

		status := domain.EvaluateRoomSchedule(roomID, events, refTime)

		if status.Occupancy != domain.OccupancyFree {
			t.Errorf("esperaba ocupación %s, obtuvo %s", domain.OccupancyFree, status.Occupancy)
		}
		if status.DesiredDoor != domain.DoorClosed {
			t.Errorf("esperaba puerta %s, obtuvo %s", domain.DoorClosed, status.DesiredDoor)
		}
		if status.ActiveEvent != nil {
			t.Errorf("no esperaba evento activo, obtuvo %+v", status.ActiveEvent)
		}
		if status.NextEvent == nil || status.NextEvent.ID != "ev-future" {
			t.Errorf("esperaba próximo evento ev-future, obtuvo %+v", status.NextEvent)
		}
	})
}
