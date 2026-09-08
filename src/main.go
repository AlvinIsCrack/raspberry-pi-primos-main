package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"primos/api"
	"primos/config"
	"primos/core"
	"primos/services"
	"primos/system"
)

// Embedded static web application assets
//
//go:embed web/dist/*
var webAssets embed.FS

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

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), provisionTimeout)
		defer cancel()

		_, port, splitErr := net.SplitHostPort(cfg.HTTPAddr)
		if splitErr != nil {
			port = strings.TrimPrefix(cfg.HTTPAddr, ":")
		}
		targetURL := fmt.Sprintf("http://127.0.0.1:%s", port)
		provisioner := system.NewProvisioner(targetURL)

		if err := provisioner.AutoProvision(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Kiosk provisioning warning: %v\n", err)
		}
	}()

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

	// Isolate sub-tree to serve root index safely
	distFS, fsErr := fs.Sub(webAssets, "web/dist")
	if fsErr != nil {
		fmt.Fprintf(os.Stderr, "Static assets initialization failed: %v\n", fsErr)
		os.Exit(1)
	}
	httpMux.Handle("/", http.FileServer(http.FS(distFS)))

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
