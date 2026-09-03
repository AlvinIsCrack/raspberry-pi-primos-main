package mqtt

import (
	"encoding/json"
	"strconv"
	"strings"

	"primos/core"
	"primos/domain"
	"primos/services"
)

// RoomsController atiende la telemetría y eventos MQTT para sensores de cuartos.
type RoomsController struct {
	lockService *services.RoomsLockService
}

// NewRoomsController crea una instancia del controlador MQTT para salas.
func NewRoomsController(lockService *services.RoomsLockService) *RoomsController {
	return &RoomsController{
		lockService: lockService,
	}
}

// RegisterMQTT vincula los tópicos de las salas al broker.
func (c *RoomsController) RegisterMQTT(broker *core.MQTTBroker) {
	// Escucha tanto el tópico base como cualquier sub-tópico (status, bateria, etc.)
	broker.OnMessage("sensores/puertas/#", c.handleDoorMessage)
}

func (c *RoomsController) handleDoorMessage(topic string, rawPayload []byte) {
	parts := strings.Split(topic, "/")
	if len(parts) < 3 {
		return
	}

	roomID := strings.ToUpper(parts[2])
	if len(roomID) != 3 || roomID == "" {
		return
	}

	payloadStr := strings.TrimSpace(string(rawPayload))

	// 1. Manejo por sub-tópico específico: sensores/puertas/{id}/...
	if len(parts) >= 4 {
		subTopic := strings.ToLower(parts[3])
		switch subTopic {
		case "status", "estado", "lwt":
			lower := strings.ToLower(payloadStr)
			if lower == "offline" || lower == "shutdown" {
				_ = c.lockService.ReportShutdown(roomID, "mqtt_lwt_disconnect")
			}
		case "bateria", "battery":
			if level, err := strconv.Atoi(payloadStr); err == nil {
				snapshot, exists := c.lockService.GetSnapshot(roomID)
				door := domain.DoorUnknown
				if exists {
					door = snapshot.Door
				}
				_ = c.lockService.ReportTelemetry(roomID, domain.TelemetryReport{
					Door:         door,
					BatteryLevel: level,
				})
			}
		}
		return
	}

	// 2. Manejo en tópico base: sensores/puertas/{id}
	// Intento A: JSON estructurado
	var payload domain.TelemetryPayload
	if err := json.Unmarshal(rawPayload, &payload); err == nil {
		_ = c.lockService.ProcessTelemetry(roomID, payload, "mqtt_shutdown_or_lwt")
		return
	}

	// Intento B: Payload de texto plano ultra-ligero ("ABR", "CER", "N/A")
	upperText := strings.ToUpper(payloadStr)
	switch domain.DoorState(upperText) {
	case domain.DoorOpen, domain.DoorClosed, domain.DoorUnknown:
		_ = c.lockService.ProcessTelemetry(roomID, domain.TelemetryPayload{
			Door: domain.DoorState(upperText),
		}, "mqtt_raw_state")
	}
}
