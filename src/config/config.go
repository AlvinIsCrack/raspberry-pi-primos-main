package config

import (
	"fmt"
	"os"
	"primos/domain"
	"strings"
	"time"
)

const (
	DefaultTimezone = "America/Santiago"
	DefaultHTTPAddr = ":8080"
	DefaultUDPAddr  = ":1884"
	DefaultRooms    = "LPA,OFI"
)

// AppConfig almacena los ajustes globales de la aplicación.
type AppConfig struct {
	Timezone string
	HTTPAddr string
	UDPAddr  string
	Location *time.Location
	Rooms    []domain.RoomID
}

// Load lee las variables de entorno o aplica los valores predeterminados y sincroniza time.Local.
func Load() (*AppConfig, error) {
	tz := getEnv("APP_TIMEZONE", DefaultTimezone)
	httpAddr := getEnv("HTTP_ADDR", DefaultHTTPAddr)
	udpAddr := getEnv("UDP_ADDR", DefaultUDPAddr)
	roomsRaw := getEnv("APP_ROOMS", DefaultRooms)

	loc, err := time.LoadLocation(tz)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config: warning loading timezone '%s': %v. Falling back to UTC\n", tz, err)
		loc = time.UTC
	}
	// Sincroniza el runtime de Go a la ubicación configurada
	time.Local = loc

	// Parsear y validar las salas configuradas
	rooms, err := parseRooms(roomsRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid rooms configuration: %w", err)
	}

	return &AppConfig{
		Timezone: tz,
		HTTPAddr: httpAddr,
		UDPAddr:  udpAddr,
		Location: loc,
		Rooms:    rooms,
	}, nil
}

func parseRooms(raw string) ([]domain.RoomID, error) {
	parts := strings.Split(raw, ",")
	rooms := make([]domain.RoomID, 0, len(parts))

	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		roomID, err := domain.NewRoomID(trimmed)
		if err != nil {
			return nil, fmt.Errorf("room %q: %w", trimmed, err)
		}
		rooms = append(rooms, roomID)
	}

	if len(rooms) == 0 {
		return nil, fmt.Errorf("at least one room must be configured")
	}

	return rooms, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
