package mqtt

import (
	"encoding/json"
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
	broker.OnMessage("sensores/puertas/+", c.handleDoorMessage)
}

func (c *RoomsController) handleDoorMessage(topic string, rawPayload []byte) {
	parts := strings.Split(topic, "/")
	if len(parts) != 3 {
		return
	}
	roomID := parts[2]

	var payload domain.TelemetryPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return
	}

	_ = c.lockService.ProcessTelemetry(roomID, payload, "mqtt_shutdown_or_lwt")
}
