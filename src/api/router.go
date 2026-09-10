package api

import (
	netHttp "net/http"
)

// HTTPRouteable define el contrato para cualquier controlador que exponga endpoints REST.
type HTTPRouteable interface {
	RegisterHTTP(mux *netHttp.ServeMux)
}

// Router centraliza y despacha el registro de controladores HTTP y MQTT.
type Router struct {
	httpRoutes []HTTPRouteable
}

func NewRouter() *Router {
	return &Router{
		httpRoutes: make([]HTTPRouteable, 0),
	}
}

// AttachHTTP agrega uno o más controladores HTTP al pipeline.
func (r *Router) AttachHTTP(routes ...HTTPRouteable) *Router {
	r.httpRoutes = append(r.httpRoutes, routes...)
	return r
}

// RegisterHTTPRoutes monta todos los módulos registrados en el mux estándar.
func (r *Router) RegisterHTTPRoutes(mux *netHttp.ServeMux) {
	for _, route := range r.httpRoutes {
		route.RegisterHTTP(mux)
	}
}
