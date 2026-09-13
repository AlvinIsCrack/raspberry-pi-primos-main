package main

import (
	"context"
	"embed"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/lmittmann/tint"

	"primos/app"
	"primos/config"
)

const webAssetsDir = "web/build"

//go:embed all:web/build
var embeddedWebAssets embed.FS

const (
	shutdownTimeout  = 5 * time.Second
	provisionTimeout = 5 * time.Minute
)

func setupLogger(env string) {
	w := os.Stdout
	timeFormat := "2006-01-02 15:04:05.000"
	logLevel := slog.LevelDebug
	isProd := strings.EqualFold(env, "production")

	if isProd {
		timeFormat = time.RFC3339
		logLevel = slog.LevelInfo
	}

	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      logLevel,
			TimeFormat: timeFormat,
			NoColor:    isProd,
		}),
	))
}

func main() {
	// Inicializar logger inmediatamente para capturar errores de booteo/configuración
	earlyEnv := os.Getenv("APP_ENV")
	if earlyEnv == "" {
		earlyEnv = "development"
	}
	setupLogger(earlyEnv)

	// Cargar configuración ya con el logger configurado
	cfg, err := config.Load()
	if err != nil {
		slog.Error("fatal configuration error",
			"error", err,
			"raw_rooms", os.Getenv("APP_ROOMS"),
			"raw_timezone", os.Getenv("APP_TIMEZONE"),
			"app_env", earlyEnv,
		)
		os.Exit(1)
	}

	// Reajustar logger si el config cambió el entorno
	if !strings.EqualFold(earlyEnv, cfg.AppEnv) {
		setupLogger(cfg.AppEnv)
	}

	cfg.LogLoaded()
	slog.Info("booting system",
		"os", runtime.GOOS,
		"arch", runtime.GOARCH,
		"go_version", runtime.Version(),
		"pid", os.Getpid(),
	)

	slog.Info("booting system",
		"os", runtime.GOOS,
		"arch", runtime.GOARCH,
		"go_version", runtime.Version(),
		"pid", os.Getpid(),
	)

	// Inicializar la aplicación centralizada
	application, err := app.New(cfg, embeddedWebAssets, webAssetsDir)
	if err != nil {
		slog.Error("application failed to initialize", "error", err)
		os.Exit(1)
	}
	if err := application.Start(); err != nil {
		slog.Error("application failed to start", "error", err)
		os.Exit(1)
	}

	// Esperar señal de apagado
	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, os.Interrupt, syscall.SIGTERM)
	sig := <-shutdownSig
	slog.Info("shutdown signal received", "signal", sig.String())

	// Apagado grácil
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := application.Shutdown(ctx); err != nil {
		slog.Error("application graceful shutdown encountered errors", "error", err)
	} else {
		slog.Info("application shutdown completed successfully")
	}
}
