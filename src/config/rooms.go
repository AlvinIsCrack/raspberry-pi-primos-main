package config

import (
	"fmt"
	"primos/domain"
	"strings"
)

func parseRooms(raw string) ([]domain.RoomID, error) {
	parts := strings.Split(raw, ",")
	rooms := make([]domain.RoomID, 0, len(parts))
	seen := make(map[domain.RoomID]struct{}, len(parts))

	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		roomID, err := domain.NewRoomID(trimmed)
		if err != nil {
			return nil, fmt.Errorf("room %q: %w", trimmed, err)
		}
		if _, exists := seen[roomID]; exists {
			return nil, fmt.Errorf("duplicate room %q", roomID)
		}
		seen[roomID] = struct{}{}
		rooms = append(rooms, roomID)
	}

	if len(rooms) == 0 {
		return nil, fmt.Errorf("at least one room must be configured")
	}

	return rooms, nil
}
