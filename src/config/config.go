package config

import (
	"fmt"
	"os"
	"time"
)

const (
	DefaultTimezone = "America/Santiago"
	DefaultHTTPAddr = ":8080"
	DefaultUDPAddr  = ":1884"
)

var (
	Rooms = []string{"LPA", "OFI"}
)

// AppConfig almacena los ajustes globales de la aplicación.
type AppConfig struct {
	Timezone string
	HTTPAddr string
	UDPAddr  string
	Location *time.Location
}

// Load lee las variables de entorno o aplica los valores predeterminados y sincroniza time.Local.
func Load() (*AppConfig, error) {
	tz := getEnv("APP_TIMEZONE", DefaultTimezone)
	httpAddr := getEnv("HTTP_ADDR", DefaultHTTPAddr)
	udpAddr := getEnv("UDP_ADDR", DefaultUDPAddr)

	loc, err := time.LoadLocation(tz)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config: warning loading timezone '%s': %v. Falling back to UTC\n", tz, err)
		loc = time.UTC
	}

	// Sincroniza el runtime de Go a la ubicación configurada
	time.Local = loc

	return &AppConfig{
		Timezone: tz,
		HTTPAddr: httpAddr,
		UDPAddr:  udpAddr,
		Location: loc,
	}, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
