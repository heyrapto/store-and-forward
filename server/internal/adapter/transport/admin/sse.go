package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"messaging/server/internal/domain"
	"messaging/server/internal/usecase/port"
)

type SSEHandler struct {
	Bus port.EventBus
}

func NewSSEHandler(bus port.EventBus) *SSEHandler {
	return &SSEHandler{Bus: bus}
}

func (h *SSEHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	eventCh := make(chan domain.Event, 100)
	
	// Subscribe to all event types
	unsub := h.Bus.Subscribe([]domain.EventType{
		domain.EventMessageQueued,
		domain.EventDeviceOnline,
		domain.EventDeviceOffline,
		domain.EventDeliveryAcked,
		domain.EventDeliveryRead,
		domain.EventDeviceJoined,
		domain.EventMailboxDrained,
	}, func(e domain.Event) {
		select {
		case eventCh <- e:
		default:
			// Client too slow, drop event to prevent blocking the bus
		}
	})
	
	defer unsub()

	ctx := r.Context()
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Send a comment to keep the connection alive
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case e := <-eventCh:
			payloadBytes, _ := json.Marshal(e.Payload)
			
			// Format as SSE
			fmt.Fprintf(w, "event: %s\n", e.Type)
			fmt.Fprintf(w, "data: %s\n\n", payloadBytes)
			flusher.Flush()
		}
	}
}
