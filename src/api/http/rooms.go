package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"primos/domain"
	"primos/services"
)

// RoomsHandler handles HTTP requests and real-time streaming for room states.
type RoomsHandler struct {
	lockService *services.RoomsLockService
	hub         *SSEHub
}

// NewRoomsHandler creates an HTTP controller with SSE capability.
func NewRoomsHandler(lockService *services.RoomsLockService, hub *SSEHub) *RoomsHandler {
	return &RoomsHandler{
		lockService: lockService,
		hub:         hub,
	}
}

// RegisterHTTP maps REST routes and the event stream to the HTTP multiplexer.
func (h *RoomsHandler) RegisterHTTP(mux *http.ServeMux) {
	mux.HandleFunc("/api/rooms", h.handleRooms)
	mux.HandleFunc("/api/rooms/", h.handleRoomByID)
	mux.HandleFunc("/api/events", h.hub.ServeHTTP)
}

func (h *RoomsHandler) handleRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	snapshots := h.lockService.GetAllSnapshots()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snapshots)
}

func (h *RoomsHandler) handleRoomByID(w http.ResponseWriter, r *http.Request) {
	roomID := strings.ToUpper(strings.TrimPrefix(r.URL.Path, "/api/rooms/"))
	if len(roomID) != 3 {
		http.Error(w, "Room ID must have exactly 3 characters", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		snapshot, exists := h.lockService.GetSnapshot(roomID)
		if !exists {
			http.Error(w, "Sensor not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(snapshot)

	case http.MethodPost:
		var payload domain.TelemetryPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		if err := h.lockService.ProcessTelemetry(roomID, payload, "http_request"); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Emit SSE update to UI
		if snapshot, exists := h.lockService.GetSnapshot(roomID); exists {
			h.hub.Broadcast("room_updated", snapshot)
		}

		// Retornar setup mode al cliente HTTP
		policy := services.GetCurrentEnergyPolicy()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "telemetry_registered",
			"setup":  policy, // 1, 2 o 3
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
