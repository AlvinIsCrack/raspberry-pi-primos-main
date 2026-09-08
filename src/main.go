package main

import (
	"context"
	"errors"
	"fmt"
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

	// Automatic kiosk verification and provisioning
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), provisionTimeout)
		defer cancel()

		_, port, err := net.SplitHostPort(cfg.HTTPAddr)
		if err != nil {
			port = strings.TrimPrefix(cfg.HTTPAddr, ":")
		}
		targetURL := fmt.Sprintf("http://127.0.0.1:%s", port)
		provisioner := system.NewProvisioner(targetURL)

		if err := provisioner.AutoProvision(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Kiosk self-healing check warning: %v\n", err)
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
