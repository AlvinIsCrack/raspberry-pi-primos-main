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

func (c *RoomsUDPController) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
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

		c.handlePacket(buf[:n], remoteAddr)
	}
}

func (c *RoomsUDPController) handlePacket(data []byte, remoteAddr *net.UDPAddr) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for i := 1; i <= 3; i++ {
		if data[i] >= 'a' && data[i] <= 'z' {
			data[i] -= 32
		}
	}
	roomID := string(data[1:4])

	if _, exists := c.lockService.GetSnapshot(ctx, roomID); !exists {
		slog.Warn("UDP telemetry dropped: unregistered room", "room", roomID)
		return
	}

	doorState := domain.DoorClosed
	if data[4] == 1 {
		doorState = domain.DoorOpen
	}

	_ = c.lockService.ProcessTelemetry(ctx, roomID, domain.TelemetryPayload{
		Door: doorState,
	}, "udp_datagram")

	if updatedSnapshot, exists := c.lockService.GetSnapshot(ctx, roomID); exists {
		c.hub.Broadcast("room_updated", updatedSnapshot)
	}

	policy := uint8(services.GetCurrentEnergyPolicy())
	response := []byte{0x5A, policy}
	_, _ = c.conn.WriteToUDP(response, remoteAddr)
	slog.Debug("UDP packet processed", "room", roomID, "door", doorState, "policySent", policy)
}
