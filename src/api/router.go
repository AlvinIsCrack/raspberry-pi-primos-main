package api

import (
	"fmt"
	"log/slog"
	netHttp "net/http"
	"time"
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

// responseWriterInterceptor captura el código HTTP de respuesta para el log.
type responseWriterInterceptor struct {
	netHttp.ResponseWriter
	statusCode int
}

func (w *responseWriterInterceptor) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterInterceptor) Flush() {
	if f, ok := w.ResponseWriter.(netHttp.Flusher); ok {
		f.Flush()
	}
}

// RegisterHTTPRoutes monta todos los módulos registrados en el mux estándar.
func (r *Router) RegisterHTTPRoutes(mux *netHttp.ServeMux) {
	tempMux := netHttp.NewServeMux()
	for _, route := range r.httpRoutes {
		route.RegisterHTTP(tempMux)
		slog.Debug("http route handler attached", "handler_type", fmt.Sprintf("%T", route))
	}

	// Middleware global para registrar tráfico HTTP de la API
	mux.HandleFunc("/api/", func(w netHttp.ResponseWriter, req *netHttp.Request) {
		start := time.Now()
		interceptor := &responseWriterInterceptor{ResponseWriter: w, statusCode: netHttp.StatusOK}

		tempMux.ServeHTTP(interceptor, req)

		slog.Debug("api http request",
			"method", req.Method,
			"path", req.URL.Path,
			"status", interceptor.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_addr", req.RemoteAddr,
		)
	})
}
