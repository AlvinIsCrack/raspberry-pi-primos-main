package config

import (
	"fmt"
	"os"
	"time"
)

const (
	DefaultTimezone = "America/Santiago"
	DefaultHTTPAddr = ":8080"
	DefaultMQTTAddr = ":1883"
)

// AppConfig almacena los ajustes globales de la aplicación.
type AppConfig struct {
	Timezone string
	HTTPAddr string
	MQTTAddr string
	Location *time.Location
}

// Load lee las variables de entorno o aplica los valores predeterminados y sincroniza time.Local.
func Load() (*AppConfig, error) {
	tz := getEnv("APP_TIMEZONE", DefaultTimezone)
	httpAddr := getEnv("HTTP_ADDR", DefaultHTTPAddr)
	mqttAddr := getEnv("MQTT_ADDR", DefaultMQTTAddr)

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
		MQTTAddr: mqttAddr,
		Location: loc,
	}, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
