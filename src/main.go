package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"primos/api"
	"primos/config"
	"primos/core"
	"primos/services"
)

//go:embed all:web/dist
var embeddedWebAssets embed.FS

const (
	shutdownTimeout  = 5 * time.Second
	provisionTimeout = 5 * time.Minute
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal configuration error: %v\n", err)
		os.Exit(1)
	}

	lockService := services.NewRoomsLockService()

	apiRouter := api.BuildDefaultRouter(api.AppServices{
		RoomsLock: lockService,
	})

	mqttBroker := core.NewMQTTBroker(cfg.MQTTAddr)
	apiRouter.RegisterMQTTRoutes(mqttBroker)

	go func() {
		if err := mqttBroker.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "MQTT Server error: %v\n", err)
		}
	}()

	httpMux := http.NewServeMux()
	apiRouter.RegisterHTTPRoutes(httpMux)

	distSubtree, subErr := fs.Sub(embeddedWebAssets, "web/dist")
	if subErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to resolve embedded web directory: %v\n", subErr)
		os.Exit(1)
	}

	// Single-page application handler with index.html root resolution
	fileServer := http.FileServer(http.FS(distSubtree))
	httpMux.Handle("/", fileServer)

	httpServer := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: httpMux,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "HTTP Server error: %v\n", err)
		}
	}()

	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, os.Interrupt, syscall.SIGTERM)
	<-shutdownSig

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "HTTP graceful shutdown failed: %v\n", err)
	}

	mqttBroker.Stop()
}
