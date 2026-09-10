package http

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"primos/system"
	"time"
)

type SystemHandler struct{}

type SystemStatusResponse struct {
	Connected bool     `json:"connected"`
	LocalIPs  []string `json:"local_ips"`
}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

func (h *SystemHandler) RegisterHTTP(mux *http.ServeMux) {
	mux.HandleFunc("/api/system/status", h.handleStatus)
}

func (h *SystemHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// Consulta si hay salida a internet utilizando el paquete system existente
	isConnected := system.HasInternetAccess(ctx)

	// Obtiene las IPs locales no loopback de la placa
	var ips []string
	if ifaces, err := net.Interfaces(); err == nil {
		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
				continue
			}
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if ip != nil && ip.To4() != nil {
					ips = append(ips, ip.String())
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(SystemStatusResponse{
		Connected: isConnected,
		LocalIPs:  ips,
	})
}
