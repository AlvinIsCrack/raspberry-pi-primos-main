package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"primos/api"
	"primos/config"
	"primos/services"
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

	lockService := services.NewRoomsLockService()

	endpoints := api.BuildEndpoints(api.AppServices{
		RoomsLock: lockService,
	})

	// Iniciar servidor UDP para sensores
	if err := endpoints.UDPController.Start(cfg.UDPAddr); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal UDP Server error: %v\n", err)
		os.Exit(1)
	}

	httpMux := http.NewServeMux()
	endpoints.Router.RegisterHTTPRoutes(httpMux)

	buildSubtree, subErr := fs.Sub(embeddedWebAssets, webAssetsDir)
	if subErr != nil {
		fmt.Fprintf(os.Stderr, "Static assets initialization failed: %v\n", subErr)
		os.Exit(1)
	}

	// SPA fallback file server ensuring index.html availability
	fileServer := http.FileServer(http.FS(buildSubtree))
	httpMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		stat, err := fs.Stat(buildSubtree, path)
		if errors.Is(err, fs.ErrNotExist) || (err == nil && stat.IsDir()) {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})

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

	endpoints.UDPController.Close()
}
