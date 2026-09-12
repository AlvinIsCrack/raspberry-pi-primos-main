package app

import "context"

// ServiceRunner define el contrato para cualquier componente que tenga
// ejecución en background y apagado limpio (HTTP, UDP, Workers).
type ServiceRunner interface {
	Name() string
	Start() error
	Shutdown(ctx context.Context) error
}
