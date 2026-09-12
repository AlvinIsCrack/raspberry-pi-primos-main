package main

import (
	"context"
	"embed"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

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

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("fatal configuration error", "error", err)
		os.Exit(1)
	}

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
