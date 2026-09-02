package http

import (
	"encoding/json"
	"net/http"
	"time"

	"primos/services"
)

type SystemHandler struct {
	updaterService *services.UpdaterService
	onRestart      func()
}

func NewSystemHandler(updaterService *services.UpdaterService, onRestart func()) *SystemHandler {
	return &SystemHandler{
		updaterService: updaterService,
		onRestart:      onRestart,
	}
}

func (h *SystemHandler) RegisterHTTP(mux *http.ServeMux) {
	mux.HandleFunc("/api/system/updates", h.handleUpdates)
}

type updateRequest struct {
	Version string `json:"version"`
}

func (h *SystemHandler) handleUpdates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req updateRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	result, err := h.updaterService.CheckAndApply(req.Version)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)

	if result.Updated && h.onRestart != nil {
		go func() {
			time.Sleep(1 * time.Second)
			h.onRestart()
		}()
	}
}
