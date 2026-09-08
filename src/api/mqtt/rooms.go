package mqtt

import (
	"encoding/json"
	"strconv"
	"strings"

	"primos/api/http"
	"primos/core"
	"primos/domain"
	"primos/services"
)

// RoomsController handles MQTT telemetry and relays updates to the SSE stream.
type RoomsController struct {
	lockService *services.RoomsLockService
	hub         *http.SSEHub
}

// NewRoomsController creates an instance of the MQTT rooms controller.
func NewRoomsController(lockService *services.RoomsLockService, hub *http.SSEHub) *RoomsController {
	return &RoomsController{
		lockService: lockService,
		hub:         hub,
	}
}

// RegisterMQTT binds room topic listeners to the embedded MQTT broker.
func (c *RoomsController) RegisterMQTT(broker *core.MQTTBroker) {
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

	if len(parts) >= 4 {
		subTopic := strings.ToLower(parts[3])
		switch subTopic {
		case "status", "estado", "lwt":
			lower := strings.ToLower(payloadStr)
			if lower == "offline" || lower == "shutdown" {
				_ = c.lockService.ReportShutdown(roomID, "mqtt_lwt_disconnect")
				c.notifyUpdate(roomID)
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
				c.notifyUpdate(roomID)
			}
		}
		return
	}

	var payload domain.TelemetryPayload
	if err := json.Unmarshal(rawPayload, &payload); err == nil {
		_ = c.lockService.ProcessTelemetry(roomID, payload, "mqtt_telemetry")
		c.notifyUpdate(roomID)
		return
	}

	upperText := strings.ToUpper(payloadStr)
	var doorState domain.DoorState
	switch upperText {
	case string(domain.DoorOpen), "1", "TRUE", "OPEN", "ABIERTO", "ABIERTA":
		doorState = domain.DoorOpen
	case string(domain.DoorClosed), "0", "FALSE", "CLOSED", "CERRADO", "CERRADA":
		doorState = domain.DoorClosed
	case string(domain.DoorUnknown), "?", "NA", "UNKNOWN":
		doorState = domain.DoorUnknown
	}

	if doorState != "" {
		_ = c.lockService.ProcessTelemetry(roomID, domain.TelemetryPayload{
			Door: doorState,
		}, "mqtt_raw_state")
		c.notifyUpdate(roomID)
	}
}

func (c *RoomsController) notifyUpdate(roomID string) {
	if snapshot, exists := c.lockService.GetSnapshot(roomID); exists {
		c.hub.Broadcast("room_updated", snapshot)
	}
}
