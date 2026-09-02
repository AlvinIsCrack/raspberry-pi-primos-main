package api

import (
	netHttp "net/http"

	"primos/core"
)

// HTTPRouteable define el contrato para cualquier controlador que exponga endpoints REST.
type HTTPRouteable interface {
	RegisterHTTP(mux *netHttp.ServeMux)
}

// MQTTRouteable define el contrato para cualquier controlador que atienda tópicos MQTT.
type MQTTRouteable interface {
	RegisterMQTT(broker *core.MQTTBroker)
}

// Router centraliza y despacha el registro de controladores HTTP y MQTT.
type Router struct {
	httpRoutes []HTTPRouteable
	mqttRoutes []MQTTRouteable
}

func NewRouter() *Router {
	return &Router{
		httpRoutes: make([]HTTPRouteable, 0),
		mqttRoutes: make([]MQTTRouteable, 0),
	}
}

// AttachHTTP agrega uno o más controladores HTTP al pipeline.
func (r *Router) AttachHTTP(routes ...HTTPRouteable) *Router {
	r.httpRoutes = append(r.httpRoutes, routes...)
	return r
}

// AttachMQTT agrega uno o más controladores MQTT al pipeline.
func (r *Router) AttachMQTT(routes ...MQTTRouteable) *Router {
	r.mqttRoutes = append(r.mqttRoutes, routes...)
	return r
}

// RegisterHTTPRoutes monta todos los módulos registrados en el mux estándar.
func (r *Router) RegisterHTTPRoutes(mux *netHttp.ServeMux) {
	for _, route := range r.httpRoutes {
		route.RegisterHTTP(mux)
	}
}

// RegisterMQTTRoutes enlaza todos los módulos registrados con el broker.
func (r *Router) RegisterMQTTRoutes(broker *core.MQTTBroker) {
	for _, route := range r.mqttRoutes {
		route.RegisterMQTT(broker)
	}
}
