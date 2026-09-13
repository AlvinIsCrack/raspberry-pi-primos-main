package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"primos/domain"
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

// LogLoaded inspecciona y registra todos los campos de AppConfig de manera dinámica.
func (c *AppConfig) LogLoaded() {
	slog.Info("configuration loaded",
		"timezone", c.Timezone,
		"http_addr", c.HTTPAddr,
		"udp_addr", c.UDPAddr,
		"location", c.Location.String(),
		"rooms", c.Rooms,
	)
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

	cfg := &AppConfig{
		Timezone: tz,
		HTTPAddr: httpAddr,
		UDPAddr:  udpAddr,
		Location: loc,
		Rooms:    rooms,
	}
	cfg.LogLoaded()

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
