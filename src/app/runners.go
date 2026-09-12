package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	apiUdp "primos/api/udp"
)

// HTTPServerRunner adapta *http.Server a la interfaz ServiceRunner.
type HTTPServerRunner struct {
	server *http.Server
	addr   string
}

func (r *HTTPServerRunner) Name() string { return "http_server" }

func (r *HTTPServerRunner) Start() error {
	ln, err := net.Listen("tcp", r.addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", r.addr, err)
	}
	go func() {
		if err := r.server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server stopped unexpectedly", "error", err, "addr", r.addr)
		}
	}()
	return nil
}

func (r *HTTPServerRunner) Shutdown(ctx context.Context) error {
	return r.server.Shutdown(ctx)
}

// UDPControllerRunner adapta *apiUdp.RoomsUDPController a ServiceRunner.
type UDPControllerRunner struct {
	ctrl *apiUdp.RoomsUDPController
	addr string
}

func (r *UDPControllerRunner) Name() string { return "udp_controller" }

func (r *UDPControllerRunner) Start() error {
	return r.ctrl.Start(r.addr)
}

func (r *UDPControllerRunner) Shutdown(_ context.Context) error {
	return r.ctrl.Close()
}
