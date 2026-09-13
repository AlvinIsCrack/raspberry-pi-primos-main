package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"primos/domain"
)

const (
	DefaultTimezone = "America/Santiago"
	DefaultHTTPAddr = ":8080"
	DefaultUDPAddr  = ":1884"
	DefaultRooms    = "LPA,OFI"
	DefaultAppEnv   = "development"
)

// AppConfig almacena los ajustes globales de la aplicación.
type AppConfig struct {
	Timezone string
	HTTPAddr string
	UDPAddr  string
	Location *time.Location
	Rooms    []domain.RoomID
	AppEnv   string
}

// LogLoaded inspecciona y registra todos los campos de AppConfig de manera dinámica.
func (c *AppConfig) LogLoaded() {
	slog.Info("configuration loaded",
		"timezone", c.Timezone,
		"http_addr", c.HTTPAddr,
		"udp_addr", c.UDPAddr,
		"location", c.Location.String(),
		"rooms", c.Rooms,
		"app_env", c.AppEnv,
	)
}

// Load lee las variables de entorno o aplica los valores predeterminados y sincroniza time.Local.
func Load() (*AppConfig, error) {
	tz := getEnv("APP_TIMEZONE", DefaultTimezone)
	httpAddr := getEnv("HTTP_ADDR", DefaultHTTPAddr)
	udpAddr := getEnv("UDP_ADDR", DefaultUDPAddr)
	roomsRaw := getEnv("APP_ROOMS", DefaultRooms)
	appEnv := getEnv("APP_ENV", DefaultAppEnv)

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q: %w", tz, err)
	}
	// Sincroniza el runtime de Go a la ubicación configurada
	time.Local = loc

	// Parsear y validar las salas configuradas
	rooms, err := parseRooms(roomsRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid rooms configuration %q: %w", roomsRaw, err)
	}

	return &AppConfig{
		Timezone: tz,
		HTTPAddr: httpAddr,
		UDPAddr:  udpAddr,
		Location: loc,
		Rooms:    rooms,
		AppEnv:   appEnv,
	}, nil
}

func (c *AppConfig) IsProduction() bool {
	return strings.EqualFold(c.AppEnv, "production")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
