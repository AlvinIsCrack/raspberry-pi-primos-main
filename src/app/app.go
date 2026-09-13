package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"primos/api"
	"primos/config"
	"primos/services"
	"primos/system"
	"primos/usm"
)

const appBanner = `
  ____  ____  ___ __  __  ___  ____ 
 |  _ \|  _ \|_ _|  \/  |/ _ \/ ___|
 | |_) | |_) || || |\/| | | | \___ \
 |  __/|  _ < | || |  | | |_| |___) |
 |_|   |_| \_\___|_|  |_|\___/|____/`

type App struct {
	cfg     *config.AppConfig
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

	return &App{
		cfg:     cfg,
		runners: runners,
	}, nil
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

	// Verificar conectividad a internet en segundo plano para no demorar el arranque
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if system.HasInternetAccess(ctx) {
			slog.Info("network connectivity verified", "internet", true)
			return
		}
		slog.Warn("network probe failed: no internet access detected", "internet", false)
	}()

	sched := usm.GetCurrentAcademicSchedule(time.Now())
	slog.Info("academic schedule evaluated",
		"period_kind", sched.Kind,
		"block", fmt.Sprintf("%d-%d", sched.Block.FirstIndex, sched.Block.SecondIndex),
		"sub_block", sched.ActiveSubBlock,
	)

	if !a.cfg.IsProduction() {
		printBanner(a.cfg)
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

func printBanner(cfg *config.AppConfig) {
	fmt.Printf("%s\n\n  -> HTTP: %s\n  -> UDP:  %s\n  -> Env:  %s\n\n",
		appBanner, resolveDisplayURL(cfg.HTTPAddr), cfg.UDPAddr, cfg.AppEnv)
}

func resolveDisplayURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		if !strings.Contains(addr, ":") {
			return fmt.Sprintf("http://localhost:%s", addr)
		}
		return "http://localhost" + addr
	}
	if host == "" || host == "0.0.0.0" || host == "::" || host == "[::]" {
		host = "localhost"
	}
	return fmt.Sprintf("http://%s:%s", host, port)
}
