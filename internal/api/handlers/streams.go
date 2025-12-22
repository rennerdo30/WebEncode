package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/errors"
	"github.com/rennerdo30/webencode/pkg/logger"
)

type StreamResponse struct {
	ID          string    `json:"id"`
	StreamKey   string    `json:"stream_key"`
	IsLive      bool      `json:"is_live"`
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	IngestURL   string    `json:"ingest_url,omitempty"`
	PlaybackURL string    `json:"playback_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func toStreamResponse(s store.Stream) StreamResponse {
	// Support both legacy ingest_url and new ingest_server column
	ingestURL := s.IngestUrl.String
	if ingestURL == "" && s.IngestServer.Valid {
		ingestURL = s.IngestServer.String
	}

	// Default ingest URL if none provided
	if ingestURL == "" {
		ingestURL = "rtmp://localhost/live"
	}

	// Generate a playback URL if missing
	playbackURL := s.PlaybackUrl.String
	if playbackURL == "" {
		// Default to HLS endpoint on the default MediaMTX port
		playbackURL = "http://localhost:8889/live/" + s.StreamKey + "/index.m3u8"
	}

	return StreamResponse{
		ID:          s.ID.String(),
		StreamKey:   s.StreamKey,
		IsLive:      s.IsLive,
		Title:       s.Title.String,
		Description: s.Description.String,
		IngestURL:   ingestURL,
		PlaybackURL: playbackURL,
		CreatedAt:   s.CreatedAt.Time,
	}
}

type StreamsHandler struct {
	db     store.Querier
	logger *logger.Logger
}

func NewStreamsHandler(db store.Querier, l *logger.Logger) *StreamsHandler {
	return &StreamsHandler{db: db, logger: l}
}

func (h *StreamsHandler) Register(r chi.Router) {
	r.Get("/v1/streams", h.ListStreams)
	r.Post("/v1/streams", h.CreateStream)
	r.Get("/v1/streams/{id}", h.GetStream)
	r.Get("/v1/streams/{id}/destinations", h.GetStreamDestinations)
	r.Put("/v1/streams/{id}/destinations", h.UpdateStreamDestinations)
}

func (h *StreamsHandler) ListStreams(w http.ResponseWriter, r *http.Request) {
	limit := 10
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	streams, err := h.db.ListStreams(r.Context(), store.ListStreamsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		h.logger.Error("Failed to list streams", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	resp := make([]StreamResponse, 0, len(streams))
	for _, s := range streams {
		resp = append(resp, toStreamResponse(s))
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *StreamsHandler) GetStream(w http.ResponseWriter, r *http.Request) {
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

	if err := json.NewEncoder(w).Encode(toStreamResponse(stream)); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

type CreateStreamRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *StreamsHandler) CreateStream(w http.ResponseWriter, r *http.Request) {
	var req CreateStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	userID := "00000000-0000-0000-0000-000000000001"
	var uid pgtype.UUID
	uid.Scan(userID)

	streamKey := generateStreamKey()

	stream, err := h.db.CreateStream(r.Context(), store.CreateStreamParams{
		StreamKey:      streamKey,
		UserID:         uid,
		Title:          pgtype.Text{String: req.Title, Valid: true},
		Description:    pgtype.Text{String: req.Description, Valid: true},
		ArchiveEnabled: pgtype.Bool{Bool: true, Valid: true},
		IngestServer:   pgtype.Text{String: "rtmp://localhost/live", Valid: true},
	})
	if err != nil {
		h.logger.Error("Failed to create stream", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	if err := json.NewEncoder(w).Encode(toStreamResponse(stream)); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func generateStreamKey() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "live_" + hex.EncodeToString(b)
}

// RestreamDestination represents a single restream target
type RestreamDestination struct {
	PluginID    string `json:"plugin_id"`    // e.g., "publisher-twitch"
	AccessToken string `json:"access_token"` // OAuth token for the platform
	Enabled     bool   `json:"enabled"`      // Whether this destination is active
}

// GetStreamDestinations returns the configured restream destinations for a stream
func (h *StreamsHandler) GetStreamDestinations(w http.ResponseWriter, r *http.Request) {
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

	var destinations []RestreamDestination
	if len(stream.RestreamDestinations) > 0 {
		if err := json.Unmarshal(stream.RestreamDestinations, &destinations); err != nil {
			h.logger.Warn("Failed to parse restream destinations", "error", err)
			destinations = []RestreamDestination{}
		}
	}

	if err := json.NewEncoder(w).Encode(destinations); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// UpdateStreamDestinations updates the restream destinations for a stream
func (h *StreamsHandler) UpdateStreamDestinations(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var uid pgtype.UUID
	if err := uid.Scan(id); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	var destinations []RestreamDestination
	if err := json.NewDecoder(r.Body).Decode(&destinations); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	destBytes, err := json.Marshal(destinations)
	if err != nil {
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	if err := h.db.UpdateStreamRestreamDestinations(r.Context(), store.UpdateStreamRestreamDestinationsParams{
		ID:                   uid,
		RestreamDestinations: destBytes,
	}); err != nil {
		h.logger.Error("Failed to update stream destinations", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	h.logger.Info("Updated stream destinations", "stream_id", id, "count", len(destinations))
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "destinations updated",
		"count":   len(destinations),
	}); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}
