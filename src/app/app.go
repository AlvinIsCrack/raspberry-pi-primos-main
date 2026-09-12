package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"

	"primos/api"
	"primos/config"
	"primos/db/repository/memory"
	"primos/domain"
	"primos/services"
)

type App struct {
	cfg        *config.AppConfig
	httpServer *http.Server
	endpoints  api.Endpoints
}

// New crea e inicializa las dependencias centrales de la aplicación.
func New(cfg *config.AppConfig, webAssets fs.FS, webAssetsDir string) (*App, error) {
	slog.Info("initializing application components")

	var roomIDs []domain.RoomID
	for _, r := range config.Rooms {
		roomIDs = append(roomIDs, domain.RoomID(r))
	}

	deviceRepository := memory.NewInMemoryDeviceRepository(roomIDs...)
	slog.Info("in-memory repository initialized", "rooms_count", len(roomIDs))

	lockService := services.NewRoomsLockService(deviceRepository)

	// Inicialización de Schedule
	scheduleRepository := memory.NewInMemoryScheduleRepository()
	scheduleService, err := services.NewRoomsScheduleService(scheduleRepository, scheduleRepository)
	if err != nil {
		return nil, fmt.Errorf("schedule service initialization failed: %w", err)
	}
	slog.Info("schedule service initialized")

	endpoints := api.BuildEndpoints(api.AppServices{
		RoomsLock:     lockService,
		RoomsSchedule: scheduleService,
	})
	slog.Info("api endpoints and controllers created")

	httpMux := http.NewServeMux()
	endpoints.Router.RegisterHTTPRoutes(httpMux)

	if err := setupSPAFallback(httpMux, webAssets, webAssetsDir); err != nil {
		return nil, fmt.Errorf("setup spa fallback (%s): %w", webAssetsDir, err)
	}
	slog.Info("static assets and spa fallback registered", "dir", webAssetsDir)

	return &App{
		cfg:       cfg,
		endpoints: endpoints,
		httpServer: &http.Server{
			Addr:    cfg.HTTPAddr,
			Handler: httpMux,
		},
	}, nil
}

// Start levanta el servidor UDP y el servidor HTTP en segundo plano.
func (a *App) Start() error {
	if err := a.endpoints.UDPController.Start(a.cfg.UDPAddr); err != nil {
		return fmt.Errorf("fatal UDP Server error: %w", err)
	}
	slog.Info("udp server listening", "addr", a.cfg.UDPAddr)

	ln, err := net.Listen("tcp", a.cfg.HTTPAddr)
	if err != nil {
		_ = a.endpoints.UDPController.Close()
		return fmt.Errorf("http listen error on %s: %w", a.cfg.HTTPAddr, err)
	}
	slog.Info("http server listening", "addr", ln.Addr().String())

	go func() {
		if err := a.httpServer.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server failed", "error", err, "addr", a.cfg.HTTPAddr)
		}
	}()

	return nil
}

// Shutdown detiene de forma ordenada los servidores HTTP y UDP.
func (a *App) Shutdown(ctx context.Context) error {
	slog.Info("shutting down application servers")
	var errs []error

	if err := a.httpServer.Shutdown(ctx); err != nil {
		slog.Warn("http server graceful shutdown failed", "error", err)
		errs = append(errs, fmt.Errorf("http shutdown: %w", err))
	} else {
		slog.Info("http server stopped")
	}

	if err := a.endpoints.UDPController.Close(); err != nil {
		slog.Warn("udp controller close failed", "error", err)
		errs = append(errs, fmt.Errorf("udp close: %w", err))
	} else {
		slog.Info("udp controller stopped")
	}

	return errors.Join(errs...)
}
