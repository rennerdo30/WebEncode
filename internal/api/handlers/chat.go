package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/internal/plugin_manager"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/errors"
	"github.com/rennerdo30/webencode/pkg/logger"
)

// ChatMessage represents a unified chat message from any platform
type ChatMessage struct {
	ID         string `json:"id"`
	Platform   string `json:"platform"`
	AuthorName string `json:"author_name"`
	Content    string `json:"content"`
	Timestamp  int64  `json:"timestamp"`
}

// ChatHandler handles chat-related API endpoints
type ChatHandler struct {
	db     store.Querier
	pm     *plugin_manager.Manager
	logger *logger.Logger
}

func NewChatHandler(db store.Querier, pm *plugin_manager.Manager, l *logger.Logger) *ChatHandler {
	return &ChatHandler{db: db, pm: pm, logger: l}
}

func (h *ChatHandler) Register(r chi.Router) {
	r.Get("/v1/streams/{id}/chat", h.GetChatMessages)
	r.Post("/v1/streams/{id}/chat", h.SendChatMessage)
}

// GetChatMessages fetches chat messages from all enabled restream destinations
func (h *ChatHandler) GetChatMessages(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var uid pgtype.UUID
	if err := uid.Scan(id); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	stream, err := h.db.GetStream(r.Context(), uid)
	if err != nil {
		errors.Response(w, r, errors.ErrNotFound)
		return
	}

	// Parse restream destinations from the stream
	var destinations []RestreamDestination
	if len(stream.RestreamDestinations) > 0 {
		if err := json.Unmarshal(stream.RestreamDestinations, &destinations); err != nil {
			h.logger.Warn("Failed to parse restream destinations", "error", err)
		}
	}

	// Collect messages from all enabled publisher plugins
	var allMessages []ChatMessage

	for _, dest := range destinations {
		if !dest.Enabled {
			continue
		}

		// Get the publisher plugin
		rawPub, ok := h.pm.Publisher[dest.PluginID]
		if !ok {
			h.logger.Debug("Publisher plugin not found", "plugin_id", dest.PluginID)
			continue
		}

		pubPlugin, ok := rawPub.(pb.PublisherServiceClient)
		if !ok {
			continue
		}

		// Get chat messages from the plugin
		resp, err := pubPlugin.GetChatMessages(r.Context(), &pb.GetChatMessagesRequest{
			AccessToken: dest.AccessToken,
			ChannelId:   stream.StreamKey,
		})
		if err != nil {
			h.logger.Debug("Failed to get chat from plugin", "plugin_id", dest.PluginID, "error", err)
			continue
		}

		// Convert to unified format
		for _, msg := range resp.Messages {
			allMessages = append(allMessages, ChatMessage{
				ID:         msg.Id,
				Platform:   msg.Platform,
				AuthorName: msg.AuthorName,
				Content:    msg.Content,
				Timestamp:  msg.Timestamp,
			})
		}
	}

	if err := json.NewEncoder(w).Encode(allMessages); err != nil {
		h.logger.Error("Failed to encode chat messages", "error", err)
	}
}

// SendChatMessageRequest represents a request to send a chat message
type SendChatMessageRequest struct {
	Message  string `json:"message"`
	Platform string `json:"platform,omitempty"` // Optional: send to specific platform only
}

// SendChatMessage sends a message to chat on one or all platforms
func (h *ChatHandler) SendChatMessage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var uid pgtype.UUID
	if err := uid.Scan(id); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	var req SendChatMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	if req.Message == "" {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	stream, err := h.db.GetStream(r.Context(), uid)
	if err != nil {
		errors.Response(w, r, errors.ErrNotFound)
		return
	}

	// Parse restream destinations
	var destinations []RestreamDestination
	if len(stream.RestreamDestinations) > 0 {
		if err := json.Unmarshal(stream.RestreamDestinations, &destinations); err != nil {
			h.logger.Warn("Failed to parse restream destinations", "error", err)
		}
	}

	// Track which platforms we sent to
	sentTo := []string{}
	errors := []string{}

	for _, dest := range destinations {
		if !dest.Enabled {
			continue
		}

		// If platform specified, only send to that platform
		if req.Platform != "" && dest.PluginID != req.Platform {
			continue
		}

		// Get the publisher plugin
		rawPub, ok := h.pm.Publisher[dest.PluginID]
		if !ok {
			continue
		}

		pubPlugin, ok := rawPub.(pb.PublisherServiceClient)
		if !ok {
			continue
		}

		// Send chat message
		_, err := pubPlugin.SendChatMessage(r.Context(), &pb.SendChatMessageRequest{
			AccessToken: dest.AccessToken,
			ChannelId:   stream.StreamKey,
			Message:     req.Message,
		})
		if err != nil {
			h.logger.Warn("Failed to send chat to plugin", "plugin_id", dest.PluginID, "error", err)
			errors = append(errors, dest.PluginID)
			continue
		}

		sentTo = append(sentTo, dest.PluginID)
	}

	response := map[string]interface{}{
		"success": len(sentTo) > 0,
		"sent_to": sentTo,
	}
	if len(errors) > 0 {
		response["errors"] = errors
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}
