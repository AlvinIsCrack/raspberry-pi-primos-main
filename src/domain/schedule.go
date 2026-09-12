package domain

import (
	"time"
)

// ScheduleEvent representa una reserva o bloque de ocupación de una sala.
type ScheduleEvent struct {
	ID        string    `json:"id"`
	RoomID    RoomID    `json:"room_id"`
	Summary   string    `json:"summary"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// IsActive determina si el evento está en curso para un instante temporal dado.
// Consideramos el intervalo semiabierto [StartTime, EndTime) o cerrado según la política.
func (e ScheduleEvent) IsActive(t time.Time) bool {
	return (t.Equal(e.StartTime) || t.After(e.StartTime)) && t.Before(e.EndTime)
}

// RoomOccupancyKind indica si la sala tiene actividad planificada o no.
type RoomOccupancyKind string

const (
	OccupancyFree RoomOccupancyKind = "FREE"
	OccupancyBusy RoomOccupancyKind = "BUSY"
)

// RoomScheduleStatus resume el estado temporal de una sala en un instante T.
type RoomScheduleStatus struct {
	RoomID      RoomID            `json:"room_id"`
	Occupancy   RoomOccupancyKind `json:"occupancy"`
	ActiveEvent *ScheduleEvent    `json:"active_event,omitempty"`
	NextEvent   *ScheduleEvent    `json:"next_event,omitempty"`
	DesiredDoor DoorState         `json:"desired_door"`
	EvaluatedAt time.Time         `json:"evaluated_at"`
}

// ResolveDesiredDoorState calcula el estado que debería tener la puerta
// según la ocupación actual de la sala.
func ResolveDesiredDoorState(occupancy RoomOccupancyKind) DoorState {
	switch occupancy {
	case OccupancyBusy:
		return DoorOpen // Si la sala está reservada/en clase, debe estar abierta
	case OccupancyFree:
		return DoorClosed // Si no hay reserva planificada, debe permanecer cerrada/bloqueada
	default:
		return DoorUnknown
	}
}

// EvaluateRoomSchedule calcula la proyección de estado temporal de una sala
// a partir de sus eventos y un instante de referencia.
func EvaluateRoomSchedule(roomID RoomID, events []ScheduleEvent, now time.Time) RoomScheduleStatus {
	var active *ScheduleEvent
	var next *ScheduleEvent

	for i := range events {
		ev := &events[i]

		if ev.IsActive(now) {
			// Si hay eventos solapados, priorizar el que termine más tarde
			if active == nil || ev.EndTime.After(active.EndTime) {
				active = ev
			}
			continue
		}

		if ev.StartTime.After(now) {
			// Priorizar el siguiente evento más próximo en el tiempo
			if next == nil || ev.StartTime.Before(next.StartTime) {
				next = ev
			}
		}
	}

	occupancy := OccupancyFree
	if active != nil {
		occupancy = OccupancyBusy
	}

	return RoomScheduleStatus{
		RoomID:      roomID,
		Occupancy:   occupancy,
		ActiveEvent: active,
		NextEvent:   next,
		DesiredDoor: ResolveDesiredDoorState(occupancy),
		EvaluatedAt: now,
	}
}
