package main

import (
	"context"
	"embed"
	"fmt"
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
		fmt.Fprintf(os.Stderr, "Fatal configuration error: %v\n", err)
		os.Exit(1)
	}

	// Inicializar la aplicación centralizada
	application := app.New(cfg, embeddedWebAssets, webAssetsDir)
	if err := application.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Esperar señal de apagado
	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, os.Interrupt, syscall.SIGTERM)
	<-shutdownSig

	// Apagado grácil
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	application.Shutdown(ctx)
}
