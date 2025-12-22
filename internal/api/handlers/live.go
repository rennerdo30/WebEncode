package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
)

type LiveHandler struct {
	logger *logger.Logger
	db     store.Querier
}

func NewLiveHandler(l *logger.Logger, db store.Querier) *LiveHandler {
	return &LiveHandler{
		logger: l,
		db:     db,
	}
}

func extractStreamKey(path string) string {
	// MediaMTX paths usually look like "live/streamkey"
	// But could also be just "streamkey" if no app name is used
	key := path
	if strings.HasPrefix(path, "live/") {
		key = strings.TrimPrefix(path, "live/")
	}
	// Further cleanup if needed
	return strings.TrimSpace(key)
}

// HandleAuth is called by MediaMTX when a streamer attempts to publish
func (h *LiveHandler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	var path string
	if r.Header.Get("Content-Type") == "application/json" {
		var req struct {
			Path string `json:"path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			path = req.Path
		}
	} else {
		path = r.FormValue("path")
	}

	key := extractStreamKey(path)
	h.logger.Info("Live auth request", "path", path, "key", key)

	if key == "" {
		h.logger.Warn("Empty stream key in auth request", "path", path)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	stream, err := h.db.GetStreamByKey(r.Context(), key)
	if err != nil {
		h.logger.Warn("Stream key not found in DB", "key", key, "path", path)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	h.logger.Info("Stream authenticated successfully", "id", stream.ID, "key", key)
	w.WriteHeader(http.StatusOK)
}

// HandleStart is called when stream actually begins
func (h *LiveHandler) HandleStart(w http.ResponseWriter, r *http.Request) {
	var path string
	if r.Header.Get("Content-Type") == "application/json" {
		var req struct {
			Path string `json:"path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			path = req.Path
		}
	} else {
		path = r.FormValue("path")
	}

	key := extractStreamKey(path)
	h.logger.Info("Stream ready hook called", "path", path, "key", key)

	if key == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	stream, err := h.db.GetStreamByKey(r.Context(), key)
	if err != nil {
		h.logger.Warn("Stream key not found in start hook", "key", key)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Update stream status in DB
	err = h.db.UpdateStreamLive(r.Context(), store.UpdateStreamLiveParams{
		ID:     stream.ID,
		IsLive: true,
	})
	if err != nil {
		h.logger.Error("Failed to update stream status to live", "error", err, "stream_id", stream.ID)
	} else {
		h.logger.Info("Stream marked as live in database", "stream_id", stream.ID)
	}

	w.WriteHeader(http.StatusOK)
}

// HandleStop is called when stream ends
func (h *LiveHandler) HandleStop(w http.ResponseWriter, r *http.Request) {
	var path string
	if r.Header.Get("Content-Type") == "application/json" {
		var req struct {
			Path string `json:"path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			path = req.Path
		}
	} else {
		path = r.FormValue("path")
	}

	key := extractStreamKey(path)
	h.logger.Info("Stream stop hook called", "path", path, "key", key)

	if key == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	stream, err := h.db.GetStreamByKey(r.Context(), key)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Update stream status in DB
	err = h.db.UpdateStreamLive(r.Context(), store.UpdateStreamLiveParams{
		ID:     stream.ID,
		IsLive: false,
	})
	if err != nil {
		h.logger.Error("Failed to update stream status to offline", "error", err, "stream_id", stream.ID)
	}

	w.WriteHeader(http.StatusOK)
}
