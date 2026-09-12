package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidRoomID = errors.New("el ID del cuarto debe tener exactamente 3 caracteres alfanuméricos")
	ErrRoomNotFound  = errors.New("sensor no registrado")
)

// RoomID representa la identidad tipada y validada de una sala.
type RoomID string

// NewRoomID valida y sanitiza la clave única de la sala.
func NewRoomID(raw string) (RoomID, error) {
	trimmed := strings.TrimSpace(raw)
	if len(trimmed) != 3 {
		return "", ErrInvalidRoomID
	}
	for i := 0; i < len(trimmed); i++ {
		c := trimmed[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return "", ErrInvalidRoomID
		}
	}
	return RoomID(strings.ToUpper(trimmed)), nil
}

func (r RoomID) String() string {
	return string(r)
}

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
	Shutdown     bool      `json:"shutdown,omitempty"`
	Reason       string    `json:"reason,omitempty"`
}

// TelemetryReport representa la telemetría validada interna del sensor.
type TelemetryReport struct {
	Door         DoorState
	BatteryLevel int
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
	RoomID       RoomID            `json:"room_id"`
	Door         DoorState         `json:"door"`
	Connectivity ConnectivityState `json:"connectivity"`
	BatteryLevel int               `json:"battery_level"`
	LastSeenAt   time.Time         `json:"last_seen_at"`
	IsStale      bool              `json:"is_stale"`
}

// SensorDevice encapsula el estado vivo y la lógica de negocio de un dispositivo IoT.
type SensorDevice struct {
	RoomID        RoomID
	LastReport    TelemetryReport
	LastSeenAt    time.Time
	IsExplicitOff bool
	ShutdownNote  string
}

// NewSensorDevice crea un dispositivo en estado inicial conocido.
func NewSensorDevice(id RoomID) SensorDevice {
	return SensorDevice{
		RoomID: id,
		LastReport: TelemetryReport{
			Door:         DoorUnknown,
			BatteryLevel: -1,
		},
		LastSeenAt:    time.Time{},
		IsExplicitOff: true,
		ShutdownNote:  "Inicialización de nodo",
	}
}

// ApplyTelemetry actualiza el dispositivo con nueva telemetría confirmada.
func (d *SensorDevice) ApplyTelemetry(report TelemetryReport, now time.Time) {
	d.LastReport = report
	d.LastSeenAt = now
	d.IsExplicitOff = false
	d.ShutdownNote = ""
}

// ApplyShutdown apaga voluntariamente el reporte del dispositivo.
func (d *SensorDevice) ApplyShutdown(reason string) {
	d.IsExplicitOff = true
	d.ShutdownNote = reason
	d.LastReport = TelemetryReport{
		Door:         DoorUnknown,
		BatteryLevel: -1,
	}
}

// Snapshot genera una proyección temporal segura según el tiempo de referencia.
func (d SensorDevice) Snapshot(now time.Time) SensorSnapshot {
	isExpired := !d.LastSeenAt.IsZero() && now.Sub(d.LastSeenAt) > HeartbeatGracePeriod
	isOffline := d.IsExplicitOff || isExpired || d.LastSeenAt.IsZero()

	doorState := d.LastReport.Door
	connectivity := StatusOnline
	batteryLevel := d.LastReport.BatteryLevel

	if isOffline || doorState == DoorUnknown {
		doorState = DoorUnknown
		connectivity = StatusOffline
		batteryLevel = -1
	}

	return SensorSnapshot{
		RoomID:       d.RoomID,
		Door:         doorState,
		Connectivity: connectivity,
		BatteryLevel: batteryLevel,
		LastSeenAt:   d.LastSeenAt,
		IsStale:      isExpired,
	}
}
