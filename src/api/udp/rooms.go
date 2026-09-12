package udp

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"time"

	"primos/api/http"
	"primos/domain"
	"primos/services"
)

// SensorPacket estructura esperada desde el ESP8266 (6 bytes):
// [0]   : Magic Byte (0x5A)
// [1:4] : Room ID (3 bytes ASCII, ej: "LDS")
// [4]   : Door State (0 = Closed / Locked, 1 = Open / Unlocked)
// [5]   : Battery Voltage / 100 (opcional, ej. 41 para 4.1V)
type SensorPacket struct {
	Magic     uint8
	RoomID    string
	State     uint8
	BatteryMv uint16
}

type RoomsUDPController struct {
	lockService *services.RoomsLockService
	hub         *http.SSEHub
	conn        *net.UDPConn
}

func NewRoomsUDPController(lockService *services.RoomsLockService, hub *http.SSEHub) *RoomsUDPController {
	return &RoomsUDPController{
		lockService: lockService,
		hub:         hub,
	}
}

// Start inicia el socket UDP en la dirección especificada (ej. ":1884")
func (c *RoomsUDPController) Start(addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	c.conn = conn

	go c.listenLoop()
	return nil
}

func (c *RoomsUDPController) Close() {
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *RoomsUDPController) listenLoop() {
	buf := make([]byte, 64)
	for {
		n, remoteAddr, err := c.conn.ReadFromUDP(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return // Clean exit on Close()
			}
			slog.Error("UDP read error", "err", err)
			continue
		}

		if n < 5 || buf[0] != 0x5A { // Magic byte 0x5A
			continue
		}

		for i := 1; i <= 3; i++ {
			if buf[i] >= 'a' && buf[i] <= 'z' {
				buf[i] -= 32
			}
		}
		roomID := string(buf[1:4])

		doorRaw := buf[4]

		var doorState domain.DoorState
		if doorRaw == 1 {
			doorState = domain.DoorOpen
		} else {
			doorState = domain.DoorClosed
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_ = c.lockService.ProcessTelemetry(ctx, roomID, domain.TelemetryPayload{
			Door: doorState,
		}, "udp_datagram")

		if snapshot, exists := c.lockService.GetSnapshot(ctx, roomID); exists {
			c.hub.Broadcast("room_updated", snapshot)
		}

		// Responder al instante al sensor con el modo de setup actual (1 byte)
		policy := uint8(services.GetCurrentEnergyPolicy())
		response := []byte{0x5A, policy} // ACK 0x5A + Modo (1, 2 o 3)
		_, _ = c.conn.WriteToUDP(response, remoteAddr)

		slog.Debug("UDP packet processed", "room", roomID, "door", doorState, "policySent", policy)
	}
}
