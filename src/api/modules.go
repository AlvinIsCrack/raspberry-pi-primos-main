package api

import (
	apiHttp "primos/api/http"
	apiMqtt "primos/api/mqtt"
	"primos/services"
)

// AppServices agrupa los servicios de negocio requeridos por los controladores.
type AppServices struct {
	RoomsLock *services.RoomsLockService
	Updater   *services.UpdaterService
	OnRestart func()
}

// BuildDefaultRouter ensambla todos los controladores del sistema.
func BuildDefaultRouter(svcs AppServices) *Router {
	router := NewRouter()

	// Registro de módulos HTTP
	router.AttachHTTP(
		apiHttp.NewRoomsHandler(svcs.RoomsLock),
		apiHttp.NewSystemHandler(svcs.Updater, svcs.OnRestart),
	)

	// Registro de módulos MQTT
	router.AttachMQTT(
		apiMqtt.NewRoomsController(svcs.RoomsLock),
	)

	return router
}
