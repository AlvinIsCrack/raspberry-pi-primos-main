package api

import (
	"log/slog"
	apiHttp "primos/api/http"
	apiUdp "primos/api/udp"
	"primos/services"
)

// AppServices agrega los servicios de dominio necesarios a través de los controladores.
type AppServices struct {
	RoomsLock     *services.RoomsLockService
	RoomsSchedule *services.RoomsScheduleService
}

type Endpoints struct {
	Router        *Router
	UDPController *apiUdp.RoomsUDPController
}

// BuildEndpoints conecta módulos HTTP, UDP y transmisión en tiempo real.
func BuildEndpoints(svcs AppServices) Endpoints {
	router := NewRouter()
	sseHub := apiHttp.NewSSEHub()

	router.AttachHTTP(
		apiHttp.NewRoomsHandler(svcs.RoomsLock, sseHub),
		apiHttp.NewSystemHandler(),
	)

	udpCtrl := apiUdp.NewRoomsUDPController(svcs.RoomsLock, sseHub)

	slog.Debug("api endpoints initialized",
		"http_handlers_count", len(router.httpRoutes),
		"udp_controller_ready", udpCtrl != nil,
	)

	return Endpoints{
		Router:        router,
		UDPController: udpCtrl,
	}
}
