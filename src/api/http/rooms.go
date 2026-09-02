package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"primos/domain"
	"primos/services"

	netHttp "net/http"
)

// RoomsHandler expone las operaciones HTTP sobre las salas.
type RoomsHandler struct {
	lockService *services.RoomsLockService
}

func NewRoomsHandler(lockService *services.RoomsLockService) *RoomsHandler {
	return &RoomsHandler{
		lockService: lockService,
	}
}

// RegisterHTTP monta las rutas HTTP del dominio de salas.
func (h *RoomsHandler) RegisterHTTP(mux *netHttp.ServeMux) {
	mux.HandleFunc("/api/rooms", h.handleRooms)
	mux.HandleFunc("/api/rooms/", h.handleRoomByID)
}

func (h *RoomsHandler) handleRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	snapshots := h.lockService.GetAllSnapshots()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snapshots)
}

func (h *RoomsHandler) handleRoomByID(w http.ResponseWriter, r *http.Request) {
	roomID := strings.ToUpper(strings.TrimPrefix(r.URL.Path, "/api/rooms/"))
	if len(roomID) != 3 {
		http.Error(w, "El ID de la sala debe tener exactamente 3 caracteres", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		snapshot, exists := h.lockService.GetSnapshot(roomID)
		if !exists {
			http.Error(w, "Sensor no registrado", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(snapshot)

	case http.MethodPost:
		var payload domain.TelemetryPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Payload JSON inválido: "+err.Error(), http.StatusBadRequest)
			return
		}

		if err := h.lockService.ProcessTelemetry(roomID, payload, "http_shutdown_request"); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"telemetry_registered"}`))

	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}
