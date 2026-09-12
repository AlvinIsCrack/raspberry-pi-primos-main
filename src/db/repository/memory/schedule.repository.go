package memory

import (
	"context"
	"slices"
	"sync"
	"time"

	"primos/domain"
)

// InMemoryScheduleRepository almacena y consulta eventos de calendario en memoria de forma concurrente.
type InMemoryScheduleRepository struct {
	mu     sync.RWMutex
	events map[domain.RoomID][]domain.ScheduleEvent
}

// NewInMemoryScheduleRepository inicializa un repositorio de calendario en memoria vacío o precargado.
func NewInMemoryScheduleRepository() *InMemoryScheduleRepository {
	return &InMemoryScheduleRepository{
		events: make(map[domain.RoomID][]domain.ScheduleEvent),
	}
}

// SetEvents reemplaza atómicamente todos los eventos de una sala dada.
func (r *InMemoryScheduleRepository) SetEvents(_ context.Context, roomID domain.RoomID, newEvents []domain.ScheduleEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := make([]domain.ScheduleEvent, len(newEvents))
	copy(copied, newEvents)

	slices.SortFunc(copied, func(a, b domain.ScheduleEvent) int {
		return a.StartTime.Compare(b.StartTime)
	})

	r.events[roomID] = copied
	return nil
}

// GetEventsForRoom busca los eventos que se solapan con la ventana [from, to].
func (r *InMemoryScheduleRepository) GetEventsForRoom(_ context.Context, roomID domain.RoomID, from, to time.Time) ([]domain.ScheduleEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	roomEvents, exists := r.events[roomID]
	if !exists || len(roomEvents) == 0 {
		return []domain.ScheduleEvent{}, nil
	}

	// Early exit general: si el rango solicitado termina antes de que comience el primer evento
	if to.Before(roomEvents[0].StartTime) {
		return []domain.ScheduleEvent{}, nil
	}

	// Búsqueda binaria para saltar eventos iniciales que terminan antes de 'from'.
	// Buscamos el primer índice cuyo StartTime se aproxime a 'from'.
	idx, _ := slices.BinarySearchFunc(roomEvents, from, func(e domain.ScheduleEvent, target time.Time) int {
		return e.StartTime.Compare(target)
	})

	// Retrocedemos si hay eventos anteriores que pudieron haber empezado antes de 'from'
	// pero que continúan vigentes dentro de la ventana (ej. eventos de larga duración).
	startIdx := idx
	for startIdx > 0 && roomEvents[startIdx-1].EndTime.After(from) {
		startIdx--
	}

	result := make([]domain.ScheduleEvent, 0, len(roomEvents)-startIdx)
	for i := startIdx; i < len(roomEvents); i++ {
		ev := roomEvents[i]

		// Early exit: dado que roomEvents está ordenado por StartTime,
		// si ev.StartTime > to, ningún evento posterior puede intersecar [from, to].
		if ev.StartTime.After(to) {
			break
		}

		// Condición de solapamiento: ev.StartTime <= to && ev.EndTime >= from
		if (ev.StartTime.Before(to) || ev.StartTime.Equal(to)) &&
			(ev.EndTime.After(from) || ev.EndTime.Equal(from)) {
			result = append(result, ev)
		}
	}

	return result, nil
}
