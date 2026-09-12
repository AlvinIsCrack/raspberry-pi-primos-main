package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"

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
	var roomIDs []domain.RoomID
	for _, r := range config.Rooms {
		roomIDs = append(roomIDs, domain.RoomID(r))
	}
	deviceRepository := memory.NewInMemoryDeviceRepository(roomIDs...)
	lockService := services.NewRoomsLockService(deviceRepository)

	endpoints := api.BuildEndpoints(api.AppServices{
		RoomsLock: lockService,
	})

	httpMux := http.NewServeMux()
	endpoints.Router.RegisterHTTPRoutes(httpMux)
	if err := setupSPAFallback(httpMux, webAssets, webAssetsDir); err != nil {
		return nil, err
	}

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

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "HTTP Server error: %v\n", err)
		}
	}()

	return nil
}

// Shutdown detiene de forma ordenada los servidores HTTP y UDP.
func (a *App) Shutdown(ctx context.Context) error {
	var errs []error
	if err := a.httpServer.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("http shutdown: %w", err))
	}

	if err := a.endpoints.UDPController.Close(); err != nil {
		errs = append(errs, fmt.Errorf("udp close: %w", err))
	}
	return errors.Join(errs...)
}
