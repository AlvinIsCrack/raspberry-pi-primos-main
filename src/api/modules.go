package api

import (
	apiHttp "primos/api/http"
	apiUdp "primos/api/udp"
	"primos/services"
)

// AppServices aggregates the domain services needed across controllers.
type AppServices struct {
	RoomsLock *services.RoomsLockService
}

type Endpoints struct {
	Router        *Router
	UDPController *apiUdp.RoomsUDPController
}

// BuildEndpoints wires HTTP, UDP, and real-time streaming modules.
func BuildEndpoints(svcs AppServices) Endpoints {
	router := NewRouter()
	sseHub := apiHttp.NewSSEHub()

	router.AttachHTTP(
		apiHttp.NewRoomsHandler(svcs.RoomsLock, sseHub),
		apiHttp.NewSystemHandler(),
	)

	udpCtrl := apiUdp.NewRoomsUDPController(svcs.RoomsLock, sseHub)

	return Endpoints{
		Router:        router,
		UDPController: udpCtrl,
	}
}
