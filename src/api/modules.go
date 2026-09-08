package api

import (
	apiHttp "primos/api/http"
	apiMqtt "primos/api/mqtt"
	"primos/services"
)

// AppServices aggregates the domain services needed across controllers.
type AppServices struct {
	RoomsLock *services.RoomsLockService
}

// BuildDefaultRouter wires HTTP, MQTT, and real-time streaming modules.
func BuildDefaultRouter(svcs AppServices) *Router {
	router := NewRouter()
	sseHub := apiHttp.NewSSEHub()

	router.AttachHTTP(
		apiHttp.NewRoomsHandler(svcs.RoomsLock, sseHub),
	)

	router.AttachMQTT(
		apiMqtt.NewRoomsController(svcs.RoomsLock, sseHub),
	)

	return router
}
