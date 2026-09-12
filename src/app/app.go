package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"primos/api"
	"primos/config"
	"primos/services"
)

type App struct {
	runners []ServiceRunner
}

func New(cfg *config.AppConfig, webAssets fs.FS, webAssetsDir string) (*App, error) {
	// Inicialización unificada de dependencias de negocio
	svcs, err := services.Bootstrap(cfg)
	if err != nil {
		return nil, fmt.Errorf("services bootstrap failed: %w", err)
	}

	// Controladores y Mux
	endpoints := api.BuildEndpoints(api.AppServices{
		RoomsLock:     svcs.RoomsLock,
		RoomsSchedule: svcs.Schedule,
	})

	httpMux := http.NewServeMux()
	endpoints.Router.RegisterHTTPRoutes(httpMux)
	if err := setupSPAFallback(httpMux, webAssets, webAssetsDir); err != nil {
		return nil, fmt.Errorf("setup spa fallback: %w", err)
	}

	// Declaración unificada de tareas ejecutables
	runners := []ServiceRunner{
		&UDPControllerRunner{ctrl: endpoints.UDPController, addr: cfg.UDPAddr},
		&HTTPServerRunner{server: &http.Server{Addr: cfg.HTTPAddr, Handler: httpMux}, addr: cfg.HTTPAddr},
	}

	return &App{runners: runners}, nil
}

func (a *App) Start() error {
	for i, r := range a.runners {
		if err := r.Start(); err != nil {
			// Roll back previously started runners in reverse order
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			for j := i - 1; j >= 0; j-- {
				_ = a.runners[j].Shutdown(shutdownCtx)
			}
			return fmt.Errorf("runner %s failed to start: %w", r.Name(), err)
		}
		slog.Info("component started", "name", r.Name())
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	slog.Info("initiating graceful shutdown")
	var errs []error

	// Apagado en orden inverso al de encendido (LIFO)
	for i := len(a.runners) - 1; i >= 0; i-- {
		r := a.runners[i]
		if err := r.Shutdown(ctx); err != nil {
			slog.Warn("component shutdown encountered error", "name", r.Name(), "error", err)
			errs = append(errs, fmt.Errorf("%s: %w", r.Name(), err))
		} else {
			slog.Info("component stopped", "name", r.Name())
		}
	}

	return errors.Join(errs...)
}
