package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rennerdo30/webencode/internal/encoder"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/errors"
	"github.com/rennerdo30/webencode/pkg/logger"
)

type ProfilesHandler struct {
	db     store.Querier
	logger *logger.Logger
}

func NewProfilesHandler(db store.Querier, l *logger.Logger) *ProfilesHandler {
	return &ProfilesHandler{db: db, logger: l}
}

func (h *ProfilesHandler) Register(r chi.Router) {
	r.Get("/v1/profiles", h.ListProfiles)
	r.Get("/v1/profiles/{id}", h.GetProfile)
	r.Post("/v1/profiles", h.CreateProfile)
	r.Put("/v1/profiles/{id}", h.UpdateProfile)
	r.Delete("/v1/profiles/{id}", h.DeleteProfile)
}

// ProfileResponse combines DB profiles with preset profiles
type ProfileResponse struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	VideoCodec  string          `json:"video_codec"`
	AudioCodec  string          `json:"audio_codec,omitempty"`
	Width       int             `json:"width,omitempty"`
	Height      int             `json:"height,omitempty"`
	BitrateKbps int             `json:"bitrate_kbps,omitempty"`
	Preset      string          `json:"preset,omitempty"`
	Container   string          `json:"container,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
	IsSystem    bool            `json:"is_system"`
}

func (h *ProfilesHandler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	// 1. Get preset (system) profiles
	presets := encoder.GetAvailableProfiles()
	profiles := make([]ProfileResponse, 0, len(presets))
	seenIDs := make(map[string]bool)

	for _, name := range presets {
		if p, ok := encoder.GetProfile(name); ok {
			profiles = append(profiles, ProfileResponse{
				ID:          name,
				Name:        name,
				VideoCodec:  p.VideoCodec,
				AudioCodec:  p.AudioCodec,
				Width:       p.Width,
				Height:      p.Height,
				BitrateKbps: p.BitrateKbps,
				Preset:      p.Preset,
				Container:   p.Container,
				IsSystem:    true,
			})
			seenIDs[name] = true
		}
	}

	// 2. Get custom profiles from DB (skip duplicates)
	dbProfiles, err := h.db.ListEncodingProfiles(r.Context())
	if err != nil {
		h.logger.Warn("Failed to list DB profiles", "error", err)
		// Continue with presets only
	} else {
		for _, dbp := range dbProfiles {
			// Skip if already in presets
			if seenIDs[dbp.ID] {
				continue
			}
			profiles = append(profiles, ProfileResponse{
				ID:          dbp.ID,
				Name:        dbp.Name,
				Description: dbp.Description.String,
				VideoCodec:  dbp.VideoCodec,
				AudioCodec:  dbp.AudioCodec.String,
				Width:       int(dbp.Width.Int32),
				Height:      int(dbp.Height.Int32),
				BitrateKbps: int(dbp.BitrateKbps.Int32),
				Preset:      dbp.Preset.String,
				Container:   dbp.Container.String,
				Config:      dbp.ConfigJson,
				IsSystem:    dbp.IsSystem.Bool,
			})
		}
	}

	if err := json.NewEncoder(w).Encode(profiles); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *ProfilesHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Check preset profiles first
	if p, ok := encoder.GetProfile(id); ok {
		if err := json.NewEncoder(w).Encode(ProfileResponse{
			ID:          id,
			Name:        id,
			VideoCodec:  p.VideoCodec,
			AudioCodec:  p.AudioCodec,
			Width:       p.Width,
			Height:      p.Height,
			BitrateKbps: p.BitrateKbps,
			Preset:      p.Preset,
			Container:   p.Container,
			IsSystem:    true,
		}); err != nil {
			h.logger.Error("Failed to encode response", "error", err)
		}
		return
	}

	// Check DB
	dbp, err := h.db.GetEncodingProfile(r.Context(), id)
	if err != nil {
		errors.Response(w, r, errors.ErrNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(ProfileResponse{
		ID:          dbp.ID,
		Name:        dbp.Name,
		Description: dbp.Description.String,
		VideoCodec:  dbp.VideoCodec,
		AudioCodec:  dbp.AudioCodec.String,
		Width:       int(dbp.Width.Int32),
		Height:      int(dbp.Height.Int32),
		BitrateKbps: int(dbp.BitrateKbps.Int32),
		Preset:      dbp.Preset.String,
		Container:   dbp.Container.String,
		Config:      dbp.ConfigJson,
		IsSystem:    dbp.IsSystem.Bool,
	}); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

type ProfileRequest struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	VideoCodec  string          `json:"video_codec"`
	AudioCodec  string          `json:"audio_codec"`
	Width       int             `json:"width"`
	Height      int             `json:"height"`
	BitrateKbps int             `json:"bitrate_kbps"`
	Preset      string          `json:"preset"`
	Container   string          `json:"container"`
	Config      json.RawMessage `json:"config"`
}

func (h *ProfilesHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var req ProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	if req.ID == "" || req.Name == "" || req.VideoCodec == "" {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	profile, err := h.db.CreateEncodingProfile(r.Context(), store.CreateEncodingProfileParams{
		ID:          req.ID,
		Name:        req.Name,
		Description: store.ToText(req.Description),
		VideoCodec:  req.VideoCodec,
		AudioCodec:  store.ToText(req.AudioCodec),
		Width:       store.ToInt4(int32(req.Width)),
		Height:      store.ToInt4(int32(req.Height)),
		BitrateKbps: store.ToInt4(int32(req.BitrateKbps)),
		Preset:      store.ToText(req.Preset),
		Container:   store.ToText(req.Container),
		ConfigJson:  req.Config,
	})
	if err != nil {
		h.logger.Error("Failed to create profile", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(profile); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *ProfilesHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req ProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	err := h.db.UpdateEncodingProfile(r.Context(), store.UpdateEncodingProfileParams{
		ID:          id,
		Name:        req.Name,
		Description: store.ToText(req.Description),
		VideoCodec:  req.VideoCodec,
		AudioCodec:  store.ToText(req.AudioCodec),
		Width:       store.ToInt4(int32(req.Width)),
		Height:      store.ToInt4(int32(req.Height)),
		BitrateKbps: store.ToInt4(int32(req.BitrateKbps)),
		Preset:      store.ToText(req.Preset),
		Container:   store.ToText(req.Container),
		ConfigJson:  req.Config,
	})
	if err != nil {
		h.logger.Error("Failed to update profile", "id", id, "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ProfilesHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Cannot delete system profiles
	if _, ok := encoder.GetProfile(id); ok {
		errors.Response(w, r, &errors.WebEncodeError{
			Code:       "PROFILE_SYSTEM",
			Message:    "Cannot delete system profile",
			HTTPStatus: 403,
		})
		return
	}

	if err := h.db.DeleteEncodingProfile(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete profile", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
