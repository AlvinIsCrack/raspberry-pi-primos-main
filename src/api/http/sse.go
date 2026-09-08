package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// SSEHub broadcasts real-time updates over HTTP Server-Sent Events.
type SSEHub struct {
	mu      sync.RWMutex
	clients map[chan []byte]struct{}
}

// NewSSEHub initializes a thread-safe SSE broker.
func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[chan []byte]struct{}),
	}
}

// Broadcast serializes and pushes payload data to all connected listeners.
func (h *SSEHub) Broadcast(eventType string, data any) {
	raw, err := json.Marshal(data)
	if err != nil {
		return
	}

	frame := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(raw))
	msg := []byte(frame)

	h.mu.RLock()
	defer h.mu.RUnlock()

	for clientChan := range h.clients {
		select {
		case clientChan <- msg:
		default:
			// Buffer full or slow consumer, drop message to prevent pipeline lock
		}
	}
}

// ServeHTTP handles the long-lived streaming connection for clients.
func (h *SSEHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	msgChan := make(chan []byte, 16)

	h.mu.Lock()
	h.clients[msgChan] = struct{}{}
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, msgChan)
		close(msgChan)
		h.mu.Unlock()
	}()

	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return
		case msg := <-msgChan:
			if _, err := w.Write(msg); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
