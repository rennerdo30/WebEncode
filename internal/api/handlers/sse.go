package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/rennerdo30/webencode/pkg/bus"
	"github.com/rennerdo30/webencode/pkg/logger"
)

// SSEHandler handles Server-Sent Events for real-time updates
type SSEHandler struct {
	bus     *bus.Bus
	logger  *logger.Logger
	clients map[string]map[chan []byte]bool
	mu      sync.RWMutex
}

func NewSSEHandler(b *bus.Bus, l *logger.Logger) *SSEHandler {
	return &SSEHandler{
		bus:     b,
		logger:  l,
		clients: make(map[string]map[chan []byte]bool),
	}
}

func (h *SSEHandler) Register(r chi.Router) {
	r.Get("/v1/events/jobs", h.JobEvents)
	r.Get("/v1/events/jobs/{id}", h.JobDetailEvents)
	r.Get("/v1/events/streams/{id}", h.StreamEvents)
	r.Get("/v1/events/dashboard", h.DashboardEvents)
}

func (h *SSEHandler) addClient(topic string, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[topic] == nil {
		h.clients[topic] = make(map[chan []byte]bool)
	}
	h.clients[topic][ch] = true
}

func (h *SSEHandler) removeClient(topic string, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[topic] != nil {
		delete(h.clients[topic], ch)
	}
}

func (h *SSEHandler) broadcast(topic string, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients[topic] {
		select {
		case ch <- data:
		default:
			// Client too slow, skip
		}
	}
}

// setupSSE sets up SSE headers and returns a channel for events
func (h *SSEHandler) setupSSE(w http.ResponseWriter, r *http.Request, topic string) (chan []byte, bool) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return nil, false
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := make(chan []byte, 100)
	h.addClient(topic, ch)

	// Send initial connection event
	w.Write([]byte("event: connected\ndata: {\"status\":\"connected\"}\n\n"))
	flusher.Flush()

	go func() {
		<-r.Context().Done()
		h.removeClient(topic, ch)
		close(ch)
	}()

	return ch, true
}

func (h *SSEHandler) streamEvents(w http.ResponseWriter, ch chan []byte) {
	flusher := w.(http.Flusher)
	for data := range ch {
		w.Write([]byte("data: "))
		w.Write(data)
		w.Write([]byte("\n\n"))
		flusher.Flush()
	}
}

// JobEvents streams all job updates
func (h *SSEHandler) JobEvents(w http.ResponseWriter, r *http.Request) {
	ch, ok := h.setupSSE(w, r, "jobs")
	if !ok {
		return
	}
	h.streamEvents(w, ch)
}

// JobDetailEvents streams updates for a specific job
func (h *SSEHandler) JobDetailEvents(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	topic := "job:" + id

	ch, ok := h.setupSSE(w, r, topic)
	if !ok {
		return
	}
	h.streamEvents(w, ch)
}

// StreamEvents streams live telemetry for a specific stream
func (h *SSEHandler) StreamEvents(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	topic := "stream:" + id

	ch, ok := h.setupSSE(w, r, topic)
	if !ok {
		return
	}
	h.streamEvents(w, ch)
}

// DashboardEvents streams dashboard stats updates
func (h *SSEHandler) DashboardEvents(w http.ResponseWriter, r *http.Request) {
	ch, ok := h.setupSSE(w, r, "dashboard")
	if !ok {
		return
	}
	h.streamEvents(w, ch)
}

// Publish sends an event to all clients subscribed to a topic
func (h *SSEHandler) Publish(topic string, event interface{}) {
	data, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("Failed to marshal SSE event", "error", err)
		return
	}
	h.broadcast(topic, data)
}

// PublishJobUpdate publishes a job update event
func (h *SSEHandler) PublishJobUpdate(jobID string, data interface{}) {
	eventData, _ := json.Marshal(data)
	h.broadcast("jobs", eventData)
	h.broadcast("job:"+jobID, eventData)
}

// PublishStreamTelemetry publishes stream telemetry
func (h *SSEHandler) PublishStreamTelemetry(streamID string, data interface{}) {
	eventData, _ := json.Marshal(data)
	h.broadcast("stream:"+streamID, eventData)
}

// PublishDashboardStats publishes dashboard stats
func (h *SSEHandler) PublishDashboardStats(data interface{}) {
	eventData, _ := json.Marshal(data)
	h.broadcast("dashboard", eventData)
}
