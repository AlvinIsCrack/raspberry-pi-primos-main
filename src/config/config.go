package config

import (
	"fmt"
	"os"
	"time"
)

const (
	DefaultTimezone   = "America/Santiago"
	DefaultHTTPAddr   = ":8080"
	DefaultMQTTAddr   = ":1883"
	DefaultGithubUser = "AlvinIsCrack"
	DefaultGithubRepo = "raspberry-pi-primos-main"
)

// AppConfig almacena los ajustes globales de la aplicación.
type AppConfig struct {
	Timezone   string
	HTTPAddr   string
	MQTTAddr   string
	Location   *time.Location
	GithubUser string
	GithubRepo string
}

// Load lee las variables de entorno o aplica los valores predeterminados y sincroniza time.Local.
func Load() (*AppConfig, error) {
	tz := getEnv("APP_TIMEZONE", DefaultTimezone)

	loc, err := time.LoadLocation(tz)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config: warning loading timezone '%s': %v. Falling back to UTC\n", tz, err)
		loc = time.UTC
	}

	// Sincroniza el runtime de Go a la ubicación configurada
	time.Local = loc

	return &AppConfig{
		Timezone:   tz,
		HTTPAddr:   getEnv("HTTP_ADDR", DefaultHTTPAddr),
		MQTTAddr:   getEnv("MQTT_ADDR", DefaultMQTTAddr),
		Location:   loc,
		GithubUser: getEnv("GITHUB_USER", DefaultGithubUser),
		GithubRepo: getEnv("GITHUB_REPO", DefaultGithubRepo),
	}, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
