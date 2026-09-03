package domain

import (
	"fmt"
	"time"
)

// DoorState define el estado físico de la puerta reportado por el sensor.
type DoorState string

const (
	DoorUnknown DoorState = "N/A"
	DoorOpen    DoorState = "ABR"
	DoorClosed  DoorState = "CER"
)

func (d DoorState) String() string {
	switch d {
	case DoorOpen:
		return " O "
	case DoorClosed:
		return " X "
	default:
		return "N/A"
	}
}

func (d DoorState) IsValid() bool {
	switch d {
	case DoorUnknown, DoorOpen, DoorClosed:
		return true
	default:
		return false
	}
}

// ConnectivityState define la disponibilidad del nodo IoT.
type ConnectivityState string

const (
	StatusOnline   ConnectivityState = "ONLINE"
	StatusOffline  ConnectivityState = "OFFLINE"
	StatusCritical ConnectivityState = "CRITICAL"
)

// HeartbeatGracePeriod determina el tiempo antes de considerar que el sensor perdió conexión.
const HeartbeatGracePeriod = 5 * time.Minute

// TelemetryPayload unifica la carga útil deserializable desde JSON (HTTP o MQTT).
type TelemetryPayload struct {
	Door         DoorState `json:"door"`
	BatteryLevel *int      `json:"battery_level,omitempty"`
	IsCharging   bool      `json:"is_charging,omitempty"`
	RSSI         int       `json:"rssi,omitempty"`
	Shutdown     bool      `json:"shutdown,omitempty"`
	Reason       string    `json:"reason,omitempty"`
}

// TelemetryReport representa la telemetría validada interna del sensor.
type TelemetryReport struct {
	Door         DoorState
	BatteryLevel int
	IsCharging   bool
	RSSI         int
}

func (r TelemetryReport) Validate() error {
	if !r.Door.IsValid() {
		return fmt.Errorf("estado de puerta inválido '%s': permitidos [%s, %s, %s]",
			r.Door, DoorUnknown, DoorOpen, DoorClosed)
	}
	if r.BatteryLevel < -1 || r.BatteryLevel > 100 {
		return fmt.Errorf("nivel de batería inválido %d: debe estar entre 0 y 100, o -1 si no usa batería", r.BatteryLevel)
	}
	return nil
}

// SensorSnapshot representa la vista calculada e inmutable de un sensor.
type SensorSnapshot struct {
	RoomID       string            `json:"room_id"`
	Door         DoorState         `json:"door"`
	Connectivity ConnectivityState `json:"connectivity"`
	BatteryLevel int               `json:"battery_level"`
	LastSeenAt   time.Time         `json:"last_seen_at"`
	IsStale      bool              `json:"is_stale"`
}
